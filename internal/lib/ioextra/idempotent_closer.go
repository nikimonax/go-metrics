package ioextra

import "io"

type IdempotentCloser struct {
	closer io.Closer
	closed bool
}

func (c *IdempotentCloser) Close() error {
	if c.closed {
		return nil
	}

	err := c.closer.Close()

	c.closed = true

	return err
}

var _ io.Closer = (*IdempotentCloser)(nil)

func NewIdempotentCloser(closer io.Closer) *IdempotentCloser {
	return &IdempotentCloser{
		closer: closer,
		closed: false,
	}
}
