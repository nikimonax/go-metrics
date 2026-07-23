package ioextra

import (
	"errors"
	"io"
)

func NewMultiCloser(closers ...io.Closer) io.Closer {
	return CloserFunc(func() error {
		errs := make([]error, 0)

		for _, closer := range closers {
			if err := closer.Close(); err != nil {
				errs = append(errs, err)
			}
		}

		if len(errs) == 0 {
			return nil
		}

		return errors.Join(errs...)
	})
}
