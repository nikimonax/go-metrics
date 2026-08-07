package scheduler_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nikimonax/go-metrics/internal/lib/scheduler"
)

func TestScheduler(t *testing.T) {
	t.Run("schedule validation", func(t *testing.T) {
		s := scheduler.New()
		callback := func(context.Context) error { return nil }

		assert.ErrorIs(t, s.Schedule("invalid interval", 0, callback), scheduler.ErrInvalidInterval)
		assert.ErrorIs(t, s.Schedule("invalid callback", time.Second, nil), scheduler.ErrInvalidCallback)
		assert.NoError(t, s.Schedule("valid", time.Second, callback))
	})

	t.Run("cannot schedule or start twice", func(t *testing.T) {
		var err error
		s := scheduler.New()

		err = s.Schedule("task", time.Second, func(context.Context) error { return nil })
		require.NoError(t, err)

		require.NoError(t, s.Start(context.Background()))
		assert.ErrorIs(t, s.Start(context.Background()), scheduler.ErrAlreadyRunning)

		err = s.Schedule("another", time.Second, func(context.Context) error { return nil })
		assert.ErrorIs(t, err, scheduler.ErrAlreadyRunning)

		require.NoError(t, s.Stop(context.Background()))
	})

	t.Run("on error callback", func(t *testing.T) {
		callbackErr := errors.New("callback error")
		called := make(chan struct{}, 1)
		errorsSeen := make(chan string, 1)

		s := scheduler.New()
		s.OnError = func(name string, err error) {
			assert.Equal(t, "task", name)
			assert.ErrorIs(t, err, callbackErr)
			errorsSeen <- name
		}

		var err error

		err = s.Schedule("task", time.Millisecond, func(context.Context) error {
			called <- struct{}{}
			return callbackErr
		})
		require.NoError(t, err)

		err = s.Start(context.Background())
		require.NoError(t, err)

		select {
		case <-called:
		case <-time.After(time.Second):
			t.Fatal("scheduled callback was not called")
		}
		select {
		case <-errorsSeen:
		case <-time.After(time.Second):
			t.Fatal("scheduler error was not reported")
		}

		require.NoError(t, s.Stop(context.Background()))
		assert.NoError(t, s.Stop(context.Background()))
	})

	t.Run("stop when context done", func(t *testing.T) {
		s := scheduler.New()
		s.ShutdownTimeout = time.Second

		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan error, 1)

		go func() { done <- s.Run(ctx) }()

		time.Sleep(time.Millisecond)
		cancel()

		select {
		case err := <-done:
			assert.NoError(t, err)
		case <-time.After(time.Second):
			t.Fatal("scheduler did not stop")
		}
	})

	t.Run("blocked stop", func(t *testing.T) {
		started := make(chan struct{})
		release := make(chan struct{})
		firstCall := true

		var err error

		s := scheduler.New()

		err = s.Schedule("blocked", time.Millisecond, func(context.Context) error {
			if firstCall {
				firstCall = false
				close(started)
				<-release
			}
			return nil
		})
		require.NoError(t, err)

		err = s.Start(context.Background())
		require.NoError(t, err)

		select {
		case <-started:
		case <-time.After(time.Second):
			t.Fatal("scheduled callback was not called")
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()

		assert.ErrorIs(t, s.Stop(ctx), context.DeadlineExceeded)

		close(release)
		assert.NoError(t, s.Stop(context.Background()))
	})

	t.Run("concurrent stop", func(t *testing.T) {
		s := scheduler.New()

		err := s.Start(context.Background())
		require.NoError(t, err)

		var wg sync.WaitGroup

		for range 2 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				assert.NoError(t, s.Stop(context.Background()))
			}()
		}

		wg.Wait()
	})
}
