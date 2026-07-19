package model

import "github.com/nikimonax/go-metrics/internal/domain"

type Metric struct {
	Name  domain.MetricName `json:"id"`              // имя метрики
	Type  domain.MetricType `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64            `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64          `json:"value,omitempty"` // значение метрики в случае передачи gauge
}
