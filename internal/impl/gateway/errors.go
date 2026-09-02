package gateway

import (
	"fmt"
)

type HTTPStatusError struct {
	StatusCode int
	Reason     string
}

func NewErrHTTPStatus(statusCode int, reason string) error {
	return &HTTPStatusError{
		StatusCode: statusCode,
		Reason:     reason,
	}
}

func (e *HTTPStatusError) Error() string {
	return fmt.Sprintf("api error (%d): %s", e.StatusCode, e.Reason)
}
