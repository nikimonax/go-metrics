package usecase

import (
	"context"

	"github.com/nikimonax/go-metrics/internal/app/interfaces"
)

// Сценарий отправки накопленных метрик на сервер.
type SendMetricsUseCase struct {
	gateway    interfaces.MetricGateway
	repository interfaces.MetricRepository
}

// Execute извлекает все накопленные метрики из хранилища и
// отправляет их другой сервер. Если метрик нет, ничего не делает.
func (useCase *SendMetricsUseCase) Execute(
	ctx context.Context,
) error {
	metrics, err := useCase.repository.PopAll(ctx)

	if err != nil {
		return err
	}

	if len(metrics) == 0 {
		return nil
	}

	return useCase.gateway.SendBatch(ctx, metrics)
}

// NewSendMetricsUseCase создаёт новый экземпляр SendMetricsUseCase
// с указанными клиентом отправки и хранилищем метрик.
func NewSendMetricsUseCase(
	gateway interfaces.MetricGateway,
	repository interfaces.MetricRepository,
) *SendMetricsUseCase {
	return &SendMetricsUseCase{
		gateway:    gateway,
		repository: repository,
	}
}
