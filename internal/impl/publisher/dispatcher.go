package publisher

import (
	"context"
	"sync"

	"github.com/nikimonax/go-metrics/internal/app/interfaces"
)

type EventDispatcher struct {
	mu       sync.RWMutex
	handlers map[string][]EventHandler
	OnError  func(interfaces.Event, error)
}

func (dp *EventDispatcher) Register(eventName string, handler EventHandler) {
	dp.mu.Lock()
	defer dp.mu.Unlock()

	if handler == nil {
		return
	}

	handlers, ok := dp.handlers[eventName]

	if !ok {
		handlers = make([]EventHandler, 0, 1)
	}

	handlers = append(handlers, handler)
	dp.handlers[eventName] = handlers
}

func (dp *EventDispatcher) Publish(
	ctx context.Context,
	event interfaces.Event,
) {
	dp.mu.RLock()

	handlers, ok := dp.handlers[event.Name()]

	if !ok {
		dp.mu.RUnlock()
		return
	}

	handlersCopy := make([]EventHandler, len(handlers))
	copy(handlersCopy, handlers)

	dp.mu.RUnlock()

	for _, handler := range handlersCopy {
		err := handler.Handle(ctx, event)
		if err != nil && dp.OnError != nil {
			dp.OnError(event, err)
		}
	}
}

var _ interfaces.EventPublisher = (*EventDispatcher)(nil)

func NewEventDispatcher() *EventDispatcher {
	return &EventDispatcher{handlers: make(map[string][]EventHandler)}
}
