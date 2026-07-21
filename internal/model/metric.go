package model

import "github.com/nikimonax/go-metrics/internal/domain"

type UpdateMetric struct {
	// Имя метрики
	Name domain.MetricName `json:"id" validate:"required"`

	// Параметр, принимающий значение gauge или counter
	Type domain.MetricType `json:"type" validate:"oneof=gauge counter"`

	// Значение метрики в случае передачи counter
	Delta *int64 `json:"delta,omitempty" validate:"required_if=Type counter,excluded_if=Type gauge,omitnil,gte=0"`

	// Значение метрики в случае передачи gauge
	Value *float64 `json:"value,omitempty" validate:"required_if=Type gauge,excluded_if=Type counter,omitnil,gte=0"`
}

func (m *UpdateMetric) ToDomain() domain.Metric {
	switch m.Type {
	case "counter":
		return domain.NewCounterMetric(m.Name, *m.Delta)
	case "gauge":
		return domain.NewGaugeMetric(m.Name, *m.Value)
	default:
		// unreachable
		return nil
	}
}
