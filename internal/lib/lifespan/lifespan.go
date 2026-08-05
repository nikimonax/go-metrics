package lifespan

import (
	"context"
	"errors"
)

type Lifespan struct {
	onStartup  []func(context.Context) error
	onShutdown []func(context.Context) error
}

func (l *Lifespan) OnStartup(action func(context.Context) error) {
	l.onStartup = append(l.onStartup, action)
}

func (l *Lifespan) OnShutdown(action func(context.Context) error) {
	l.onShutdown = append(l.onShutdown, action)
}

func (l *Lifespan) Open(ctx context.Context) error {
	for _, f := range l.onStartup {
		if err := f(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (l *Lifespan) Close(ctx context.Context) error {
	var err error

	for i := len(l.onShutdown) - 1; i >= 0; i-- {
		err = errors.Join(err, l.onShutdown[i](ctx))
	}

	return err
}

func New() *Lifespan {
	return &Lifespan{
		onStartup:  make([]func(context.Context) error, 0),
		onShutdown: make([]func(context.Context) error, 0),
	}
}
