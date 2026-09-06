package gateway

import (
	"net/http"
	"net/url"

	"github.com/nikimonax/go-metrics/internal/app/interfaces"
)

type GatewayFactory = func(
	client *http.Client,
	baseURL *url.URL,
) interfaces.MetricGateway

var gatewayFactoryByVersion = []GatewayFactory{
	NewHTTPMetricGateway,
	NewHTTPMetricV2Gateway,
	NewHTTPMetricV3Gateway,
}

func GetMaxAPIVersion() uint {
	return uint(len(gatewayFactoryByVersion))
}

func GetGatewayFactory(
	apiVersion uint,
) (GatewayFactory, error) {
	if apiVersion > GetMaxAPIVersion() {
		err := NewErrUnknownAPIVersion(apiVersion)
		return nil, err
	}

	return gatewayFactoryByVersion[apiVersion-1], nil
}
