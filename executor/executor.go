package executor

import "errors"

var (
	defaultPool = NewPool(10)
)

func Submit[T any](function func() (T, error)) Future[T] {
	return SubmitWithPool[T](defaultPool, function)
}

func SubmitWithPool[T any](pool Pool, function func() (T, error)) Future[T] {
	var promise Promise[T]

	w, err := pool.Apply()
	if err != nil {
		promise = NewPromise[T]()
		promise.Failure(err)
		return promise.Future()
	}

	if w == nil {
		promise = NewPromise[T]()
		promise.Failure(errors.New(""))
		return promise.Future()
	}

	task := NewTask[T](function)
	w.append(task)
	return task.Future()
}
