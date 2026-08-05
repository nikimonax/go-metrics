package scheduler

import (
	"context"
	"sync"
	"time"
)

const defaultShutdownTimeout = 5 * time.Second

type Scheduler struct {
	tasks           []Task
	mu              sync.Mutex
	cancel          context.CancelFunc
	wg              *sync.WaitGroup
	running         bool
	ShutdownTimeout time.Duration
	OnError         func(name string, err error)
}

func (s *Scheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return ErrAlreadyRunning
	}

	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	s.running = true

	s.wg = new(sync.WaitGroup)
	s.wg.Add(len(s.tasks))

	for i := range s.tasks {
		go s.runWorker(ctx, s.wg, &s.tasks[i])
	}

	return nil
}

func (s *Scheduler) Run(ctx context.Context) error {
	if err := s.Start(ctx); err != nil {
		return err
	}

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		s.ShutdownTimeout,
	)
	defer cancel()

	return s.Stop(shutdownCtx)
}

func (s *Scheduler) Stop(ctx context.Context) error {
	s.mu.Lock()

	if !s.running {
		s.mu.Unlock()
		return nil
	}

	cancel := s.cancel
	wg := s.wg

	s.mu.Unlock()

	cancel()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		s.mu.Lock()
		s.cancel = nil
		s.wg = nil
		s.running = false
		s.mu.Unlock()
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *Scheduler) Schedule(
	name string,
	interval time.Duration,
	callback func(context.Context) error,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return ErrAlreadyRunning
	}

	if interval <= 0 {
		return ErrInvalidInterval
	}

	if callback == nil {
		return ErrInvalidCallback
	}

	task := Task{
		Name:     name,
		Interval: interval,
		Callback: callback,
	}

	s.tasks = append(s.tasks, task)
	return nil
}

func (s *Scheduler) runWorker(ctx context.Context, wg *sync.WaitGroup, task *Task) {
	defer wg.Done()

	ticker := time.NewTicker(task.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := task.Callback(ctx)
			if err != nil && s.OnError != nil {
				s.OnError(task.Name, err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func New() *Scheduler {
	return &Scheduler{
		tasks:           make([]Task, 0),
		running:         false,
		ShutdownTimeout: defaultShutdownTimeout,
	}
}
