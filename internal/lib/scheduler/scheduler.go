package scheduler

import (
	"context"
	"sync"
	"time"
)

type Scheduler struct {
	tasks   []Task
	mu      sync.Mutex
	cancel  context.CancelFunc
	wg      *sync.WaitGroup
	running bool
	OnError func(name string, err error)
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

	s.Stop()
	return nil
}

func (s *Scheduler) Stop() {
	s.mu.Lock()

	if !s.running {
		s.mu.Unlock()
		return
	}

	cancel := s.cancel
	wg := s.wg

	s.cancel = nil
	s.wg = nil
	s.running = false

	s.mu.Unlock()

	cancel()
	wg.Wait()
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
		tasks:   make([]Task, 0),
		running: false,
	}
}
