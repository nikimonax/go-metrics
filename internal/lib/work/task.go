package work

import (
	"context"
	"errors"
)

type Task struct {
	Name     string
	Callback func(context.Context) error
}

func NewTask(name string, callback func(context.Context) error) (Task, error) {
	if callback == nil {
		return Task{}, errors.New("required non-nil callback for task")
	}

	return Task{Name: name, Callback: callback}, nil
}
