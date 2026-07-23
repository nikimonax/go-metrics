package ioextra

type CloserFunc func() error

func (f CloserFunc) Close() error {
	return f()
}
