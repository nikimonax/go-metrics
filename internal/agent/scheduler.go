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

func runScheduler(tasks []Task) {
	now := time.Now()

	scheduled := make([]time.Time, 0, len(tasks))
	for _, task := range tasks {
		scheduled = append(scheduled, now.Add(task.Interval))
	}

	for {
		for i, task := range tasks {
			now := time.Now()

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

		if now := time.Now(); nextTaskTime.After(now) {
			time.Sleep(nextTaskTime.Sub(now))
		}
	}
}
