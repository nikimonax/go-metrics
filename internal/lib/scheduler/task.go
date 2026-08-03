package scheduler

import "time"

type Task struct {
	Name     string
	Interval time.Duration
	Callback func() error
}
