// Пакет предоставляет общие структуры для передачи
// и обновления метрик между агентом и сервером.

package model

import "github.com/nikimonax/go-metrics/internal/domain"

// Информация о метрике.
// Передаётся в запросах на получение метрики.
type MetricInfo struct {
	// Имя метрики.
	Name domain.MetricName `json:"id" validate:"required"`

	// Параметр, принимающий значение gauge или counter.
	Type domain.MetricType `json:"type" validate:"oneof=gauge counter"`
}

// Метрика.
// Передаётся в запросах на обновление метрики.
type Metric struct {
	MetricInfo

	// Значение метрики в случае передачи counter.
	Delta *int64 `json:"delta,omitempty" validate:"required_if=Type counter,excluded_if=Type gauge,omitnil,gte=0"`

	// Значение метрики в случае передачи gauge.
	Value *float64 `json:"value,omitempty" validate:"required_if=Type gauge,excluded_if=Type counter,omitnil,gte=0"`
}

// Перевод метрики в доменный вид.
func (payload *Metric) ToDomain() domain.Metric {
	switch payload.Type {
	case domain.Counter:
		return domain.NewCounterMetric(payload.Name, *payload.Delta)
	case domain.Gauge:
		return domain.NewGaugeMetric(payload.Name, *payload.Value)
	default:
		// unreachable
		return nil
	}
}

// Создание метрики из доменного вида.
func NewMetricFromDomain(metric domain.Metric) *Metric {
	payload := &Metric{
		MetricInfo: MetricInfo{
			Name: metric.Name(),
			Type: metric.Type(),
		},
	}

	switch m := metric.(type) {
	case *domain.CounterMetric:
		delta := m.InternalValue()
		payload.Delta = &delta
	case *domain.GaugeMetric:
		value := m.InternalValue()
		payload.Value = &value
	}

	return payload
}
