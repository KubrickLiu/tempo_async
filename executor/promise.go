package executor

import (
	"sync"
)

type Promise[T any] interface {
	// Success completes the underlying Future with a value
	Success(T)

	// Failure fails the underlying Future with an error
	Failure(error)

	// Future return the underlying Future
	Future() Future[T]
}

// promiseImpl implements the Promise interface
type promiseImpl[T any] struct {
	future Future[T]
	once   sync.Once
}

// Verify promiseImpl satisfies the Promise interface
var _ Promise[any] = (*promiseImpl[any])(nil)

// NewPromise return new Promise
func NewPromise[T any]() Promise[T] {
	promise := &promiseImpl[T]{
		future: newFuture[T](),
	}
	return promise
}

func (impl *promiseImpl[T]) Success(result T) {
	impl.once.Do(func() {
		impl.future.complete(result)
	})
}

func (impl *promiseImpl[T]) Failure(err error) {
	impl.once.Do(func() {
		impl.future.complete(err)
	})
}

func (impl *promiseImpl[T]) Future() Future[T] {
	return impl.future
}
