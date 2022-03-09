package worker

import "errors"

var (
	ErrFull   = errors.New("worker queue is full, make it larger in the config")
	ErrClosed = errors.New("async handlers are closed")
)

func IsErrFull(err error) bool {
	return err == ErrFull
}

type IMsg interface {
	ID() uint64
	Headers() map[string]string
	Body() string
}

type MsgHandler = func(msg IMsg) error

type IWorkerPool interface {
	TryPut(task IMsg) error
	Start()
	Stop()
	AddHandler(msgType string, handler MsgHandler)
	DelHandler(msgType string)
}
