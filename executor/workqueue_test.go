package executor

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCommon(t *testing.T) {
	queue := newQueue()
	assert.Equal(t, 0, queue.len())

	flag := queue.acceptWorker(1, func() Worker {
		return NewWorker(1)
	})
	assert.Equal(t, true, flag)
	assert.Equal(t, 1, queue.len())
	assert.Equal(t, false, queue.isEmpty())
}

func TestDetach(t *testing.T) {
	queue := newQueue()
	queue.acceptWorker(1, func() Worker {
		return NewWorker(1)
	})

	w := queue.acquire()
	go func() {
		time.Sleep(100 * time.Millisecond)
		w.finish()
	}()

	queue.waitAcquire()
}
