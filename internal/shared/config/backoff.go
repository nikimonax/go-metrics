package config

import (
	"time"

	"github.com/sethvargo/go-retry"
)

type BackoffConfig struct {
	Retry int64
	Seed  time.Duration
	Add   time.Duration
	Mul   time.Duration
}

func (cfg BackoffConfig) Build() retry.Backoff {
	if cfg.Retry == 0 {
		return retry.BackoffFunc(func() (time.Duration, bool) {
			return 0, true
		})
	}

	if cfg.Mul == 0 {
		cfg.Mul = 1
	}

	var (
		wait  time.Duration
		calls int64
	)

	return retry.BackoffFunc(func() (time.Duration, bool) {
		if calls >= cfg.Retry {
			return 0, true
		}

		calls++

		if wait == 0 && cfg.Seed != 0 {
			wait = cfg.Seed
		} else {
			wait = wait*cfg.Mul + cfg.Add
		}

		return wait, false
	})

}
