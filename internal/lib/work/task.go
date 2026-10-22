package work

import (
	"context"
	"errors"
)

// Описывает единицу работы, выполняемую воркером.
type Task struct {
	// Имя задачи.
	// Используется в колбэках OnError и OnPanic.
	Name string

	// Функция, выполняемая воркером с контекстом пула.
	Callback func(context.Context) error

	// Пропускать задачу, если очередь пула заполнена.
	SkipIfFullChan bool
}

func NewTask(name string, callback func(context.Context) error) (Task, error) {
	if callback == nil {
		return Task{}, errors.New("required non-nil callback for task")
	}

	return Task{Name: name, Callback: callback}, nil
}
