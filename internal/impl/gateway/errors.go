package gateway

import (
	"errors"
	"fmt"
)

var ErrAPI = errors.New("API error")

func NewErrAPI(msg string) error {
	return fmt.Errorf("%w: %s", ErrAPI, msg)
}
