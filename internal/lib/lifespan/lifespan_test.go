package lifespan_test

import (
	"context"
	"errors"
	"testing"

	"github.com/nikimonax/go-metrics/internal/lib/lifespan"
	"github.com/stretchr/testify/assert"
)

func TestLifespanOpen(t *testing.T) {
	t.Run("open order", func(t *testing.T) {
		var calls []string
		l := lifespan.New()
		l.OnStartup(func(context.Context) error { calls = append(calls, "first"); return nil })
		l.OnStartup(func(context.Context) error { calls = append(calls, "second"); return nil })

		assert.NoError(t, l.Open(context.Background()))
		assert.Equal(t, []string{"first", "second"}, calls)
	})

	t.Run("startup error", func(t *testing.T) {
		wantErr := errors.New("startup error")
		called := false

		l := lifespan.New()
		l.OnStartup(func(context.Context) error { return wantErr })
		l.OnStartup(func(context.Context) error { called = true; return nil })

		assert.ErrorIs(t, l.Open(context.Background()), wantErr)
		assert.False(t, called)
	})
}

func TestLifespanClose(t *testing.T) {
	t.Run("emtry close", func(t *testing.T) {
		l := lifespan.New()
		assert.NoError(t, l.Close(context.Background()))
	})

	t.Run("close order", func(t *testing.T) {
		var calls []string
		errFirst := errors.New("first error")
		errSecond := errors.New("second error")

		l := lifespan.New()
		l.OnShutdown(func(context.Context) error { calls = append(calls, "first"); return errFirst })
		l.OnShutdown(func(context.Context) error { calls = append(calls, "second"); return errSecond })

		err := l.Close(context.Background())

		assert.Equal(t, []string{"second", "first"}, calls)
		assert.ErrorIs(t, err, errFirst)
		assert.ErrorIs(t, err, errSecond)
	})
}
