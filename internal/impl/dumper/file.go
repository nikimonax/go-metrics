package dumper

import (
	"errors"
	"os"
	"sync"

	"github.com/nikimonax/go-metrics/internal/app/interfaces"
	"github.com/nikimonax/go-metrics/internal/domain"
	"github.com/nikimonax/go-metrics/internal/impl/serializer"
)

type FileMetricDumper struct {
	mu          sync.Mutex
	filePath    string
	filePathTmp string
	serializer  serializer.MetricSerializer
}

// Load implements [interfaces.MetricDumper].
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

// Save implements [interfaces.MetricDumper].
func (dumper *FileMetricDumper) Save(metrics []domain.Metric) error {
	content, err := dumper.serializer.Encode(metrics)

	if err != nil {
		return err
	}

	dumper.mu.Lock()
	defer dumper.mu.Unlock()

	err = os.WriteFile(dumper.filePathTmp, content, 0644)

	if err != nil {
		return err
	}

	defer func() { _ = os.Remove(dumper.filePathTmp) }()

	return os.Rename(dumper.filePathTmp, dumper.filePath)
}

func NewFileMetricDumper(
	filePath string,
	serializer serializer.MetricSerializer,
) interfaces.MetricDumper {
	return &FileMetricDumper{
		filePath:    filePath,
		filePathTmp: filePath + ".tmp",
		serializer:  serializer,
	}
}
