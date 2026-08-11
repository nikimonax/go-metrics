package scheduler

import (
	"context"
	"time"
)

type Task struct {
	Name     string
	Interval time.Duration
	Callback func(context.Context) error
}

func NewTask(
	name string,
	interval time.Duration,
	callback func(context.Context) error,
) (*Task, error) {
	if interval <= 0 {
		return nil, ErrInvalidInterval
	}

	if callback == nil {
		return nil, ErrInvalidCallback
	}

	return &Task{
		Name:     name,
		Interval: interval,
		Callback: callback,
	}, nil
}
