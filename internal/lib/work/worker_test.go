package work

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorker(t *testing.T) {
	t.Run("handles error", func(t *testing.T) {
		wantErr := errors.New("task failed")
		errorSeen := make(chan error, 1)

		callback := func(context.Context) error {
			return wantErr
		}

		config := WorkerConfig{
			OnError: func(_ string, err error) {
				errorSeen <- err
			},
		}

		worker := NewWorker(config)
		task, err := NewTask("task", callback)

		require.NoError(t, err)

		worker.Handle(context.Background(), task)

		select {
		case got := <-errorSeen:
			assert.ErrorIs(t, got, wantErr)
		case <-time.After(time.Second):
			require.Fail(t, "error callback was not called")
		}
	})

	t.Run("recovers panic", func(t *testing.T) {
		panicSeen := make(chan any, 1)
		panicValue := "boom"

		callback := func(context.Context) error {
			panic(panicValue)
		}

		config := WorkerConfig{
			OnPanic: func(_ string, value any) {
				panicSeen <- value
			},
		}

		worker := NewWorker(config)
		task, err := NewTask("task", callback)

		require.NoError(t, err)

		worker.Handle(context.Background(), task)

		select {
		case value := <-panicSeen:
			assert.Equal(t, panicValue, value)
		case <-time.After(time.Second):
			require.Fail(t, "panic callback was not called")
		}
	})

	t.Run("run stops on context cancellation", func(t *testing.T) {
		queue := make(chan Task)
		done := make(chan struct{})

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			NewWorker(WorkerConfig{}).Run(ctx, queue)
			close(done)
		}()

		cancel()

		select {
		case <-done:
		case <-time.After(time.Second):
			require.Fail(t, "worker did not stop")
		}
	})
}
