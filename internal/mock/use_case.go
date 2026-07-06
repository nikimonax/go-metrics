package mock

import (
	"github.com/nikimonax/go-metrics/internal/domain"
	"github.com/nikimonax/go-metrics/internal/server"
	"github.com/stretchr/testify/mock"
)

type UpdateMetricUseCase struct {
	mock.Mock
}

// Execute implements [server.UpdateMetricUseCase].
func (u *UpdateMetricUseCase) Execute(metric domain.Metric) error {
	return u.Called(metric).Error(0)
}

var _ server.UpdateMetricUseCase = (*UpdateMetricUseCase)(nil)
