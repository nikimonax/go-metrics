package work

import (
	"errors"
	"time"
)

// Связывает задачу с интервалом её периодического запуска.
type Schedule struct {
	Task

	// Период запуска задачи; должен быть положительным.
	Interval time.Duration
}

func NewSchedule(task Task, interval time.Duration) (Schedule, error) {
	if interval <= 0 {
		return Schedule{}, errors.New("required positive interval for schedule")
	}

	return Schedule{Task: task, Interval: interval}, nil
}
