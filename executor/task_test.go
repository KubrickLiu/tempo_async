package executor

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestTaskSuccess(t *testing.T) {
	expectRet := "hello"

	task := NewTask[string](func() (string, error) {
		time.Sleep(1 * time.Second)
		return expectRet, nil
	})

	future := task.Future()
	go task.Call()
	ret, _ := future.Join()
	assert.Equal(t, expectRet, ret)
}

func TestTaskFailure(t *testing.T) {
	expectError := errors.New("mock error")

	task := NewTask[string](func() (string, error) {
		time.Sleep(1 * time.Second)
		return "", expectError
	})

	future := task.Future()
	go task.Call()
	_, err := future.Join()
	assert.EqualError(t, err, expectError.Error())
}

func TestTaskContextCancel(t *testing.T) {
	task := NewTask[string](func() (string, error) {
		time.Sleep(1 * time.Second)
		return "", nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	future := task.Future()
	go task.Call()

	_, err := future.Get(ctx)

	assert.EqualError(t, err, "context canceled")
}
