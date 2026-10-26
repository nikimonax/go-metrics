package publisher

import (
	"context"
	"time"

	"github.com/nikimonax/go-metrics/internal/app/event"
	"github.com/nikimonax/go-metrics/internal/app/interfaces"
	"github.com/nikimonax/go-metrics/internal/impl/audit"
	"github.com/nikimonax/go-metrics/internal/lib/work"
)

type clientIPProvider = func(ctx context.Context) string

type AuditEventHandler struct {
	client    audit.Client
	submitter work.Submitter
	// client ip passed implicitly using context
	getClientIP clientIPProvider
}

// Handle implements [EventHandler].
func (handler *AuditEventHandler) Handle(
	ctx context.Context,
	e interfaces.Event,
) error {
	metricsUpdatedEvent, ok := e.(event.MetricsUpdatedEvent)

	if !ok {
		// skip event
		return nil
	}

	metricNames := make([]string, 0, len(metricsUpdatedEvent.Metrics))

	for _, metric := range metricsUpdatedEvent.Metrics {
		metricNames = append(metricNames, string(metric.Name()))
	}

	record := audit.Record{
		Timestamp: time.Now().Unix(),
		Metrics:   metricNames,
		IPAddress: handler.getClientIP(ctx),
	}

	task, err := work.NewTask(
		"audit record send",
		func(ctx context.Context) error {
			return handler.client.Send(ctx, record)
		},
	)

	if err != nil {
		return err
	}

	return handler.submitter.Submit(ctx, task)
}

var _ EventHandler = (*AuditEventHandler)(nil)

func NewAuditEventHandler(
	client audit.Client,
	submitter work.Submitter,
	getClientIP clientIPProvider,
) *AuditEventHandler {
	return &AuditEventHandler{
		client:      client,
		submitter:   submitter,
		getClientIP: getClientIP,
	}
}
