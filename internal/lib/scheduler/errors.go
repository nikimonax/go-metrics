package scheduler

import "errors"

var (
	ErrAlreadyRunning  = errors.New("scheduler already running")
	ErrInvalidInterval = errors.New("scheduler interval must be greater than zero")
	ErrInvalidCallback = errors.New("scheduler callback is nil")
)
