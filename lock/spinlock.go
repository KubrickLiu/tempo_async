package lock

import (
	"runtime"
	"sync"
	"sync/atomic"
)

type spinLock uint32

const maxBackoff = 32

func (lock *spinLock) Lock() {
	backoff := 1
	for !atomic.CompareAndSwapUint32((*uint32)(lock), 0, 1) {
		//  see https://en.wikipedia.org/wiki/Exponential_backoff.
		for i := 0; i < backoff; i++ {
			runtime.Gosched()
		}

		if backoff < maxBackoff {
			backoff <<= 1
		}
	}
}

func (lock *spinLock) Unlock() {
	atomic.StoreUint32((*uint32)(lock), 0)
}

func NewSpinLock() sync.Locker {
	return new(spinLock)
}
