package config

import (
	"errors"
	"fmt"
)

var ErrInvalidConfig = errors.New("invalid config")

func newErrInvalidConfig(msg string) error {
	return fmt.Errorf("%w: %s", ErrInvalidConfig, msg)
}
