package work

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLifecycle(t *testing.T) {
	t.Run("starts and stops", func(t *testing.T) {
		started := make(chan struct{})
		stopped := make(chan struct{})

		lc := lifecycle{
			onStart: func(context.Context) error {
				close(started)
				return nil
			},
			onStop: func() error {
				close(stopped)
				return nil
			},
		}

		require.NoError(t, lc.Start(context.Background()))

		select {
		case <-started:
		case <-time.After(time.Second):
			assert.Fail(t, "onStart callback not executed")
		}

		assert.ErrorIs(t, lc.Start(context.Background()), ErrInvalidState)

		require.NoError(t, lc.Stop(context.Background()))

		select {
		case <-stopped:
		case <-time.After(time.Second):
			assert.Fail(t, "onStop callback not executed")
		}

		require.NoError(t, lc.Stop(context.Background()))
	})

	t.Run("startup error resets state", func(t *testing.T) {
		wantErr := errors.New("start failed")

		lc := lifecycle{
			onStart: func(context.Context) error {
				return wantErr
			},
		}

		assert.ErrorIs(t, lc.Start(context.Background()), wantErr)
		assert.Equal(t, lifecycleStopped, lc.state)
	})

	t.Run("stop waits for onStop", func(t *testing.T) {
		release := make(chan struct{})

		lc := lifecycle{
			onStop: func() error {
				<-release
				return nil
			},
		}

		require.NoError(t, lc.Start(context.Background()))

		stopCtx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()

		assert.ErrorIs(t, lc.Stop(stopCtx), context.DeadlineExceeded)
		close(release)

		require.NoError(t, lc.Stop(context.Background()))
	})

	t.Run("run stops when context is cancelled", func(t *testing.T) {
		lc := lifecycle{}

		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan error, 1)

		go func() { done <- lc.Run(ctx) }()

		cancel()

		select {
		case err := <-done:
			assert.NoError(t, err)
		case <-time.After(time.Second):
			require.Fail(t, "lifecycle did not stop")
		}
	})
}
