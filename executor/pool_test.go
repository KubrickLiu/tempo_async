package executor

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestApply(t *testing.T) {
	pool := NewPool(1)
	w, _ := pool.Apply()

	assert.NotEmpty(t, w)
}

func TestRelease(t *testing.T) {
	pool := NewPool(1)

	w, _ := pool.Apply()
	assert.Equal(t, 1, pool.RunningNums())

	pool.Release(w)
	assert.Equal(t, 0, pool.RunningNums())
}
