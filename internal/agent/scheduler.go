package agent

import (
	"slices"
	"time"

	"go.uber.org/zap"
)

type Task struct {
	Name     string
	Interval time.Duration
	Callback func() error
}

type Scheduler struct {
	timer func() time.Time
	sugar *zap.SugaredLogger
}

func (s *Scheduler) Run(tasks []Task) {
	now := s.timer()

	scheduled := make([]time.Time, 0, len(tasks))
	for _, task := range tasks {
		scheduled = append(scheduled, now.Add(task.Interval))
	}

	for {
		for i, task := range tasks {
			now := s.timer()

			if scheduled[i].After(now) {
				continue
			}

			if err := task.Callback(); err != nil && s.sugar != nil {
				s.sugar.Errorw(
					"task failed",
					"task", task.Name,
					"err", err,
				)
			}

			passedPeriods := now.Sub(scheduled[i])/task.Interval + 1

			if passedPeriods > 1 && s.sugar != nil {
				s.sugar.Warnw(
					"missed periods",
					"task", task.Name,
					"periods", passedPeriods-1,
				)
			}

			scheduled[i] = scheduled[i].Add(task.Interval * passedPeriods)
		}

		nextTaskTime := slices.MinFunc(
			scheduled,
			func(a, b time.Time) int {
				return a.Compare(b)
			},
		)

		if now := s.timer(); nextTaskTime.After(now) {
			time.Sleep(nextTaskTime.Sub(now))
		}
	}
}

func NewScheduler(
	timer func() time.Time,
	logger *zap.Logger,
) *Scheduler {
	var sugar *zap.SugaredLogger

	if logger != nil {
		sugar = logger.Sugar()
	}

	return &Scheduler{timer: timer, sugar: sugar}
}
