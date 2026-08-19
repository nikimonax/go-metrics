package work

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type lifecycleState uint8

const (
	lifecycleStopped lifecycleState = iota
	lifecycleStarting
	lifecycleRunning
	lifecycleStopping
)

type lifecycle struct {
	mu          sync.Mutex
	state       lifecycleState
	runCtx      context.Context
	cancel      context.CancelFunc
	onStart     func(context.Context) error
	onStop      func() error
	stopTimeout time.Duration
	stopDone    chan struct{}
	stopErr     error
}

func (l *lifecycle) Start(runCtx context.Context) error {
	l.mu.Lock()

	if l.state != lifecycleStopped {
		l.mu.Unlock()
		return newErrInvalidState("already running")
	}

	l.state = lifecycleStarting
	l.mu.Unlock()

	runCtx, cancel := context.WithCancel(runCtx)
	if l.onStart != nil {
		if err := l.onStart(runCtx); err != nil {
			cancel()
			l.mu.Lock()
			l.state = lifecycleStopped
			l.mu.Unlock()
			return fmt.Errorf("failed start: %w", err)
		}
	}

	l.mu.Lock()
	l.runCtx = runCtx
	l.cancel = cancel
	l.state = lifecycleRunning
	l.mu.Unlock()

	return nil
}

func (l *lifecycle) Stop(stopCtx context.Context) error {
	l.mu.Lock()

	switch l.state {
	case lifecycleStopped:
		l.mu.Unlock()
		return nil
	case lifecycleStarting:
		l.mu.Unlock()
		return newErrInvalidState("starting")
	case lifecycleStopping:
		done := l.stopDone
		l.mu.Unlock()
		return l.waitForStop(stopCtx, done)
	case lifecycleRunning:
		done := make(chan struct{})
		l.cancel()
		l.state = lifecycleStopping
		l.stopDone = done
		l.mu.Unlock()

		go l.performStop(done)
		return l.waitForStop(stopCtx, done)
	default:
		l.mu.Unlock()
		return nil
	}
}

func (l *lifecycle) performStop(done chan struct{}) {
	var err error
	if l.onStop != nil {
		err = l.onStop()
	}

	l.mu.Lock()
	l.stopErr = err
	l.runCtx = nil
	l.cancel = nil
	l.state = lifecycleStopped
	close(done)
	l.mu.Unlock()
}

func (l *lifecycle) waitForStop(stopCtx context.Context, done <-chan struct{}) error {
	select {
	case <-done:
		l.mu.Lock()
		err := l.stopErr
		l.mu.Unlock()
		return err
	case <-stopCtx.Done():
		return stopCtx.Err()
	}
}

func (l *lifecycle) Run(runCtx context.Context) error {
	if err := l.Start(runCtx); err != nil {
		return err
	}

	<-runCtx.Done()

	stopCtx := context.Background()
	if l.stopTimeout > 0 {
		var cancel context.CancelFunc
		stopCtx, cancel = context.WithTimeout(stopCtx, l.stopTimeout)
		defer cancel()
	}

	return l.Stop(stopCtx)
}
