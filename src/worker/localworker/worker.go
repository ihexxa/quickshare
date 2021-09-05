package localworker

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/ihexxa/quickshare/src/depidx"
	"github.com/ihexxa/quickshare/src/worker"
)

const (
	MsgTypeKey = "msg-type"
)

type WorkerPool struct {
	on          bool
	queue       chan worker.IMsg
	sleep       int
	workerCount int
	started     int
	mtx         *sync.RWMutex
	deps        *depidx.Deps
}

type Msg struct {
	id      uint64
	headers map[string]string
	body    string
}

func NewMsg(id uint64, headers map[string]string, body string) *Msg {
	return &Msg{
		id:      id,
		headers: headers,
		body:    body,
	}
}

func (m *Msg) ID() uint64 {
	return m.id
}

func (m *Msg) Headers() map[string]string {
	return m.headers
}

func (m *Msg) Body() string {
	return m.body
}

func NewWorkerPool(queueSize, sleep, workerCount int, deps *depidx.Deps) *WorkerPool {
	return &WorkerPool{
		on:          true,
		deps:        deps,
		mtx:         &sync.RWMutex{},
		sleep:       sleep,
		workerCount: workerCount,
		queue:       make(chan worker.IMsg, queueSize),
	}
}

func (wp *WorkerPool) TryPut(task worker.IMsg) error {
	// this close the window that queue can be full after checking
	wp.mtx.Lock()
	defer wp.mtx.Unlock()

	if len(wp.queue) == cap(wp.queue) {
		return worker.ErrFull
	}
	wp.queue <- task
	return nil
}

type Sha1Params struct {
	FilePath string
}

func (wp *WorkerPool) Start() {
	wp.on = true
	for wp.started < wp.workerCount {
		go wp.startWorker()
		wp.started++
	}
}

func (wp *WorkerPool) Stop() {
	wp.on = false
}

func (wp *WorkerPool) startWorker() {
	var err error

	// TODO: make it stateful
	for wp.on {
		func() {
			defer func() {
				if p := recover(); p != nil {
					wp.deps.Log().Errorf("worker panic: %s", p)
				}
			}()

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
		}()

		time.Sleep(time.Duration(wp.sleep) * time.Second)
	}

	wp.started--
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

	sha1Sign := fmt.Sprintf("% x", h.Sum(nil))
	err = wp.deps.FileInfos().SetSha1(filePath, sha1Sign)
	if err != nil {
		return fmt.Errorf("fail to set sha1: %s", err)
	}
	return nil
}
