package server

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/nikimonax/go-metrics/internal/domain"
)

func parseMetricType(r *http.Request, mt *domain.MetricType) error {
	metricTypeRaw := chi.URLParam(r, "metricType")

	if metricTypeRaw == "" {
		message := newErrMsgParamNotProvided("Path", "metricType")
		return errors.New(message)
	}

	metricType := domain.MetricType(metricTypeRaw)

	if !metricType.IsValid() {
		message := newErrMsgParamNotValid("metricType")
		return errors.New(message)
	}

	*mt = metricType
	return nil
}

func parseMetricName(r *http.Request, mn *domain.MetricName) error {
	metricNameRaw := chi.URLParam(r, "metricName")

	if metricNameRaw == "" {
		message := newErrMsgParamNotProvided("Path", "metricName")
		return errors.New(message)
	}

	// TODO: ограничения на имя метрики?
	*mn = domain.MetricName(metricNameRaw)
	return nil
}

func parseCounterMetricValue(r *http.Request, mv *int64) error {
	metricValueRaw := chi.URLParam(r, "metricValue")

	if metricValueRaw == "" {
		message := newErrMsgParamNotProvided("Path", "metricValue")
		return errors.New(message)
	}

	metricValue, err := strconv.ParseInt(metricValueRaw, 10, 64)

	if err != nil {
		return err
	}

	*mv = metricValue
	return nil
}

func parseGaugeMetricValue(r *http.Request, mv *float64) error {
	metricValueRaw := chi.URLParam(r, "metricValue")

	if metricValueRaw == "" {
		message := newErrMsgParamNotProvided("Path", "metricValue")
		return errors.New(message)
	}

	metricValue, err := strconv.ParseFloat(metricValueRaw, 64)

	if err != nil {
		return err
	}

	*mv = metricValue
	return nil
}

func parseMetricFromRequest(r *http.Request) (metric domain.Metric, err error) {
	var metricType domain.MetricType

	if err = parseMetricType(r, &metricType); err != nil {
		return
	}

	var metricName domain.MetricName

	if err = parseMetricName(r, &metricName); err != nil {
		return
	}

	switch metricType {
	case domain.Counter:
		var metricValue int64
		if err = parseCounterMetricValue(r, &metricValue); err != nil {
			return
		}
		metric = domain.NewCounterMetric(metricName, metricValue)
	case domain.Gauge:
		var metricValue float64
		if err = parseGaugeMetricValue(r, &metricValue); err != nil {
			return
		}
		metric = domain.NewGaugeMetric(metricName, metricValue)
	default:
		message := newErrMsgParamNotValid("metricType")
		err = errors.New(message)
	}

	return
}
