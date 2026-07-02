package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/nikimonax/go-metrics/pkg"
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

func parseMetricFromRequest(r *http.Request) (metric pkg.Metric, err error) {
	var metricType pkg.MetricType

	if err = parseMetricType(r, &metricType); err != nil {
		return
	}

	var metricName pkg.MetricName

	if err = parseMetricName(r, &metricName); err != nil {
		return
	}

	switch metricType {
	case pkg.Counter:
		var metricValue int64
		if err = parseCounterMetricValue(r, &metricValue); err != nil {
			return
		}
		metric = pkg.NewCounterMetric(metricName, metricValue)
	case pkg.Gauge:
		var metricValue float64
		if err = parseGaugeMetricValue(r, &metricValue); err != nil {
			return
		}
		metric = pkg.NewGaugeMetric(metricName, metricValue)
	default:
		message := newErrMsgParamNotValid("metricType")
		err = errors.New(message)
	}

	return
}
