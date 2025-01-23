package executor

type Pool interface {
	Apply() (Worker, error)
	Release(worker Worker)
	WaitingNums() int
	RunningNums() int
}

type poolImpl struct {
	capacity    int
	workerQueue WorkerQueue
}

func NewPool(capacity int) Pool {
	pool := &poolImpl{
		capacity:    capacity,
		workerQueue: newQueue(),
	}
	return pool
}

func (pool *poolImpl) Apply() (Worker, error) {
retry:
	// idle not empty, acquire success
	w := pool.workerQueue.acquire()
	if w != nil {
		return w, nil
	}

	// 1. queue 可以创建新的 worker，重新 acquire
	if isSuccess := pool.workerQueue.acceptWorker(pool.capacity, func() Worker {
		return NewWorker(1)
	}); isSuccess {
		goto retry
	}

	// 2. queue 已经满，需要等待
	// 假设是无限等待队列
	pool.workerQueue.waitAcquire()

	goto retry
}

func (pool *poolImpl) Release(worker Worker) {
	pool.workerQueue.release(worker)
}

func (pool *poolImpl) WaitingNums() int {
	return pool.workerQueue.waitingNums()
}

func (pool *poolImpl) RunningNums() int {
	return pool.workerQueue.runningNums()
}
