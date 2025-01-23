package executor

import (
	"github.com/KubrickLiu/tempo_async/lock"
	"sync"
	"sync/atomic"
)

type Pool interface {
	Apply() (Worker, error)
	Release(Worker)
	RunningNums() int
	WaitingNums() int
}

type poolImpl struct {
	capacity    int
	runningNums atomic.Int32
	waitingNums atomic.Int32
	workerQueue WorkerQueue
	waitingLock sync.Locker
}

func NewPool(capacity int) Pool {
	pool := &poolImpl{
		capacity:    capacity,
		workerQueue: NewQueue(),
		waitingLock: lock.NewSpinLock(),
	}
	return pool
}

func (pool *poolImpl) Apply() (Worker, error) {
retry:
	// idle not empty, detach success
	w := pool.workerQueue.detach()
	if w != nil {
		pool.runningNums.Add(1)
		w.run()
		return w, nil
	}

	// 1. queue 可以创建新的 worker
	if isSuccess := pool.workerQueue.newWorker(pool.capacity, func() Worker {
		w = NewWorker(1, pool)
		return w
	}); isSuccess {
		pool.runningNums.Add(1)
		w.run()
		return w, nil
	}

	// 2. queue 已经满，需要等待
	// 假设是无限等待队列
	pool.waitingNums.Add(1)
	pool.workerQueue.waitDetach()
	pool.waitingNums.Add(-1)

	goto retry
}

func (pool *poolImpl) Release(worker Worker) {
	pool.workerQueue.release(worker)
	pool.runningNums.Add(-1)
}

func (pool *poolImpl) RunningNums() int {
	return int(pool.runningNums.Load())
}

func (pool *poolImpl) WaitingNums() int {
	return int(pool.waitingNums.Load())
}
