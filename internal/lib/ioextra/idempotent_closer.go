package ioextra

import "io"

type IdempotentCloser struct {
	closer io.Closer
	closed bool
}

func (c *IdempotentCloser) Close() error {
	err := c.closer.Close()

	if err != nil {
		c.closed = true
	}

	return err
}

func NewIdempotentCloser(closer io.Closer) io.Closer {
	return &IdempotentCloser{
		closer: closer,
		closed: false,
	}
}
