package gateway

import (
	"net/url"

	"github.com/nikimonax/go-metrics/internal/app/interfaces"
)

var gatewayFactoryByVersion = []func(baseURL *url.URL) interfaces.MetricGateway{
	NewHTTPMetricGateway,
	NewHTTPMetricV2Gateway,
	NewHTTPMetricV3Gateway,
}

func GetMaxAPIVersion() uint {
	return uint(len(gatewayFactoryByVersion))
}

func GetGatewayFactory(
	apiVersion uint,
) (func(*url.URL) interfaces.MetricGateway, error) {
	if apiVersion > GetMaxAPIVersion() {
		err := NewErrUnknownAPIVersion(apiVersion)
		return nil, err
	}

	return gatewayFactoryByVersion[apiVersion-1], nil
}
