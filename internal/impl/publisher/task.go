package publisher

import (
	"context"

	"github.com/nikimonax/go-metrics/internal/app/interfaces"
	"github.com/nikimonax/go-metrics/internal/lib/work"
)

type SubmitTaskOnEventHandler struct {
	task      work.Task
	submitter work.Submitter
}

// Handle implements [interfaces.EventHandler].
func (handler *SubmitTaskOnEventHandler) Handle(
	ctx context.Context,
	_ interfaces.Event,
) error {
	return handler.submitter.Submit(ctx, handler.task)
}

var _ EventHandler = (*SubmitTaskOnEventHandler)(nil)

func NewSubmitTaskOnEventHandler(
	task work.Task,
	submitter work.Submitter,
) *SubmitTaskOnEventHandler {
	return &SubmitTaskOnEventHandler{
		task:      task,
		submitter: submitter,
	}
}
