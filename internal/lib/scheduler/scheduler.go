package scheduler

import (
	"context"
	"sync"
	"time"
)

type Scheduler struct {
	tasks   []Task
	cancel  context.CancelFunc
	wg      *sync.WaitGroup
	running bool
	OnError func(name string, err error)
}

func (s *Scheduler) Start() error {
	if s.running {
		return ErrAlreadyRunning
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	s.running = true
	go s.run(ctx)

	return nil
}

func (s *Scheduler) Schedule(
	name string,
	interval time.Duration,
	callback func() error,
) error {
	if s.running {
		return ErrAlreadyRunning
	}

	task := Task{Name: name, Interval: interval, Callback: callback}
	s.tasks = append(s.tasks, task)
	return nil
}

func (s *Scheduler) run(ctx context.Context) {
	s.wg = new(sync.WaitGroup)

	for _, task := range s.tasks {
		s.wg.Add(1)
		s.runWorker(ctx, s.wg, &task)
	}

}

func (s *Scheduler) runWorker(ctx context.Context, wg *sync.WaitGroup, task *Task) {
	defer wg.Done()

	ticker := time.NewTicker(task.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := task.Callback()
			if err != nil && s.OnError != nil {
				s.OnError(task.Name, err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *Scheduler) Stop() error {
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}

	if s.wg != nil {
		s.wg.Wait()
		s.wg = nil
	}

	s.running = false

	return nil
}

func New() *Scheduler {
	return &Scheduler{
		tasks:   make([]Task, 0),
		running: false,
	}
}
