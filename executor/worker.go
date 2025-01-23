package executor

import (
	"sync"
	"sync/atomic"
	"time"
)

const (
	maxIdleCounter = 6000 // 1 min
	waitMillisStep = 10   // 10 ms
)

type Worker interface {
	isOpen() bool
	run()
	append(task ITask)
	finish()
	close()
}

type workerImpl struct {
	tasksChannel chan ITask
	closeOnce    sync.Once
	open         atomic.Bool
	pool         Pool
}

func NewWorker(len int, pool Pool) Worker {
	worker := &workerImpl{
		tasksChannel: make(chan ITask, len),
		pool:         pool,
	}
	return worker
}

func (worker *workerImpl) isOpen() bool {
	flag := worker.open.Load()
	return flag
}

func (worker *workerImpl) run() {
	if !worker.open.CompareAndSwap(false, true) {
		return
	}

	drain := func() {
		counter := 0

		for {
			if !worker.isOpen() {
				break
			}

			select {
			case val, ok := <-worker.tasksChannel:
				// channel has closed
				if !ok {
					worker.open.Store(false)
					break
				}

				switch val.(type) {
				case ITask:
					val.Call()
				default:
					time.Sleep(waitMillisStep * time.Millisecond)
					counter += 1

					if counter > maxIdleCounter {
						// worker finish and return
						worker.open.Store(false)
						break
					}
				}
			}
		}

		// channel is empty and pool is not nil, release current worker from pool
		if worker.pool != nil {
			worker.pool.Release(worker)
		}
	}

	go drain()
}

func (worker *workerImpl) append(task ITask) {
	if task != nil && worker.isOpen() {
		worker.tasksChannel <- task
	}
}

func (worker *workerImpl) finish() {
	worker.tasksChannel <- nil
}

func (worker *workerImpl) close() {
	worker.closeOnce.Do(func() {
		worker.open.Store(false)

		close(worker.tasksChannel)
		worker.tasksChannel = nil
	})
}
