package work

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testSubmitter struct {
	submitted chan Task
	err       error
}

func (s *testSubmitter) Submit(_ context.Context, task Task) error {
	if s.err != nil {
		return s.err
	}
	s.submitted <- task
	return nil
}

func TestScheduler(t *testing.T) {

	t.Run("rejects nil pool", func(t *testing.T) {
		_, err := NewScheduler(nil, SchedulerConfig{})
		assert.Error(t, err)
	})

	t.Run("validates constructor and schedules", func(t *testing.T) {
		pool := &testSubmitter{submitted: make(chan Task, 1)}

		scheduler, err := NewScheduler(pool, SchedulerConfig{})
		require.NoError(t, err)

		callback := func(context.Context) error { return nil }
		task, err := NewTask("task", callback)
		require.NoError(t, err)

		assert.Error(t, scheduler.Schedule(task, 0))
		require.NoError(t, scheduler.Schedule(task, time.Millisecond))

		require.NoError(t, scheduler.Start(context.Background()))
		assert.ErrorIs(t, scheduler.Schedule(task, time.Millisecond), ErrInvalidState)

		select {
		case got := <-pool.submitted:
			assert.Equal(t, task.Name, got.Name)
		case <-time.After(time.Second):
			require.Fail(t, "task was not submitted")
		}

		require.NoError(t, scheduler.Stop(context.Background()))
	})

	t.Run("reports submit errors", func(t *testing.T) {
		wantErr := errors.New("submit failed")

		taskQueue := make(chan Task)
		errorSeen := make(chan error, 1)

		pool := &testSubmitter{err: wantErr, submitted: taskQueue}

		onError := func(_ string, err error) { errorSeen <- err }
		config := SchedulerConfig{OnError: onError}

		scheduler, err := NewScheduler(pool, config)
		require.NoError(t, err)

		task, err := NewTask("task", func(context.Context) error { return nil })

		require.NoError(t, err)
		require.NoError(t, scheduler.Schedule(task, time.Millisecond))
		require.NoError(t, scheduler.Start(context.Background()))

		defer func() { assert.NoError(t, scheduler.Stop(context.Background())) }()

		select {
		case got := <-errorSeen:
			assert.ErrorIs(t, got, wantErr)
		case <-time.After(time.Second):
			require.Fail(t, "error callback was not called")
		}
	})

	t.Run("run stops on context cancellation", func(t *testing.T) {
		taskQueue := make(chan Task, 1)
		done := make(chan error, 1)

		pool := &testSubmitter{submitted: taskQueue}

		scheduler, err := NewScheduler(pool, SchedulerConfig{})
		require.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())

		go func() { done <- scheduler.Run(ctx) }()

		cancel()

		select {
		case err := <-done:
			assert.NoError(t, err)
		case <-time.After(time.Second):
			require.Fail(t, "scheduler did not stop")
		}
	})
}
