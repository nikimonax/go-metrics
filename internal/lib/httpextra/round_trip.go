package httpextra

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"golang.org/x/sync/semaphore"

	"github.com/nikimonax/go-metrics/internal/lib/compress"
	"github.com/nikimonax/go-metrics/internal/lib/ioextra"
	"github.com/nikimonax/go-metrics/internal/lib/syncpool"
)

const maxPooledBufCap = 1 << 20

type RoundTripperFunc func(*http.Request) (*http.Response, error)

func (f RoundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type encoderFunc func(w io.Writer, level int) io.Writer

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

type CompressRoundTripper struct {
	next        http.RoundTripper
	encoderName string
	encoderFunc encoderFunc
	level       int

	writers syncpool.Pool[io.WriteCloser]
	buffers syncpool.Pool[*bytes.Buffer]
}

func (rt *CompressRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Body == nil || req.ContentLength == 0 {
		return rt.next.RoundTrip(req)
	}

	buf, err := rt.compress(req.Body)

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
	buf := rt.buffers.Get()
	buf.Reset()

	w := rt.writers.Get()

	if err := compress.ResetEncoder(w, buf); err != nil {
		return nil, err
	}

	if _, err := io.Copy(w, r); err != nil {
		rt.putWriter(w)
		rt.putBuffer(buf)
		return nil, err
	}

	if err := w.Close(); err != nil {
		rt.putWriter(w)
		rt.putBuffer(buf)
		return nil, err
	}

	rt.putWriter(w)

	return buf, nil
}

func (rt *CompressRoundTripper) putWriter(w io.WriteCloser) {
	rt.writers.Put(w)
}

func (rt *CompressRoundTripper) putBuffer(buf *bytes.Buffer) {
	if buf.Cap() > maxPooledBufCap {
		return
	}

	rt.buffers.Put(buf)
}

func NewCompressRoundTripper(next http.RoundTripper, encoder string, level int) http.RoundTripper {
	encoderFunc, ok := encoders[encoder]

	if !ok {
		msg := fmt.Sprintf("unknown encoder '%s'", encoder)
		panic(msg)
	}

	return &CompressRoundTripper{
		next:        next,
		encoderName: encoder,
		encoderFunc: encoderFunc,
		level:       level,

		writers: syncpool.New(func() io.WriteCloser {
			w := encoderFunc(io.Discard, level)

			if w == nil {
				panic(fmt.Sprintf("invalid compress level %d", level))
			}

			wc, ok := w.(io.WriteCloser)

			if ok {
				return wc
			}

			return ioextra.NopWriteCloser(w)
		}),

		buffers: syncpool.New(func() *bytes.Buffer {
			return new(bytes.Buffer)
		}),
	}
}

func NewRoundTripperFromHandler(handler http.Handler) http.RoundTripper {
	fn := func(r *http.Request) (*http.Response, error) {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		return w.Result(), nil
	}
	return RoundTripperFunc(fn)
}

type RateLimitRoundTripper struct {
	next http.RoundTripper
	sem  *semaphore.Weighted
}

// RoundTrip implements [http.RoundTripper].
func (rt *RateLimitRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if err := rt.sem.Acquire(req.Context(), 1); err != nil {
		return nil, err
	}
	defer rt.sem.Release(1)

	return rt.next.RoundTrip(req)
}

func NewRateLimitRoundTripper(
	next http.RoundTripper,
	limit int64,
) http.RoundTripper {
	return &RateLimitRoundTripper{
		next: next,
		sem:  semaphore.NewWeighted(limit),
	}
}
