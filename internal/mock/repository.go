package mock

import (
	"github.com/nikimonax/go-metrics/internal/app"
	"github.com/nikimonax/go-metrics/internal/domain"
	"github.com/stretchr/testify/mock"
)

type MetricRepository struct {
	mock.Mock
}

// Update implements [MetricRepository].
func (repo *MetricRepository) Update(metric domain.Metric) error {
	return repo.Called(metric).Error(0)
}

// UpdateBatch implements [MetricRepository].
func (repo *MetricRepository) UpdateBatch(metrics []domain.Metric) error {
	return repo.Called(metrics).Error(0)
}

// Get implements [MetricRepository].
func (repo *MetricRepository) Get(
	metricType domain.MetricType,
	metricName domain.MetricName,
) (domain.Metric, error) {
	args := repo.Called(metricType, metricName)
	return args.Get(0).(domain.Metric), args.Error(1)
}

// GetAll implements [MetricRepository].
func (repo *MetricRepository) GetAll() ([]domain.Metric, error) {
	args := repo.Called()
	return args.Get(0).([]domain.Metric), args.Error(1)
}

// Clear implements [MetricRepository].
func (repo *MetricRepository) Clear() error {
	return repo.Called().Error(0)
}

var _ app.MetricRepository = (*MetricRepository)(nil)
