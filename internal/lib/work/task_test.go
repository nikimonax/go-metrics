package work

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTask(t *testing.T) {
	t.Run("creates task", func(t *testing.T) {
		callback := func(context.Context) error { return nil }
		task, err := NewTask("task", callback)

		require.NoError(t, err)
		assert.Equal(t, "task", task.Name)
		assert.NotNil(t, task.Callback)
	})

	t.Run("rejects nil callback", func(t *testing.T) {
		_, err := NewTask("task", nil)

		assert.EqualError(t, err, "required non-nil callback for task")
	})
}
