package usecase_test

import (
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
				updateCall := repo.On(
					"Update",
					mock.MatchContext(),
					metric,
				).Return(nil).Once()
				pub.On(
					"Publish",
					mock.MatchContext(),
					mock.MatchedBy(mock.IsImplements(new(interfaces.Event))),
				).NotBefore(updateCall).Once()
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
				repo.On(
					"Update",
					mock.MatchContext(),
					metric,
				).Return(tc.err).Once()
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
			err := useCase.Execute(t.Context(), metric)

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
				repo.On(
					"Get",
					mock.MatchContext(),
					tc.metric.Type(),
					tc.metric.Name(),
				).Return(tc.metric, nil).Once()
			},
			wantErr: false,
		},
		{
			name:   "error",
			metric: domain.NewCounterMetric(shared.TestMetricName, 42),
			setup: func(tc *TestCase, repo *mock.MetricRepository) {
				err := errors.New("test error")
				repo.On(
					"Get",
					mock.MatchContext(),
					tc.metric.Type(),
					tc.metric.Name(),
				).Return(nil, err).Once()
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repository := new(mock.MetricRepository)
			tc.setup(&tc, repository)

			useCase := usecase.NewGetMetricUseCase(repository)
			metric, err := useCase.Execute(t.Context(), tc.metric.Type(), tc.metric.Name())

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
				repo.On(
					"GetAll",
					mock.MatchContext(),
				).Return(tc.metrics, nil).Once()
			},
			wantErr: false,
		},
		{
			name:    "error",
			metrics: nil,
			setup: func(tc *TestCase, repo *mock.MetricRepository) {
				err := errors.New("test error")
				repo.On(
					"GetAll",
					mock.MatchContext(),
				).Return(tc.metrics, err).Once()
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repository := new(mock.MetricRepository)
			tc.setup(&tc, repository)

			useCase := usecase.NewGetAllMetricsUseCase(repository)
			metrics, err := useCase.Execute(t.Context())

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
				collectCall := collector.On("Collect").Return(metricsNonEmpty, nil).Once()
				repository.On(
					"UpdateBatch",
					mock.MatchContext(),
					metricsNonEmpty,
				).Return(nil).NotBefore(collectCall).Once()
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
				collectCall := collector.On("Collect").Return(metricsNonEmpty, nil).Once()
				repository.On(
					"UpdateBatch",
					mock.MatchContext(),
					metricsNonEmpty,
				).Return(tc.err).NotBefore(collectCall).Once()
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
			err := useCase.Execute(t.Context())

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
				getAllCall := repository.On(
					"PopAll",
					mock.MatchContext(),
				).Return(metricsNonEmpty, nil).Once()
				gateway.On(
					"SendBatch",
					mock.MatchContext(),
					metricsNonEmpty,
				).Return(nil).NotBefore(getAllCall).Once()
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
				repository.On(
					"PopAll",
					mock.MatchContext(),
				).Return(metricsEmpty, nil).Once()
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
				repository.On(
					"PopAll",
					mock.MatchContext(),
				).Return(metricsEmpty, tc.err).Once()
			},
			err: errors.New("get error"),
		},
		{
			name: "send error",
			setup: func(
				tc *TestCase,
				gateway *mock.MetricGateway,
				repository *mock.MetricRepository,
			) {
				getAllCall := repository.On(
					"PopAll",
					mock.MatchContext(),
				).Return(metricsNonEmpty, nil).Once()
				gateway.On(
					"SendBatch",
					mock.MatchContext(),
					metricsNonEmpty,
				).Return(tc.err).NotBefore(getAllCall).Once()
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
			err := useCase.Execute(t.Context())

			assert.ErrorIs(t, err, tc.err)
			gateway.AssertExpectations(t)
			repository.AssertExpectations(t)
		})
	}
}
