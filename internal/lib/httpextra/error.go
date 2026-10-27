package httpextra

import "fmt"

type HTTPStatusError struct {
	StatusCode int
	Reason     string
}

var _ error = (*HTTPStatusError)(nil)

func NewErrHTTPStatus(statusCode int, reason string) *HTTPStatusError {
	return &HTTPStatusError{
		StatusCode: statusCode,
		Reason:     reason,
	}
}

func (e *HTTPStatusError) Error() string {
	return fmt.Sprintf("api error (%d): %s", e.StatusCode, e.Reason)
}
