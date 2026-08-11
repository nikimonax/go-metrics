package ioextra_test

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nikimonax/go-metrics/internal/lib/ioextra"
)

type closer struct {
	closed int
	err    error
}

func (c *closer) Close() error {
	c.closed++
	return c.err
}

func TestCloserFunc(t *testing.T) {
	wantErr := errors.New("close error")
	called := false

	err := ioextra.CloserFunc(func() error {
		called = true
		return wantErr
	}).Close()

	assert.True(t, called)
	assert.ErrorIs(t, err, wantErr)
}

func TestMultiCloser(t *testing.T) {
	errA := errors.New("error A")
	errB := errors.New("error B")

	t.Run("success", func(t *testing.T) {
		closerA := &closer{}
		closerB := &closer{}

		assert.NoError(t, ioextra.NewMultiCloser(closerA, closerB).Close())
	})

	t.Run("error", func(t *testing.T) {
		closerA := &closer{err: errA}
		closerB := &closer{}
		closerC := &closer{err: errB}

		err := ioextra.NewMultiCloser(closerA, closerB, closerC).Close()

		assert.Equal(t, 1, closerA.closed)
		assert.Equal(t, 1, closerB.closed)
		assert.Equal(t, 1, closerC.closed)
		assert.ErrorIs(t, err, errA)
		assert.ErrorIs(t, err, errB)
	})

}

func TestIdempotentCloser(t *testing.T) {
	underlying := &closer{err: errors.New("close error")}
	wrapped := ioextra.NewIdempotentCloser(underlying)

	err := wrapped.Close()
	assert.Error(t, err)
	assert.NoError(t, wrapped.Close())
	assert.Equal(t, 1, underlying.closed)
}

func TestReadCloser(t *testing.T) {
	underlying := &closer{}
	wrapped := ioextra.NewReadCloser(strings.NewReader("payload"), underlying)

	data, err := io.ReadAll(wrapped)
	require.NoError(t, err)
	assert.Equal(t, "payload", string(data))
	assert.NoError(t, wrapped.Close())
	assert.Equal(t, 1, underlying.closed)
}
