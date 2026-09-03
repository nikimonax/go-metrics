package work

import (
	"context"
	"errors"
	"sync"
	"time"
)

type Submitter interface {
	Submit(context.Context, Task) error
}

type Scheduler struct {
	lifecycle

	pool      Submitter
	config    SchedulerConfig
	schedules []Schedule
	wg        sync.WaitGroup
}

func (s *Scheduler) onStart(ctx context.Context) error {
	for _, schedule := range s.schedules {
		s.wg.Add(1)
		go func(schedule Schedule) {
			defer s.wg.Done()
			ticker := time.NewTicker(schedule.Interval)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					err := s.pool.Submit(ctx, schedule.Task)

					if err != nil && ctx.Err() == nil && s.config.OnError != nil {
						s.config.OnError(schedule.Name, err)
					}
				}
			}
		}(schedule)
	}

	return nil
}

func (s *Scheduler) onStop() error {
	s.wg.Wait()
	return nil
}

func (s *Scheduler) Schedule(task Task, interval time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state != lifecycleStopped {
		return newErrInvalidState("already running")
	}

	schedule, err := NewSchedule(task, interval)

	if err != nil {
		return err
	}

	s.schedules = append(s.schedules, schedule)
	return nil
}

func NewScheduler(pool Submitter, config SchedulerConfig) (*Scheduler, error) {
	if pool == nil {
		return nil, errors.New("required non-nil pool")
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	s := &Scheduler{pool: pool, config: config}
	s.lifecycle.onStart = s.onStart
	s.lifecycle.onStop = s.onStop
	s.lifecycle.stopTimeout = s.config.StopTimeout //nolint:staticcheck

	return s, nil
}
