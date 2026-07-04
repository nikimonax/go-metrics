package agent

import (
	"fmt"
	"slices"
	"time"
)

type Task struct {
	Name     string
	Interval time.Duration
	Callback func() error
}

type Scheduler struct {
	Timer func() time.Time
}

func (s *Scheduler) Run(tasks []Task) {
	now := s.Timer()

	scheduled := make([]time.Time, 0, len(tasks))
	for _, task := range tasks {
		scheduled = append(scheduled, now.Add(task.Interval))
	}

	for {
		for i, task := range tasks {
			now := s.Timer()

			if scheduled[i].After(now) {
				continue
			}

			if err := task.Callback(); err != nil {
				fmt.Printf("[ERROR] Task '%s' failed, reason: %s\n", task.Name, err)
			}

			passedPeriods := now.Sub(scheduled[i])/task.Interval + 1

			if passedPeriods > 1 {
				fmt.Printf("[WARN] Task '%s', missed %d periods", task.Name, passedPeriods-1)
			}

			scheduled[i] = scheduled[i].Add(task.Interval * passedPeriods)
		}

		nextTaskTime := slices.MinFunc(
			scheduled,
			func(a, b time.Time) int {
				return a.Compare(b)
			},
		)

		if now := s.Timer(); nextTaskTime.After(now) {
			time.Sleep(nextTaskTime.Sub(now))
		}
	}
}

func NewScheduler(timer func() time.Time) *Scheduler {
	return &Scheduler{Timer: timer}
}
