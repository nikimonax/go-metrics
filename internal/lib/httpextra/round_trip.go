package httpextra

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type RoundTripperFunc func(*http.Request) (*http.Response, error)

func (f RoundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type encoderFunc func(io.Writer) io.WriteCloser

func encoderGzip(w io.Writer) io.WriteCloser {
	return gzip.NewWriter(w)
}

var encoders = map[string]encoderFunc{
	"gzip": encoderGzip,
}

type CompressRoundTripper struct {
	next        http.RoundTripper
	encoderName string
	encoderFunc encoderFunc
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
	newReq.Body = io.NopCloser(buf)
	newReq.ContentLength = int64(buf.Len())
	newReq.Header.Set(HDRContentEncoding, rt.encoderName)

	return rt.next.RoundTrip(newReq)
}

func (rt *CompressRoundTripper) compress(r io.Reader) (buf *bytes.Buffer, err error) {
	buf = new(bytes.Buffer)

	w := rt.encoderFunc(buf)
	defer func() { err = errors.Join(err, w.Close()) }()

	if _, err = io.Copy(w, r); err != nil {
		return nil, err
	}

	return buf, nil
}

func NewCompressRoundTripper(next http.RoundTripper, encoder string) http.RoundTripper {
	encoderFunc, ok := encoders[encoder]

	if !ok {
		msg := fmt.Sprintf("unknown encoder '%s'", encoder)
		panic(msg)
	}

	return &CompressRoundTripper{
		next:        next,
		encoderName: encoder,
		encoderFunc: encoderFunc,
	}
}
