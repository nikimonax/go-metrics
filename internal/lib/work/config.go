package work

import "time"

// Содержит параметры жизненного цикла, общие для
// пула воркеров и планировщика.
type LifecycleConfig struct {
	// Максимальное время ожидания завершения остановки.
	// Если равен нулю, ожидание выполняется без таймаута.
	StopTimeout time.Duration
}

// Проверяет корректность значений конфигурации lifecycle.
func (c LifecycleConfig) Validate() error {
	if c.StopTimeout < 0 {
		return newErrInvalidConfig("stop timeout must equal or greater than 0")
	}
	return nil
}

// Содержит колбэки для обработки событий воркера.
type WorkerConfig struct {
	// Вызывается при ошибке выполнения задачи.
	OnError func(string, error)
	// Вызывается при панике во время выполнения задачи.
	OnPanic func(string, any)
}

// Содержит конфигурацию пула воркеров.
type PoolConfig struct {
	WorkerConfig
	LifecycleConfig

	// Количество воркеров в пуле.
	WorkerCount uint
	// Ёмкость очереди задач.
	QueueSize uint
}

// Проверяет корректность значений конфигурации пула.
func (c PoolConfig) Validate() error {
	if c.WorkerCount == 0 {
		return newErrInvalidConfig("worker count must be greater than 0")
	}

	if err := c.LifecycleConfig.Validate(); err != nil {
		return err
	}

	return nil
}

// Содержит конфигурацию планировщика.
type SchedulerConfig struct {
	LifecycleConfig

	// Вызывается при ошибке постановки задачи в пул.
	OnError func(string, error)
}

// Проверяет корректность значений конфигурации планировщика.
func (c SchedulerConfig) Validate() error {
	return c.LifecycleConfig.Validate()
}
