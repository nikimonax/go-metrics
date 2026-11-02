package httpextra

import (
	"io"
	"net/http"
	"net/http/httptest"
)

type RoundTripperFunc func(*http.Request) (*http.Response, error)

func (f RoundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type encoderFunc func(w io.Writer, level int) io.Writer

func NewRoundTripperFromHandler(handler http.Handler) http.RoundTripper {
	fn := func(r *http.Request) (*http.Response, error) {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		return w.Result(), nil
	}
	return RoundTripperFunc(fn)
}
