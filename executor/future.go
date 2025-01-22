package executor

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type Future[T any] interface {
	// Join blocks until the Future is complete
	Join() (T, error)

	// Get blocks until the Future is complete or context is canceled
	Get(context.Context) (T, error)

	// Complete completes the Future with either result or an error
	complete(any)
}

// futureImpl implements the Future interface
type futureImpl[T any] struct {
	value T
	err   error
	done  chan any

	completeOnce sync.Once
	acceptOnce   sync.Once
}

// Verify futureImpl satisfies the Future interface
var _ Future[any] = (*futureImpl[any])(nil)

// newFuture return New Future
func newFuture[T any]() Future[T] {
	future := &futureImpl[T]{
		done: make(chan any),
	}
	return future
}

func (impl *futureImpl[T]) Join() (T, error) {
	impl.acceptWithContext(nil)
	return impl.value, impl.err
}

func (impl *futureImpl[T]) Get(ctx context.Context) (T, error) {
	impl.acceptWithContext(ctx)
	return impl.value, impl.err
}

func (impl *futureImpl[T]) acceptWithContext(ctx context.Context) {
	impl.acceptOnce.Do(func() {
		if ctx == nil {
			select {
			case result := <-impl.done:
				impl.setResult(result)
				return
			}
		} else {
			select {
			case result := <-impl.done:
				impl.setResult(result)
				return
			case <-ctx.Done():
				fmt.Println("```````")
				impl.setResult(errors.New("context canceled"))
				return
			}
		}
	})
}

func (impl *futureImpl[T]) complete(result any) {
	impl.completeOnce.Do(func() {
		impl.done <- result
	})
}

func (impl *futureImpl[T]) setResult(result any) {
	switch value := result.(type) {
	case error:
		impl.err = value
		return
	case T:
		impl.value = value
	default:
		panic("result type invalid")
	}
}
