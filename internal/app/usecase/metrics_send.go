package usecase

import "github.com/nikimonax/go-metrics/internal/app/interfaces"

type SendMetricsUseCase struct {
	gateway    interfaces.MetricGateway
	repository interfaces.MetricRepository
}

func (useCase *SendMetricsUseCase) Execute() error {
	// TODO: заменить на атомарный GetAllAndClear
	metrics, err := useCase.repository.GetAll()

	if err != nil {
		return err
	}

	if len(metrics) == 0 {
		return nil
	}

	err = useCase.repository.Clear()

	if err != nil {
		return err
	}

	return useCase.gateway.SendBatch(metrics)
}

func NewSendMetricsUseCase(
	gateway interfaces.MetricGateway,
	repository interfaces.MetricRepository,
) *SendMetricsUseCase {
	return &SendMetricsUseCase{
		gateway:    gateway,
		repository: repository,
	}
}
