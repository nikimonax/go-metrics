package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/nikimonax/go-metrics/internal/domain"
	"github.com/nikimonax/go-metrics/internal/lib/httpextra"
	"github.com/nikimonax/go-metrics/internal/model"
	"github.com/nikimonax/go-metrics/internal/testing/shared"
)

type ServerClient struct {
	http    *http.Client
	baseURL *url.URL
}

func (client *ServerClient) Update(
	ctx context.Context,
	metrics []domain.Metric,
) (err error) {
	var content bytes.Buffer

	var payload []*model.Metric

	for _, metric := range metrics {
		payload = append(payload, model.NewMetricFromDomain(metric))
	}

	if err := json.NewEncoder(&content).Encode(payload); err != nil {
		return err
	}

	updateRequest, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		client.baseURL.JoinPath("/updates").String(),
		&content,
	)

	if err != nil {
		return err
	}

	updateRequest.Header.Set(httpextra.HDRContentType, httpextra.MIMEJSON)

	resp, err := client.http.Do(updateRequest)

	if err != nil {
		return err
	}

	defer func() {
		_, readErr := io.Copy(io.Discard, resp.Body)
		err = errors.Join(err, readErr, resp.Body.Close())
	}()

	if resp.StatusCode >= 400 {
		reason := "unknown"

		if resp.ContentLength > 0 {
			if body, err := io.ReadAll(resp.Body); err == nil {
				reason = string(body)
			}
		}

		return httpextra.NewErrHTTPStatus(resp.StatusCode, reason)
	}

	return nil
}

func (client *ServerClient) Get(
	ctx context.Context,
	metricType domain.MetricType,
	metricName domain.MetricName,
) (metric domain.Metric, err error) {
	var content bytes.Buffer

	payload := model.MetricInfo{Name: metricName, Type: metricType}

	if err := json.NewEncoder(&content).Encode(payload); err != nil {
		return nil, err
	}

	getRequest, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		client.baseURL.JoinPath("/value").String(),
		&content,
	)

	if err != nil {
		return nil, err
	}

	getRequest.Header.Set(httpextra.HDRContentType, httpextra.MIMEJSON)

	resp, err := client.http.Do(getRequest)

	if err != nil {
		return nil, err
	}

	defer func() {
		err = errors.Join(err, resp.Body.Close())
	}()

	if resp.StatusCode >= 400 {
		reason := "unknown"

		if resp.ContentLength > 0 {
			if body, err := io.ReadAll(resp.Body); err == nil {
				reason = string(body)
			}
		}

		return nil, httpextra.NewErrHTTPStatus(resp.StatusCode, reason)
	} else if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("not expected status code: %d", resp.StatusCode)
	}

	var decoded model.Metric

	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return nil, err
	}

	return decoded.ToDomain(), nil
}

func NewServerClient(baseURL *url.URL) *ServerClient {
	return &ServerClient{
		http:    new(http.Client),
		baseURL: baseURL,
	}
}

func Example() {
	baseURL, shutdown := shared.NewTestServer(nil)
	defer shutdown()

	client := NewServerClient(baseURL)

	var (
		err    error
		metric domain.Metric
	)

	// update metrics
	metrics := []domain.Metric{
		domain.NewGaugeMetric("TestMetricGauge", 3.14),
		domain.NewCounterMetric("TestMetricCounter", 1),
	}

	err = client.Update(context.Background(), metrics)

	if err != nil {
		log.Fatalln(err)
	}

	// read metrics
	metric, err = client.Get(context.Background(), domain.Gauge, "TestMetricGauge")

	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("%+v\n", metric)

	metric, err = client.Get(context.Background(), domain.Counter, "TestMetricCounter")

	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("%+v\n", metric)

	// Output:
	// &{name:TestMetricGauge value:3.14}
	// &{name:TestMetricCounter value:1}
}
