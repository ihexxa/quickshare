package worker

import "errors"

var ErrFull = errors.New("worker queue is full, make it larger in the config.")

func IsErrFull(err error) bool {
	return err == ErrFull
}

type IMsg interface {
	ID() uint64
	Headers() map[string]string
	Body() string
}

type IWorkerPool interface {
	TryPut(task IMsg) error
	Start()
	Stop()
}
