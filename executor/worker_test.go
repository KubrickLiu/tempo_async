package executor

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestWorkerRun(t *testing.T) {
	worker := NewWorker(1)
	worker.run()

	task := NewTask[string](func() (string, error) {
		time.Sleep(100 * time.Millisecond)
		return "mock", nil
	})
	future := task.Future()

	worker.append(task)

	ret, err := future.Join()
	worker.finish()

	assert.Equal(t, ret, "mock")
	assert.Empty(t, err)
}
