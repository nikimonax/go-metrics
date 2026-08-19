package work

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type Validatable interface {
	Validate() error
}

func TestConfig(t *testing.T) {
	type TestCase struct {
		name    string
		config  Validatable
		wantErr bool
	}

	tests := []TestCase{
		{
			name:    "lifecycle: zero timeout",
			config:  LifecycleConfig{StopTimeout: 0},
			wantErr: false,
		},
		{
			name:    "lifecycle: positive timeout",
			config:  LifecycleConfig{StopTimeout: time.Second},
			wantErr: false,
		},
		{
			name:    "lifecycle: negative timeout",
			config:  LifecycleConfig{StopTimeout: -time.Second},
			wantErr: true,
		},
		{
			name:    "pool: valid",
			config:  PoolConfig{WorkerCount: 1},
			wantErr: false,
		},
		{
			name:    "pool: zero workers",
			config:  PoolConfig{},
			wantErr: true,
		},
		{
			name: "pool: negative timeout",
			config: PoolConfig{
				WorkerCount:     1,
				LifecycleConfig: LifecycleConfig{StopTimeout: -1},
			},
			wantErr: true,
		},
		{
			name:    "scheduler: valid",
			config:  SchedulerConfig{},
			wantErr: false,
		},
		{
			name:    "scheduler: negative timeout",
			config:  SchedulerConfig{LifecycleConfig: LifecycleConfig{StopTimeout: -1}},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.config.Validate()

			if tc.wantErr {
				assert.ErrorIs(t, err, ErrInvalidConfig)
				return
			}

			assert.NoError(t, err)
		})
	}
}
