package work

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchedule(t *testing.T) {
	callback := func(context.Context) error { return nil }
	task, err := NewTask("task", callback)
	require.NoError(t, err)

	t.Run("creates schedule", func(t *testing.T) {
		schedule, err := NewSchedule(task, time.Second)

		require.NoError(t, err)
		assert.Equal(t, task.Name, schedule.Name)
		assert.Equal(t, time.Second, schedule.Interval)
	})

	t.Run("rejects negative timeout", func(t *testing.T) {
		_, err = NewSchedule(task, -time.Second)
		assert.Error(t, err)
	})
}
