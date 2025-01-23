package executor

type callFunc[T any] func() (T, error)

type ITask interface {
	Call()
}

type Task[T any] interface {
	ITask

	Future() Future[T]
}

type taskImpl[T any] struct {
	function callFunc[T]
	promise  Promise[T]
}

var _ Task[any] = (*taskImpl[any])(nil)

func NewTask[T any](function callFunc[T]) Task[T] {
	task := &taskImpl[T]{
		function: function,
		promise:  NewPromise[T](),
	}
	return task
}

func (task *taskImpl[T]) Call() {
	if task.function == nil {
		panic("invalid task function is nil")
	}

	ret, err := task.function()
	if err != nil {
		task.promise.Failure(err)
	} else {
		task.promise.Success(ret)
	}
}

func (task *taskImpl[T]) Future() Future[T] {
	return task.promise.Future()
}
