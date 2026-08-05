package handler_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/nikimonax/go-metrics/internal/app"
	"github.com/nikimonax/go-metrics/internal/domain"
	"github.com/nikimonax/go-metrics/internal/lib/httpextra"
	"github.com/nikimonax/go-metrics/internal/server/handler"
	"github.com/nikimonax/go-metrics/internal/server/presenter"
	"github.com/nikimonax/go-metrics/internal/testing/mock"
	"github.com/nikimonax/go-metrics/internal/testing/shared"
	"github.com/stretchr/testify/assert"
	m "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUpdateMetricHandler(t *testing.T) {
	type TestCase struct {
		name            string
		method          string
		metricType      domain.MetricType
		metricName      domain.MetricName
		metricValue     string
		setup           func(*TestCase, *mock.UpdateMetricUseCase)
		wantStatus      int
		wantContentType string
	}

	tests := []TestCase{
		{
			name:        "counter success",
			method:      http.MethodPost,
			metricType:  domain.Counter,
			metricName:  shared.TestMetricName,
			metricValue: "42",
			setup: func(_ *TestCase, useCase *mock.UpdateMetricUseCase) {
				wantMetricValue := domain.NewCounterMetric(shared.TestMetricName, 42)
				useCase.On("Execute", wantMetricValue).Return(nil).Once()
			},
			wantStatus:      http.StatusOK,
			wantContentType: httpextra.MIMEText,
		},
		{
			name:        "gauge success",
			method:      http.MethodPost,
			metricType:  domain.Gauge,
			metricName:  shared.TestMetricName,
			metricValue: "3.14",
			setup: func(_ *TestCase, useCase *mock.UpdateMetricUseCase) {
				wantMetricValue := domain.NewGaugeMetric(shared.TestMetricName, 3.14)
				useCase.On("Execute", wantMetricValue).Return(nil).Once()
			},
			wantStatus:      http.StatusOK,
			wantContentType: httpextra.MIMEText,
		},
		{
			name:        "update error",
			method:      http.MethodPost,
			metricType:  domain.Counter,
			metricName:  shared.TestMetricName,
			metricValue: "42",
			setup: func(_ *TestCase, useCase *mock.UpdateMetricUseCase) {
				wantMetricValue := domain.NewCounterMetric(shared.TestMetricName, 42)
				useCaseErr := errors.New("test error")
				useCase.On("Execute", wantMetricValue).Return(useCaseErr).Once()
			},
			wantStatus:      http.StatusInternalServerError,
			wantContentType: httpextra.MIMEText,
		},
		{
			name:            "invalid counter metric value",
			method:          http.MethodPost,
			metricType:      domain.Counter,
			metricName:      shared.TestMetricName,
			metricValue:     "foo42",
			wantStatus:      http.StatusBadRequest,
			wantContentType: httpextra.MIMEText,
		},
		{
			name:            "invalid gauge metric value",
			method:          http.MethodPost,
			metricType:      domain.Gauge,
			metricName:      shared.TestMetricName,
			metricValue:     "foo3.14",
			wantStatus:      http.StatusBadRequest,
			wantContentType: httpextra.MIMEText,
		},
		{
			name:            "invalid metric type",
			method:          http.MethodPost,
			metricType:      domain.MetricType("other"),
			metricName:      shared.TestMetricName,
			metricValue:     "43",
			wantStatus:      http.StatusBadRequest,
			wantContentType: httpextra.MIMEText,
		},
		{
			name:            "empty metric type",
			method:          http.MethodPost,
			metricName:      shared.TestMetricName,
			metricValue:     "42",
			wantStatus:      http.StatusBadRequest,
			wantContentType: httpextra.MIMEText,
		},
		{
			name:            "empty metric name",
			method:          http.MethodPost,
			metricType:      domain.Counter,
			metricValue:     "42",
			wantStatus:      http.StatusBadRequest,
			wantContentType: httpextra.MIMEText,
		},
		{
			name:            "empty metric counter value",
			method:          http.MethodPost,
			metricType:      domain.Counter,
			metricName:      shared.TestMetricName,
			wantStatus:      http.StatusBadRequest,
			wantContentType: httpextra.MIMEText,
		},
		{
			name:            "empty metric gauge value",
			method:          http.MethodPost,
			metricType:      domain.Gauge,
			metricName:      shared.TestMetricName,
			wantStatus:      http.StatusBadRequest,
			wantContentType: httpextra.MIMEText,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			useCase := new(mock.UpdateMetricUseCase)
			errorPresenter := presenter.NewPlainTextErrorPresenter()

			if tc.setup != nil {
				tc.setup(&tc, useCase)
			}

			handler := handler.NewUpdateMetricHandler(useCase, errorPresenter)

			req := httptest.NewRequestWithContext(context.Background(), tc.method, "/update", nil)
			rr := httptest.NewRecorder()

			chiCtx := chi.NewRouteContext()
			chiCtx.URLParams.Add("metricType", string(tc.metricType))
			chiCtx.URLParams.Add("metricName", string(tc.metricName))
			chiCtx.URLParams.Add("metricValue", tc.metricValue)

			ctx := context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx)
			req = req.WithContext(ctx)

			handler.ServeHTTP(rr, req)

			resp := rr.Result()
			defer func() { assert.NoError(t, resp.Body.Close()) }()

			body, err := io.ReadAll(resp.Body)

			actualContentType := resp.Header.Get(httpextra.HDRContentType)

			require.NoError(t, err)
			assert.Equal(t, tc.wantStatus, resp.StatusCode)
			assert.Contains(t, actualContentType, tc.wantContentType)

			if resp.StatusCode < 400 {
				assert.Empty(t, string(body))
			} else {
				assert.NotEmpty(t, string(body))
			}

			useCase.AssertExpectations(t)
		})
	}
}

func TestGetMetricHandler(t *testing.T) {
	type TestCase struct {
		name       string
		method     string
		metricType domain.MetricType
		metricName domain.MetricName
		setup      func(
			*TestCase,
			*mock.GetMetricUseCase,
			*mock.ErrorPresenter,
			*mock.MetricPresenter,
		)
	}

	tests := []TestCase{
		{
			name:       "success",
			method:     http.MethodGet,
			metricType: domain.Counter,
			metricName: shared.TestMetricName,
			setup: func(
				tc *TestCase,
				useCase *mock.GetMetricUseCase,
				_ *mock.ErrorPresenter,
				metricPresenter *mock.MetricPresenter,
			) {
				metric := domain.NewCounterMetric(tc.metricName, 42)
				useCase.On(
					"Execute",
					tc.metricType,
					tc.metricName,
				).Return(metric, nil).Once()
				metricPresenter.On(
					"Render",
					m.MatchedBy(func(arg any) bool {
						_, ok := arg.(http.ResponseWriter)
						return ok
					}),
					metric,
					http.StatusOK,
				).Return().Once()
			},
		},
		{
			name:       "usecase error",
			method:     http.MethodGet,
			metricType: domain.Counter,
			metricName: shared.TestMetricName,
			setup: func(
				tc *TestCase,
				useCase *mock.GetMetricUseCase,
				errorPresenter *mock.ErrorPresenter,
				_ *mock.MetricPresenter,
			) {
				err := errors.New("test error")
				useCase.On(
					"Execute",
					tc.metricType,
					tc.metricName,
				).Return(nil, err).Once()
				errorPresenter.On(
					"Render",
					m.MatchedBy(func(arg any) bool {
						_, ok := arg.(http.ResponseWriter)
						return ok
					}),
					m.MatchedBy(func(arg any) bool {
						_, ok := arg.(error)
						return ok
					}),
					http.StatusInternalServerError,
				).Return().Once()
			},
		},
		{
			name:       "metric not found",
			method:     http.MethodGet,
			metricType: domain.Counter,
			metricName: shared.TestMetricName,
			setup: func(
				tc *TestCase,
				useCase *mock.GetMetricUseCase,
				errorPresenter *mock.ErrorPresenter,
				_ *mock.MetricPresenter,
			) {
				useCase.On(
					"Execute",
					tc.metricType,
					tc.metricName,
				).Return(nil, app.ErrMetricNotFound).Once()
				errorPresenter.On(
					"Render",
					m.MatchedBy(func(arg any) bool {
						_, ok := arg.(http.ResponseWriter)
						return ok
					}),
					m.MatchedBy(func(arg any) bool {
						_, ok := arg.(error)
						return ok
					}),
					http.StatusNotFound,
				).Return().Once()
			},
		},
		{
			name:       "empty metric name",
			method:     http.MethodGet,
			metricType: domain.Counter,
			metricName: "",
			setup: func(
				_ *TestCase,
				_ *mock.GetMetricUseCase,
				errorPresenter *mock.ErrorPresenter,
				_ *mock.MetricPresenter,
			) {
				errorPresenter.On(
					"Render",
					m.MatchedBy(func(arg any) bool {
						_, ok := arg.(http.ResponseWriter)
						return ok
					}),
					m.MatchedBy(func(arg any) bool {
						_, ok := arg.(error)
						return ok
					}),
					http.StatusBadRequest,
				).Return().Once()
			},
		},
		{
			name:       "invalid metric type",
			method:     http.MethodGet,
			metricType: "foo",
			metricName: shared.TestMetricName,
			setup: func(
				_ *TestCase,
				_ *mock.GetMetricUseCase,
				errorPresenter *mock.ErrorPresenter,
				_ *mock.MetricPresenter,
			) {
				errorPresenter.On(
					"Render",
					m.MatchedBy(func(arg any) bool {
						_, ok := arg.(http.ResponseWriter)
						return ok
					}),
					m.MatchedBy(func(arg any) bool {
						_, ok := arg.(error)
						return ok
					}),
					http.StatusBadRequest,
				).Return().Once()
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			useCase := new(mock.GetMetricUseCase)
			errorPresenter := new(mock.ErrorPresenter)
			metricPresenter := new(mock.MetricPresenter)

			if tc.setup != nil {
				tc.setup(&tc, useCase, errorPresenter, metricPresenter)
			}

			handler := handler.NewGetMetricHandler(
				useCase,
				errorPresenter,
				metricPresenter,
			)

			req := httptest.NewRequestWithContext(context.Background(), tc.method, "/value", nil)
			rr := httptest.NewRecorder()

			chiCtx := chi.NewRouteContext()
			chiCtx.URLParams.Add("metricType", string(tc.metricType))
			chiCtx.URLParams.Add("metricName", string(tc.metricName))

			ctx := context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx)
			req = req.WithContext(ctx)

			handler.ServeHTTP(rr, req)

			useCase.AssertExpectations(t)
			errorPresenter.AssertExpectations(t)
			metricPresenter.AssertExpectations(t)
		})
	}
}

func TestPreviewMetricsHandler(t *testing.T) {
	type TestCase struct {
		name   string
		method string
		setup  func(
			*TestCase,
			*mock.GetAllMetricsUseCase,
			*mock.ErrorPresenter,
			*mock.MetricsPresenter,
		)
	}

	tests := []TestCase{
		{
			name:   "success",
			method: http.MethodGet,
			setup: func(
				_ *TestCase,
				useCase *mock.GetAllMetricsUseCase,
				_ *mock.ErrorPresenter,
				metricsPresenter *mock.MetricsPresenter,
			) {
				metrics := []domain.Metric{
					domain.NewCounterMetric("A", 42),
					domain.NewGaugeMetric("B", 3.14),
				}

				useCase.On("Execute").Return(metrics, nil).Once()
				metricsPresenter.On(
					"Render",
					m.MatchedBy(func(arg any) bool {
						_, ok := arg.(http.ResponseWriter)
						return ok
					}),
					metrics,
					http.StatusOK,
				)
			},
		},
		{
			name:   "usecase error",
			method: http.MethodGet,
			setup: func(
				_ *TestCase,
				useCase *mock.GetAllMetricsUseCase,
				errorPresenter *mock.ErrorPresenter,
				_ *mock.MetricsPresenter,
			) {
				err := errors.New("test error")

				useCase.On("Execute").Return(nil, err).Once()
				errorPresenter.On(
					"Render",
					m.MatchedBy(func(arg any) bool {
						_, ok := arg.(http.ResponseWriter)
						return ok
					}),
					err,
					http.StatusInternalServerError,
				)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			useCase := new(mock.GetAllMetricsUseCase)
			errorPresenter := new(mock.ErrorPresenter)
			metricsPresenter := new(mock.MetricsPresenter)

			if tc.setup != nil {
				tc.setup(&tc, useCase, errorPresenter, metricsPresenter)
			}

			handler := handler.NewPreviewMetricsHandler(
				useCase,
				errorPresenter,
				metricsPresenter,
			)

			req := httptest.NewRequestWithContext(context.Background(), tc.method, "/", nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			useCase.AssertExpectations(t)
			errorPresenter.AssertExpectations(t)
			metricsPresenter.AssertExpectations(t)
		})
	}
}
