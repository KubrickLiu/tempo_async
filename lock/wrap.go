package lock

import "sync"

func Wrap(locker sync.Locker, fn func()) {
	locker.Lock()
	defer func() {
		locker.Unlock()
		// TODO recover
	}()

	fn()
}
