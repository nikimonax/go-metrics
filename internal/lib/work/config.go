package work

import "time"

type LifecycleConfig struct {
	StopTimeout time.Duration
}

func (c LifecycleConfig) Validate() error {
	if c.StopTimeout < 0 {
		return newErrInvalidConfig("stop timeout must equal or greater than 0")
	}
	return nil
}

type WorkerConfig struct {
	OnError func(string, error)
	OnPanic func(string, any)
}

type PoolConfig struct {
	WorkerConfig
	LifecycleConfig

	WorkerCount uint
	QueueSize   uint
}

func (c PoolConfig) Validate() error {
	if c.WorkerCount == 0 {
		return newErrInvalidConfig("worker count must be greater than 0")
	}

	if err := c.LifecycleConfig.Validate(); err != nil {
		return err
	}

	return nil
}

type SchedulerConfig struct {
	LifecycleConfig

	OnError func(string, error)
}

func (c SchedulerConfig) Validate() error {
	return c.LifecycleConfig.Validate()
}
