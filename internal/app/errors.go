package app

import (
	"errors"
	"fmt"

	"github.com/nikimonax/go-metrics/internal/domain"
)

var ErrMetricNotFound = errors.New("metric not found")

func NewErrMetricNotFound(metricType domain.MetricType, metricName domain.MetricName) error {
	return fmt.Errorf("%w: type=%s, name=%s", ErrMetricNotFound, metricType, metricName)
}
