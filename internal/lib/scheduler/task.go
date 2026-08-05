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
