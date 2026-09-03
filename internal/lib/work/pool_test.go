package work

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkerPool(t *testing.T) {
	t.Run("validates configuration", func(t *testing.T) {
		_, err := NewPool(PoolConfig{})
		assert.ErrorIs(t, err, ErrInvalidConfig)
	})

	t.Run("submits and executes task", func(t *testing.T) {
		pool, err := NewPool(PoolConfig{WorkerCount: 1, QueueSize: 1})
		require.NoError(t, err)

		done := make(chan struct{})
		task, err := NewTask(
			"task",
			func(_ context.Context) error {
				close(done)
				return nil
			},
		)

		require.NoError(t, err)
		assert.ErrorIs(t, pool.Submit(context.Background(), task), ErrInvalidState)
		require.NoError(t, pool.Start(context.Background()))
		require.NoError(t, pool.Submit(context.Background(), task))

		select {
		case <-done:
		case <-time.After(time.Second):
			require.Fail(t, "task was not executed")
		}

		require.NoError(t, pool.Stop(context.Background()))
	})

	t.Run("rejects submit after stop", func(t *testing.T) {
		pool, err := NewPool(PoolConfig{WorkerCount: 1})

		require.NoError(t, err)
		require.NoError(t, pool.Start(context.Background()))
		require.NoError(t, pool.Stop(context.Background()))
		assert.ErrorIs(t, pool.Submit(context.Background(), Task{}), ErrInvalidState)
	})

	t.Run("submit observes caller cancellation", func(t *testing.T) {
		pool, err := NewPool(PoolConfig{WorkerCount: 1, QueueSize: 1})
		require.NoError(t, err)

		release := make(chan struct{})

		task1, err := NewTask(
			"task1",
			func(context.Context) error {
				<-release
				return nil
			},
		)
		require.NoError(t, err)

		task2, err := NewTask(
			"task2",
			func(context.Context) error { return nil },
		)
		require.NoError(t, err)

		require.NoError(t, pool.Start(context.Background()))
		defer func() { assert.NoError(t, pool.Stop(context.Background())) }()

		require.NoError(t, pool.Submit(context.Background(), task1))
		require.NoError(t, pool.Submit(context.Background(), task2))

		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()

		err = pool.Submit(ctx, Task{})
		assert.ErrorIs(t, err, context.DeadlineExceeded)

		close(release)
	})
}
