package executor

type callFunc[T any] func() (T, error)

type Task[T any] interface {
	// Call starts executing task using a goroutine.
	Call() Future[T]
}

type taskImpl[T any] struct {
	promise  Promise[T]
	function callFunc[T]
}

var _ Task[any] = (*taskImpl[any])(nil)

func NewTask[T any](function callFunc[T]) Task[T] {
	task := &taskImpl[T]{
		function: function,
		promise:  NewPromise[T](),
	}
	return task
}

func (task *taskImpl[T]) Call() Future[T] {
	go func() {
		result, err := task.function()
		if err != nil {
			task.promise.Failure(err)
		} else {
			task.promise.Success(result)
		}
	}()
	return task.promise.Future()
}
