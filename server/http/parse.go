package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/nikimonax/go-metrics/pkg"
	"github.com/nikimonax/go-metrics/server/app"
)

func parseMetricType(r *http.Request, mt *pkg.MetricType) error {
	metricTypeRaw := r.PathValue("metricType")

	if metricTypeRaw == "" {
		message := newErrMsgParamNotProvided("Path", "metricType")
		return errors.New(message)
	}

	metricType := pkg.MetricType(metricTypeRaw)

	if !metricType.IsValid() {
		message := newErrMsgParamNotValid("metricType")
		return errors.New(message)
	}

	*mt = metricType
	return nil
}

func parseMetricName(r *http.Request, mn *pkg.MetricName) error {
	metricNameRaw := r.PathValue("metricName")

	if metricNameRaw == "" {
		message := newErrMsgParamNotProvided("Path", "metricName")
		return errors.New(message)
	}

	// TODO: ограничения на имя метрики?
	*mn = pkg.MetricName(metricNameRaw)
	return nil
}

func parseCounterMetricValue(r *http.Request, mv *int64) error {
	metricValueRaw := r.PathValue("metricValue")

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
	metricValueRaw := r.PathValue("metricValue")

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

func parseUpdateMetricDTO(r *http.Request, dto *app.UpdateMetricDTO) error {
	var err error

	err = parseMetricType(r, &dto.MetricType)

	if err != nil {
		return err
	}

	err = parseMetricName(r, &dto.MetricName)

	if err != nil {
		return err
	}

	switch dto.MetricType {
	case pkg.Counter:
		return parseCounterMetricValue(r, &dto.ValueAdd)
	case pkg.Gauge:
		return parseGaugeMetricValue(r, &dto.ValueSet)
	default:
		message := newErrMsgParamNotValid("metricType")
		return errors.New(message)
	}
}
