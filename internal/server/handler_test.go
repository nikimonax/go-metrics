package server_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/nikimonax/go-metrics/internal/domain"
	"github.com/nikimonax/go-metrics/internal/lib/httpextra"
	"github.com/nikimonax/go-metrics/internal/mock"
	"github.com/nikimonax/go-metrics/internal/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testMetricName = "TestMetric"

func TestUpdateMetricHandler(t *testing.T) {
	type TestCase struct {
		name            string
		method          string
		metricType      string
		metricName      string
		metricValue     string
		contentType     string
		setup           func(*TestCase, *mock.UpdateMetricUseCase)
		wantStatus      int
		wantContentType string
	}

	tests := []TestCase{
		{
			name:        "counter success",
			method:      http.MethodPost,
			metricType:  "counter",
			metricName:  testMetricName,
			metricValue: "42",
			contentType: httpextra.MIMEText,
			setup: func(tc *TestCase, useCase *mock.UpdateMetricUseCase) {
				wantMetricValue := domain.NewCounterMetric(testMetricName, 42)
				useCase.On("Execute", wantMetricValue).Return(nil).Once()
			},
			wantStatus:      http.StatusOK,
			wantContentType: httpextra.MIMEText,
		},
		{
			name:        "gauge success",
			method:      http.MethodPost,
			metricType:  "gauge",
			metricName:  testMetricName,
			metricValue: "3.14",
			contentType: httpextra.MIMEText,
			setup: func(tc *TestCase, useCase *mock.UpdateMetricUseCase) {
				wantMetricValue := domain.NewGaugeMetric(testMetricName, 3.14)
				useCase.On("Execute", wantMetricValue).Return(nil).Once()
			},
			wantStatus:      http.StatusOK,
			wantContentType: httpextra.MIMEText,
		},
		{
			name:        "update error",
			method:      http.MethodPost,
			metricType:  "counter",
			metricName:  testMetricName,
			metricValue: "42",
			contentType: httpextra.MIMEText,
			setup: func(tc *TestCase, useCase *mock.UpdateMetricUseCase) {
				wantMetricValue := domain.NewCounterMetric(testMetricName, 42)
				useCaseErr := errors.New("test error")
				useCase.On("Execute", wantMetricValue).Return(useCaseErr).Once()
			},
			wantStatus:      http.StatusInternalServerError,
			wantContentType: httpextra.MIMEText,
		},
		{
			name:            "invalid content type",
			method:          http.MethodPost,
			metricType:      "counter",
			metricName:      testMetricName,
			metricValue:     "42",
			contentType:     httpextra.MIMEJSON,
			wantStatus:      http.StatusBadRequest,
			wantContentType: httpextra.MIMEText,
		},
		{
			name:            "invalid counter metric value",
			method:          http.MethodPost,
			metricType:      "counter",
			metricName:      testMetricName,
			metricValue:     "foo42",
			contentType:     httpextra.MIMEText,
			wantStatus:      http.StatusBadRequest,
			wantContentType: httpextra.MIMEText,
		},
		{
			name:            "invalid gauge metric value",
			method:          http.MethodPost,
			metricType:      "gauge",
			metricName:      testMetricName,
			metricValue:     "foo3.14",
			contentType:     httpextra.MIMEText,
			wantStatus:      http.StatusBadRequest,
			wantContentType: httpextra.MIMEText,
		},
		{
			name:            "invalid metric type",
			method:          http.MethodPost,
			metricType:      "foobar",
			metricName:      testMetricName,
			metricValue:     "43",
			contentType:     httpextra.MIMEText,
			wantStatus:      http.StatusBadRequest,
			wantContentType: httpextra.MIMEText,
		},
		{
			name:            "empty metric type",
			method:          http.MethodPost,
			metricType:      "",
			metricName:      testMetricName,
			metricValue:     "42",
			contentType:     httpextra.MIMEText,
			wantStatus:      http.StatusBadRequest,
			wantContentType: httpextra.MIMEText,
		},
		{
			name:            "empty metric name",
			method:          http.MethodPost,
			metricType:      "counter",
			metricName:      "",
			metricValue:     "42",
			contentType:     httpextra.MIMEText,
			wantStatus:      http.StatusBadRequest,
			wantContentType: httpextra.MIMEText,
		},
		{
			name:            "empty metric counter value",
			method:          http.MethodPost,
			metricType:      "counter",
			metricName:      testMetricName,
			metricValue:     "",
			contentType:     httpextra.MIMEText,
			wantStatus:      http.StatusBadRequest,
			wantContentType: httpextra.MIMEText,
		},
		{
			name:            "empty metric gauge value",
			method:          http.MethodPost,
			metricType:      "gauge",
			metricName:      testMetricName,
			metricValue:     "",
			contentType:     httpextra.MIMEText,
			wantStatus:      http.StatusBadRequest,
			wantContentType: httpextra.MIMEText,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			useCase := new(mock.UpdateMetricUseCase)

			if tc.setup != nil {
				tc.setup(&tc, useCase)
			}

			handler := server.NewUpdateMetricHandler(useCase)

			req := httptest.NewRequest(tc.method, "/update", nil) // nolint:noctx
			rr := httptest.NewRecorder()

			chiCtx := chi.NewRouteContext()
			chiCtx.URLParams.Add("metricType", string(tc.metricType))
			chiCtx.URLParams.Add("metricName", string(tc.metricName))
			chiCtx.URLParams.Add("metricValue", tc.metricValue)

			ctx := context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx)
			req = req.WithContext(ctx)

			if tc.contentType != "" {
				req.Header.Add(httpextra.HDRContentType, tc.contentType)
			}

			handler.ServeHTTP(rr, req)

			resp := rr.Result()
			defer resp.Body.Close()

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
