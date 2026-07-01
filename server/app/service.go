package app

import (
	"fmt"

	"github.com/nikimonax/go-metrics/pkg"
)

type MetricService struct {
	metricRepository MetricRepository
}

func (s *MetricService) Update(dto UpdateMetricDTO) error {
	switch dto.MetricType {
	case pkg.Counter:
		if err := s.metricRepository.CounterAdd(dto.MetricName, dto.ValueAdd); err != nil {
			return fmt.Errorf("counter add: %w", err)
		}
		return nil
	case pkg.Gauge:
		if err := s.metricRepository.GaugeSet(dto.MetricName, dto.ValueSet); err != nil {
			return fmt.Errorf("gauge set: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("unknown metric type: %s", dto.MetricType)
	}
}

func NewMetricService(metricRepository MetricRepository) *MetricService {
	return &MetricService{
		metricRepository: metricRepository,
	}
}
