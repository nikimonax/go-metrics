package dumper

import (
	"errors"
	"os"

	"github.com/nikimonax/go-metrics/internal/app"
	"github.com/nikimonax/go-metrics/internal/domain"
	"github.com/nikimonax/go-metrics/internal/impl/serializer"
)

type FileMetricDumper struct {
	filePath   string
	serializer serializer.MetricSerializer
}

// Load implements [app.MetricDumper].
func (dumper *FileMetricDumper) Load() ([]domain.Metric, error) {
	content, err := os.ReadFile(dumper.filePath)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return make([]domain.Metric, 0), nil
		}

		return nil, err
	}

	return dumper.serializer.Decode(content)
}

// Save implements [app.MetricDumper].
func (dumper *FileMetricDumper) Save(metrics []domain.Metric) error {
	content, err := dumper.serializer.Encode(metrics)

	if err != nil {
		return err
	}

	return os.WriteFile(dumper.filePath, content, 0644)
}

var _ app.MetricDumper = (*FileMetricDumper)(nil)

func NewFileMetricDumper(
	filePath string,
	serializer serializer.MetricSerializer,
) *FileMetricDumper {
	return &FileMetricDumper{
		filePath:   filePath,
		serializer: serializer,
	}
}
