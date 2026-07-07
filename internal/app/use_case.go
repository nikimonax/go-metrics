package app

import (
	"github.com/nikimonax/go-metrics/internal/domain"
)

// update metric

type UpdateMetricUseCase struct {
	metricRepository MetricRepository
}

func (useCase *UpdateMetricUseCase) Execute(metric domain.Metric) error {
	return useCase.metricRepository.Update(metric)
}

func NewUpdateMetricUseCase(metricRepository MetricRepository) *UpdateMetricUseCase {
	return &UpdateMetricUseCase{
		metricRepository: metricRepository,
	}
}

// get metric

type GetMetricUseCase struct {
	metricRepository MetricRepository
}

func (useCase *GetMetricUseCase) Execute(
	metricType domain.MetricType,
	metricName domain.MetricName,
) (domain.Metric, error) {
	return useCase.metricRepository.Get(metricType, metricName)
}

func NewGetMetricUseCase(metricRepository MetricRepository) *GetMetricUseCase {
	return &GetMetricUseCase{
		metricRepository: metricRepository,
	}
}

// collect metrics

type CollectMetricsUseCase struct {
	collector  MetricCollector
	repository MetricRepository
}

func (useCase *CollectMetricsUseCase) Execute() error {
	metrics, err := useCase.collector.Collect()

	if err != nil {
		return err
	}

	if len(metrics) == 0 {
		return nil
	}

	return useCase.repository.UpdateBatch(metrics)
}

func NewCollectMetricsUseCase(
	collector MetricCollector,
	repository MetricRepository,
) *CollectMetricsUseCase {
	return &CollectMetricsUseCase{
		collector:  collector,
		repository: repository,
	}
}

// send metrics

type SendMetricsUseCase struct {
	gateway    MetricGateway
	repository MetricRepository
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
	gateway MetricGateway,
	repository MetricRepository,
) *SendMetricsUseCase {
	return &SendMetricsUseCase{
		gateway:    gateway,
		repository: repository,
	}
}
