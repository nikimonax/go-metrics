package httpextra

import "net/http"

type ExtraHeadersRoundTripper struct {
	next    http.RoundTripper
	headers map[string]string
}

// RoundTrip implements [http.RoundTripper].
func (rt *ExtraHeadersRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	for key, value := range rt.headers {
		req.Header.Set(key, value)
	}

	return rt.next.RoundTrip(req)
}

var _ http.RoundTripper = (*ExtraHeadersRoundTripper)(nil)

func NewExtraHeadersRoundTripper(
	next http.RoundTripper,
	headers map[string]string,
) *ExtraHeadersRoundTripper {
	return &ExtraHeadersRoundTripper{next: next, headers: headers}
}
