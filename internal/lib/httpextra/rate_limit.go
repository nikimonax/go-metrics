package httpextra

import (
	"net/http"

	"golang.org/x/sync/semaphore"
)

type RateLimitRoundTripper struct {
	next http.RoundTripper
	sem  *semaphore.Weighted
}

// RoundTrip implements [http.RoundTripper].
func (rt *RateLimitRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if err := rt.sem.Acquire(req.Context(), 1); err != nil {
		return nil, err
	}
	defer rt.sem.Release(1)

	return rt.next.RoundTrip(req)
}

var _ http.RoundTripper = (*RateLimitRoundTripper)(nil)

func NewRateLimitRoundTripper(
	next http.RoundTripper,
	limit int64,
) *RateLimitRoundTripper {
	return &RateLimitRoundTripper{
		next: next,
		sem:  semaphore.NewWeighted(limit),
	}
}
