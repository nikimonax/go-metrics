package server

import "errors"

type Lifespan struct {
	onStartup  []func() error
	onShutdown []func() error
}

func (l *Lifespan) OnStartup(action func() error) {
	l.onStartup = append(l.onStartup, action)
}

func (l *Lifespan) OnShutdown(action func() error) {
	l.onShutdown = append(l.onShutdown, action)
}

func (l *Lifespan) Open() error {
	for _, f := range l.onStartup {
		if err := f(); err != nil {
			return err
		}
	}
	return nil
}

func (l *Lifespan) Close() (err error) {
	for _, f := range l.onShutdown {
		err = errors.Join(err, f())
	}
	return
}

func NewLifespan() *Lifespan {
	return &Lifespan{
		onStartup:  make([]func() error, 0),
		onShutdown: make([]func() error, 0),
	}
}
