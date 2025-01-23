package executor

import (
	"github.com/KubrickLiu/tempo_async/lock"
	"sync"
)

type WorkerQueue interface {
	len() int
	isEmpty() bool
	newWorker(threshold int, newFunction func() Worker) bool
	detach() Worker
	waitDetach()
	release(Worker)
	clean()
}

type loopQueue struct {
	WorkerQueue

	locker sync.Locker
	cond   *sync.Cond
	idle   []Worker
	using  map[Worker]bool
}

func NewQueue() WorkerQueue {
	queue := &loopQueue{
		locker: lock.NewSpinLock(),
		idle:   make([]Worker, 0),
		using:  make(map[Worker]bool),
	}

	queue.cond = sync.NewCond(queue.locker)
	return queue
}

func (queue *loopQueue) len() int {
	var ret int
	lock.Wrap(queue.locker, func() {
		ret = len(queue.idle) + len(queue.using)
	})
	return ret
}

func (queue *loopQueue) isEmpty() bool {
	return queue.len() == 0
}

func (queue *loopQueue) newWorker(threshold int, newFunction func() Worker) bool {
	flag := false

	lock.Wrap(queue.locker, func() {
		idleLen := len(queue.idle)
		if idleLen >= threshold {
			return
		}

		worker := newFunction()
		if worker == nil {
			return
		}

		queue.idle = append(queue.idle, worker)
		flag = true
	})

	queue.cond.Broadcast()
	return flag
}

func (queue *loopQueue) detach() Worker {
	var worker Worker = nil

	lock.Wrap(queue.locker, func() {
		idleLen := len(queue.idle)
		if idleLen == 0 {
			return
		}

		worker = queue.idle[idleLen-1]
		queue.idle = queue.idle[:idleLen-1]
		queue.using[worker] = true
	})

	return worker
}

func (queue *loopQueue) release(worker Worker) {
	if worker == nil {
		return
	}

	lock.Wrap(queue.locker, func() {
		if f := queue.using[worker]; !f {
			return
		}

		delete(queue.using, worker)
		queue.idle = append(queue.idle, worker)
	})

	queue.cond.Broadcast()
}

func (queue *loopQueue) clean() {
	cleanUsing := func() {
		for worker, _ := range queue.using {
			worker.close()
		}

		queue.using = make(map[Worker]bool)
	}

	cleanIdle := func() {
		for _, worker := range queue.idle {
			worker.close()
		}

		queue.idle = make([]Worker, 0)
	}

	lock.Wrap(queue.locker, func() {
		cleanUsing()
		cleanIdle()
	})

	queue.cond.Broadcast()
}

func (queue *loopQueue) waitDetach() {
	queue.cond.Wait()
}
