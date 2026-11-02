package httpextra

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/nikimonax/go-metrics/internal/lib/compress"
	"github.com/nikimonax/go-metrics/internal/lib/syncpool"
)

const maxPooledBufCap = 1 << 20

var encoders = map[string]encoderFunc{
	"gzip":    compress.EncoderGzip,
	"deflate": compress.EncoderDeflate,
}

type pooledBody struct {
	io.Reader
	buf *bytes.Buffer
	rt  *CompressRoundTripper
}

func newPooledBody(buf *bytes.Buffer, rt *CompressRoundTripper) io.ReadCloser {
	return &pooledBody{Reader: buf, buf: buf, rt: rt}
}

func (b *pooledBody) Close() error {
	b.rt.putBuffer(b.buf)
	return nil
}

type resettableWriter interface {
	io.Writer
	Reset(io.Writer)
}

type resetWriter struct {
	resettableWriter

	buffers       *syncpool.Pool[*bytes.Buffer]
	currentBuffer *bytes.Buffer
}

func (rw *resetWriter) Reset() {
	buf := rw.buffers.Get()
	rw.resettableWriter.Reset(buf)

	if rw.currentBuffer != nil {
		rw.buffers.Put(rw.currentBuffer)
	}

	rw.currentBuffer = buf
}

func (rw *resetWriter) DetachBuffer() *bytes.Buffer {
	buf := rw.currentBuffer
	rw.resettableWriter.Reset(io.Discard)
	rw.currentBuffer = nil

	return buf
}

func (rw *resetWriter) Close() error {
	if closer, ok := rw.resettableWriter.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

func newResetWriter(
	writer resettableWriter,
	buffers *syncpool.Pool[*bytes.Buffer],
) *resetWriter {
	rw := resetWriter{resettableWriter: writer, buffers: buffers}
	rw.Reset()
	return &rw
}

type CompressRoundTripper struct {
	next        http.RoundTripper
	encoderName string

	writers *syncpool.Pool[*resetWriter]
	buffers *syncpool.Pool[*bytes.Buffer]
}

// RoundTrip implements [http.RoundTripper].
func (rt *CompressRoundTripper) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	if req.Body == nil || req.ContentLength == 0 {
		return rt.next.RoundTrip(req)
	}

	if req.Header.Get(HDRContentEncoding) != "" {
		return rt.next.RoundTrip(req)
	}

	buf, err := rt.compress(req.Body)
	defer func() { err = errors.Join(err, req.Body.Close()) }()

	if err != nil {
		return nil, err
	}

	newReq := req.Clone(req.Context())
	newReq.Body = newPooledBody(buf, rt)
	newReq.ContentLength = int64(buf.Len())
	newReq.Header.Set(HDRContentEncoding, rt.encoderName)

	return rt.next.RoundTrip(newReq)
}

func (rt *CompressRoundTripper) compress(r io.Reader) (*bytes.Buffer, error) {
	w := rt.writers.Get()
	defer rt.writers.Put(w)

	if _, err := io.Copy(w, r); err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	return w.DetachBuffer(), nil
}

func (rt *CompressRoundTripper) putBuffer(buf *bytes.Buffer) {
	if buf.Cap() > maxPooledBufCap {
		return
	}

	rt.buffers.Put(buf)
}

var _ http.RoundTripper = (*CompressRoundTripper)(nil)

func NewCompressRoundTripper(
	next http.RoundTripper,
	encoder string,
	level int,
) (*CompressRoundTripper, error) {
	factory, ok := encoders[encoder]

	if !ok {
		return nil, fmt.Errorf("unknown encoder '%s'", encoder)
	}

	tmpWriter := factory(io.Discard, level)

	if tmpWriter == nil {
		return nil, fmt.Errorf("invalid compress level %d", level)
	}

	_, ok = tmpWriter.(resettableWriter)

	if !ok {
		return nil, fmt.Errorf("encoder '%s' does not support resets", encoder)
	}

	buffers := syncpool.New(func() *bytes.Buffer {
		return new(bytes.Buffer)
	})

	writers := syncpool.New(func() *resetWriter {
		rw := factory(io.Discard, level).(resettableWriter)
		return newResetWriter(rw, &buffers)
	})

	return &CompressRoundTripper{
		next:        next,
		encoderName: encoder,

		writers: &writers,
		buffers: &buffers,
	}, nil
}
