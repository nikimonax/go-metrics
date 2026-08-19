package work

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidState  = errors.New("invalid state")
	ErrInvalidConfig = errors.New("invalid config")
)

func newErrInvalidState(msg string) error {
	return fmt.Errorf("%w: %s", ErrInvalidState, msg)
}

func newErrInvalidConfig(msg string) error {
	return fmt.Errorf("%w: %s", ErrInvalidConfig, msg)
}
