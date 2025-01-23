package executor

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestJoinResult(t *testing.T) {
	future := newFuture[string]()

	expect := "hello"
	go func() {
		time.Sleep(100 * time.Millisecond)
		future.complete(expect)
	}()

	ret, _ := future.Join()
	assert.Equal(t, expect, ret)
}
