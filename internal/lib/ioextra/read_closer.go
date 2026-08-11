package ioextra

import "io"

type ReadCloser struct {
	io.Reader
	io.Closer
}

func NewReadCloser(r io.Reader, c io.Closer) io.ReadCloser {
	return &ReadCloser{
		Reader: r,
		Closer: c,
	}
}
