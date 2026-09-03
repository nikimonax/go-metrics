package work

import (
	"errors"
	"time"
)

type Schedule struct {
	Task

	Interval time.Duration
}

func NewSchedule(task Task, interval time.Duration) (Schedule, error) {
	if interval <= 0 {
		return Schedule{}, errors.New("required positive interval for schedule")
	}

	return Schedule{Task: task, Interval: interval}, nil
}
