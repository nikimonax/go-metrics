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
	func(c *http.Client, u *url.URL) interfaces.MetricGateway {
		return NewHTTPMetricGateway(c, u)
	},
	func(c *http.Client, u *url.URL) interfaces.MetricGateway {
		return NewHTTPMetricV2Gateway(c, u)
	},
	func(c *http.Client, u *url.URL) interfaces.MetricGateway {
		return NewHTTPMetricV3Gateway(c, u)
	},
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
