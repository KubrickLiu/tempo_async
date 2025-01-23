package executor

import (
	"github.com/KubrickLiu/tempo_async/lock"
	"sync"
	"sync/atomic"
	"time"
)

type WorkerQueue interface {
	len() int
	isEmpty() bool
	acceptWorker(threshold int, newFunction func() Worker) bool
	acquire() Worker
	waitAcquire()
	release(Worker)
	runningNums() int
	waitingNums() int
}

type loopQueue struct {
	locker     sync.Locker
	cond       *sync.Cond
	condTicker *time.Ticker

	idle  []Worker
	using map[Worker]bool

	waitNums atomic.Int32
	runNums  atomic.Int32
}

func newQueue() WorkerQueue {
	queue := &loopQueue{
		locker: lock.NewSpinLock(),
		idle:   make([]Worker, 0),
		using:  make(map[Worker]bool),
	}

	queue.cond = sync.NewCond(queue.locker)
	queue.condTicker = time.NewTicker(20 * time.Millisecond)
	queue.periodBroadcast()
	return queue
}

func (queue *loopQueue) periodBroadcast() {
	go func() {
		select {
		case <-queue.condTicker.C:
			if queue.waitingNums() > 0 && len(queue.idle) > 0 {
				queue.broadcast()
			}
		}
	}()
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

func (queue *loopQueue) acceptWorker(threshold int, newFunction func() Worker) bool {
	flag := false

	lock.Wrap(queue.locker, func() {
		idleLen := len(queue.idle)
		if idleLen > threshold {
			return
		}

		worker := newFunction()
		if worker == nil {
			return
		}

		queue.idle = append(queue.idle, worker)
		worker.registerQueue(queue)
		flag = true
	})

	queue.broadcast()
	return flag
}

func (queue *loopQueue) acquire() Worker {
	var worker Worker = nil

	lock.Wrap(queue.locker, func() {
		idleLen := len(queue.idle)
		if idleLen == 0 {
			return
		}

		worker = queue.idle[idleLen-1]
		queue.idle = queue.idle[:idleLen-1]
		queue.using[worker] = true
		queue.runNums.Add(1)
	})

	if worker != nil {
		worker.run()
	}
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
		queue.runNums.Add(-1)
	})

	queue.broadcast()
}

func (queue *loopQueue) broadcast() {
	queue.cond.Broadcast()
}

func (queue *loopQueue) waitAcquire() {
	if len(queue.idle) == 0 {
		queue.waitNums.Add(1)
		queue.cond.Wait()
		queue.waitNums.Add(-1)
	}
}

func (queue *loopQueue) runningNums() int {
	return int(queue.runNums.Load())
}

func (queue *loopQueue) waitingNums() int {
	return int(queue.waitNums.Load())
}
