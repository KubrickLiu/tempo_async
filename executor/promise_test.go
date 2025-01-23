package executor

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPromiseSuccess(t *testing.T) {
	promise := NewPromise[string]()
	future := promise.Future()

	expect := "mock"
	promise.Success(expect)

	ret, _ := future.Join()
	assert.Equal(t, expect, ret)
}
