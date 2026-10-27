package gateway

import (
	"errors"
	"fmt"
)

var ErrUnknownAPIVersion = errors.New("unknown api version")

func NewErrUnknownAPIVersion(apiVersion uint) error {
	return fmt.Errorf("%w: %d", ErrUnknownAPIVersion, apiVersion)
}
