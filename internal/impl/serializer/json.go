package serializer

import (
	"encoding/json"

	"github.com/nikimonax/go-metrics/internal/domain"
	"github.com/nikimonax/go-metrics/internal/model"
)

type MetricSerializer interface {
	Encode([]domain.Metric) ([]byte, error)
	Decode([]byte) ([]domain.Metric, error)
}

type JsonMetricSerializer struct {
}

func (s *JsonMetricSerializer) Encode(metrics []domain.Metric) ([]byte, error) {
	models := make([]*model.Metric, 0, len(metrics))

	for _, metric := range metrics {
		models = append(models, model.NewMetricFromDomain(metric))
	}

	return json.Marshal(models)
}

func (s *JsonMetricSerializer) Decode(data []byte) ([]domain.Metric, error) {
	if len(data) == 0 {
		return make([]domain.Metric, 0), nil
	}

	var models []model.Metric

	if err := json.Unmarshal(data, &models); err != nil {
		return nil, err
	}

	metrics := make([]domain.Metric, 0, len(models))

	for _, model := range models {
		metrics = append(metrics, model.ToDomain())
	}

	return metrics, nil
}

func NewJsonMetricSerializer() MetricSerializer {
	return new(JsonMetricSerializer)
}
