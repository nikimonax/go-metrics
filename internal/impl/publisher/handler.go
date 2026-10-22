package publisher

import (
	"context"

	"github.com/nikimonax/go-metrics/internal/app/interfaces"
)

type EventHandler interface {
	Handle(context.Context, interfaces.Event) error
}

type EventHandlerFunc func(context.Context, interfaces.Event) error

func (f EventHandlerFunc) Handle(
	ctx context.Context,
	e interfaces.Event,
) error {
	return f(ctx, e)
}

var _ EventHandler = EventHandlerFunc(nil)
