package ioextra

import (
	"errors"
	"io"
	"sync"
)

type MultiCloser struct {
	mu      sync.Mutex
	closers []io.Closer
}

func (m *MultiCloser) Append(closer io.Closer) {
	if closer == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closers = append(m.closers, closer)
}

func (m *MultiCloser) Close() (err error) {
	for _, closer := range m.closers {
		err = errors.Join(err, closer.Close())
	}
	return err
}

var _ io.Closer = (*MultiCloser)(nil)

func NewMultiCloser(closers ...io.Closer) *MultiCloser {
	return &MultiCloser{closers: closers}
}
