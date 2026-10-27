package ioextra

import "io"

type ReadCloser struct {
	io.Reader
	io.Closer
}

var _ io.ReadCloser = (*ReadCloser)(nil)

func NewReadCloser(r io.Reader, c io.Closer) *ReadCloser {
	return &ReadCloser{
		Reader: r,
		Closer: c,
	}
}
