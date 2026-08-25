package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nikimonax/go-metrics/internal/app/interfaces"
	"github.com/nikimonax/go-metrics/internal/app/usecase"
	"github.com/nikimonax/go-metrics/internal/domain"
	"github.com/nikimonax/go-metrics/internal/testing/mock"
	"github.com/nikimonax/go-metrics/internal/testing/shared"
)

var metric domain.Metric = new(mock.Metric)
var metricsEmpty = make([]domain.Metric, 0)
var metricsNonEmpty = []domain.Metric{metric}

func TestUpdateMetricUseCase(t *testing.T) {
	type TestCase struct {
		name  string
		setup func(*TestCase, *mock.MetricRepository, *mock.PublisherMock)
		err   error
	}

	tests := []TestCase{
		{
			name: "success",
			setup: func(
				_ *TestCase,
				repo *mock.MetricRepository,
				pub *mock.PublisherMock,
			) {
				repo.On("Update", metric).Return(nil).Once()
				pub.On(
					"Publish",
					mock.MatchedBy(mock.IsImplements(new(context.Context))),
					mock.MatchedBy(mock.IsImplements(new(interfaces.Event))),
				).Once()
			},
			err: nil,
		},
		{
			name: "repo error",
			setup: func(
				tc *TestCase,
				repo *mock.MetricRepository,
				_ *mock.PublisherMock,
			) {
				repo.On("Update", metric).Return(tc.err).Once()
			},
			err: errors.New("test"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repository := new(mock.MetricRepository)
			publisher := new(mock.PublisherMock)

			tc.setup(&tc, repository, publisher)

			useCase := usecase.NewUpdateMetricUseCase(repository, publisher)
			err := useCase.Execute(metric)

			assert.ErrorIs(t, err, tc.err)
			repository.AssertExpectations(t)
		})
	}
}

func TestGetMetricUseCase(t *testing.T) {
	type TestCase struct {
		name    string
		metric  domain.Metric
		setup   func(*TestCase, *mock.MetricRepository)
		wantErr bool
	}

	tests := []TestCase{
		{
			name:   "success",
			metric: domain.NewCounterMetric(shared.TestMetricName, 42),
			setup: func(tc *TestCase, repo *mock.MetricRepository) {
				repo.On("Get", tc.metric.Type(), tc.metric.Name()).Return(tc.metric, nil).Once()
			},
			wantErr: false,
		},
		{
			name:   "error",
			metric: domain.NewCounterMetric(shared.TestMetricName, 42),
			setup: func(tc *TestCase, repo *mock.MetricRepository) {
				err := errors.New("test error")
				repo.On("Get", tc.metric.Type(), tc.metric.Name()).Return(nil, err).Once()
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repository := new(mock.MetricRepository)
			tc.setup(&tc, repository)

			useCase := usecase.NewGetMetricUseCase(repository)
			metric, err := useCase.Execute(tc.metric.Type(), tc.metric.Name())

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Same(t, tc.metric, metric)
			}
		})
	}
}

func TestGetAllMetricsUseCase(t *testing.T) {
	type TestCase struct {
		name    string
		metrics []domain.Metric
		setup   func(*TestCase, *mock.MetricRepository)
		wantErr bool
	}

	tests := []TestCase{
		{
			name: "success",
			metrics: []domain.Metric{
				domain.NewCounterMetric("CounterMetric", 67),
				domain.NewGaugeMetric("GaugeMetric", 3.14),
			},
			setup: func(tc *TestCase, repo *mock.MetricRepository) {
				repo.On("GetAll").Return(tc.metrics, nil).Once()
			},
			wantErr: false,
		},
		{
			name:    "error",
			metrics: nil,
			setup: func(tc *TestCase, repo *mock.MetricRepository) {
				err := errors.New("test error")
				repo.On("GetAll").Return(tc.metrics, err).Once()
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repository := new(mock.MetricRepository)
			tc.setup(&tc, repository)

			useCase := usecase.NewGetAllMetricsUseCase(repository)
			metrics, err := useCase.Execute()

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.ElementsMatch(t, tc.metrics, metrics)
			}
		})
	}
}

func TestCollectMetricsUseCase(t *testing.T) {
	type TestCase struct {
		name  string
		setup func(*TestCase, *mock.MetricCollector, *mock.MetricRepository)
		err   error
	}

	tests := []TestCase{
		{
			name: "success",
			setup: func(
				_ *TestCase,
				collector *mock.MetricCollector,
				repository *mock.MetricRepository,
			) {
				collector.On("Collect").Return(metricsNonEmpty, nil).Once()
				repository.On("UpdateBatch", metricsNonEmpty).Return(nil).Once()
			},
			err: nil,
		},
		{
			name: "empty metrics",
			setup: func(
				_ *TestCase,
				collector *mock.MetricCollector,
				_ *mock.MetricRepository,
			) {
				collector.On("Collect").Return(metricsEmpty, nil).Once()
				// repository must not be called
			},
			err: nil,
		},
		{
			name: "collect error",
			setup: func(
				tc *TestCase,
				collector *mock.MetricCollector,
				_ *mock.MetricRepository,
			) {
				collector.On("Collect").Return(metricsEmpty, tc.err).Once()
				// repository must not be called
			},
			err: errors.New("collect error"),
		},
		{
			name: "update error",
			setup: func(
				tc *TestCase,
				collector *mock.MetricCollector,
				repository *mock.MetricRepository,
			) {
				collector.On("Collect").Return(metricsNonEmpty, nil).Once()
				repository.On("UpdateBatch", metricsNonEmpty).Return(tc.err).Once()
			},
			err: errors.New("collect error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			collector := new(mock.MetricCollector)
			repository := new(mock.MetricRepository)

			tc.setup(&tc, collector, repository)

			useCase := usecase.NewCollectMetricsUseCase(collector, repository)
			err := useCase.Execute()

			assert.ErrorIs(t, err, tc.err)
			collector.AssertExpectations(t)
			repository.AssertExpectations(t)
		})
	}
}

func TestSendMetricsUseCase(t *testing.T) {
	type TestCase struct {
		name  string
		setup func(*TestCase, *mock.MetricGateway, *mock.MetricRepository)
		err   error
	}

	tests := []TestCase{
		{
			name: "success",
			setup: func(
				_ *TestCase,
				gateway *mock.MetricGateway,
				repository *mock.MetricRepository,
			) {
				getAllCall := repository.On("GetAll").Return(metricsNonEmpty, nil).Once()
				repository.On("Clear").Return(nil).NotBefore(getAllCall).Once()
				gateway.On("SendBatch", metricsNonEmpty).Return(nil).Once()
			},
			err: nil,
		},
		{
			name: "empty metrics",
			setup: func(
				_ *TestCase,
				_ *mock.MetricGateway,
				repository *mock.MetricRepository,
			) {
				repository.On("GetAll").Return(metricsEmpty, nil).Once()
			},
			err: nil,
		},
		{
			name: "get error",
			setup: func(
				tc *TestCase,
				_ *mock.MetricGateway,
				repository *mock.MetricRepository,
			) {
				repository.On("GetAll").Return(metricsEmpty, tc.err).Once()
			},
			err: errors.New("get error"),
		},
		{
			name: "clear error",
			setup: func(
				tc *TestCase,
				_ *mock.MetricGateway,
				repository *mock.MetricRepository,
			) {
				getAllCall := repository.On("GetAll").Return(metricsNonEmpty, nil).Once()
				repository.On("Clear").Return(tc.err).NotBefore(getAllCall).Once()
			},
			err: errors.New("clear error"),
		},
		{
			name: "send error",
			setup: func(
				tc *TestCase,
				gateway *mock.MetricGateway,
				repository *mock.MetricRepository,
			) {
				getAllCall := repository.On("GetAll").Return(metricsNonEmpty, nil).Once()
				repository.On("Clear").Return(nil).NotBefore(getAllCall).Once()
				gateway.On("SendBatch", metricsNonEmpty).Return(tc.err).Once()
			},
			err: errors.New("send error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gateway := new(mock.MetricGateway)
			repository := new(mock.MetricRepository)

			tc.setup(&tc, gateway, repository)

			useCase := usecase.NewSendMetricsUseCase(gateway, repository)
			err := useCase.Execute()

			assert.ErrorIs(t, err, tc.err)
			gateway.AssertExpectations(t)
			repository.AssertExpectations(t)
		})
	}
}
