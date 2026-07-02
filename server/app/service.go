package app

import (
	"github.com/nikimonax/go-metrics/pkg"
)

type MetricService struct {
	metricRepository MetricRepository
}

func (s *MetricService) Update(metric pkg.Metric) error {
	return s.metricRepository.Update(metric)
}

func NewMetricService(metricRepository MetricRepository) *MetricService {
	return &MetricService{
		metricRepository: metricRepository,
	}
}
