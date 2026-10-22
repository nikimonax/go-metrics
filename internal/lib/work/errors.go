package work

import (
	"errors"
	"fmt"
)

var (
	// Возвращается при вызове операции в недопустимом
	// состоянии (например, запуск уже запущенного объекта).
	ErrInvalidState = errors.New("invalid state")

	// Возвращается при некорректной конфигурации.
	ErrInvalidConfig = errors.New("invalid config")
)

func newErrInvalidState(msg string) error {
	return fmt.Errorf("%w: %s", ErrInvalidState, msg)
}

func newErrInvalidConfig(msg string) error {
	return fmt.Errorf("%w: %s", ErrInvalidConfig, msg)
}
