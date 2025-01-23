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
	append(ITask)
	finish()
	close()
	registerQueue(WorkerQueue)
}

type workerImpl struct {
	tasksChannel chan ITask
	closeOnce    sync.Once
	open         atomic.Bool
	workQueue    WorkerQueue
}

func NewWorker(bufferSize int) Worker {
	worker := &workerImpl{
		tasksChannel: make(chan ITask, bufferSize),
	}
	return worker
}

func (worker *workerImpl) isOpen() bool {
	return worker.open.Load()
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

				// channel accept finish signal
				if val == nil {
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
		if worker.workQueue != nil {
			worker.workQueue.release(worker)
		}
	}

	go drain()
}

func (worker *workerImpl) append(task ITask) {
	if task != nil && worker.isOpen() {
		worker.tasksChannel <- task
		worker.run()
	}
}

func (worker *workerImpl) finish() {
	worker.tasksChannel <- nil
}

func (worker *workerImpl) close() {
	worker.closeOnce.Do(func() {
		close(worker.tasksChannel)
		worker.tasksChannel = nil
	})
}

func (worker *workerImpl) registerQueue(queue WorkerQueue) {
	worker.workQueue = queue
}
