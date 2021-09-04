package worker

import (
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/ihexxa/quickshare/src/depidx"
)

const (
	MsgTypeKey = "msg-type"
)

type Msg interface {
	ID() uint64
	Headers() map[string]string
	Body() string
}

var ErrFull = errors.New("worker queue is full, make it larger in the config.")

func IsErrFull(err error) bool {
	return err == ErrFull
}

type WorkerPool struct {
	on    bool
	queue chan Msg
	mtx   *sync.RWMutex
	deps  *depidx.Deps
}

func NewWorkerPool(queueSize int, deps *depidx.Deps) *WorkerPool {
	return &WorkerPool{
		on:    true,
		deps:  deps,
		mtx:   &sync.RWMutex{},
		queue: make(chan Msg, queueSize),
	}
}

func (wp *WorkerPool) TryPut(task Msg) error {
	// this close the window that queue can be full after checking
	wp.mtx.Lock()
	defer wp.mtx.Unlock()

	if len(wp.queue) == cap(wp.queue) {
		return ErrFull
	}
	wp.queue <- task
	return nil
}

type Sha1Params struct {
	FilePath string
}

func (wp *WorkerPool) startWorker() {
	var err error

	for wp.on {
		msg := <-wp.queue
		headers := msg.Headers()
		msgType, ok := headers[MsgTypeKey]
		if !ok {
			wp.deps.Log().Errorf("msg type not found: %v", headers)
		}

		switch msgType {
		case "sha1":
			sha1Params := &Sha1Params{}
			err = json.Unmarshal([]byte(msg.Body()), sha1Params)
			if err != nil {
				wp.deps.Log().Errorf("fail to unmarshal sha1 msg: %s", err)
			}
			err = wp.sha1Task(sha1Params.FilePath)
			if err != nil {
				wp.deps.Log().Errorf("fail to do sha1: %s", err)
			}
		default:
			wp.deps.Log().Errorf("unknown message tyope: %s", msgType)
		}
	}
}

func (wp *WorkerPool) sha1Task(filePath string) error {
	f, err := wp.deps.FS().GetFileReader(filePath)
	if err != nil {
		return fmt.Errorf("fail to get reader: %s", err)
	}

	h := sha1.New()
	buf := make([]byte, 4096)
	_, err = io.CopyBuffer(h, f, buf)
	if err != nil {
		return err
	}

	// sha1Sign := fmt.Sprintf("% x", h.Sum(nil))
	// save it to db
	return nil
}
