package localworker

import (
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/ihexxa/quickshare/src/worker"
)

const (
	MsgTypeKey = "msg-type"
)

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

type WorkerPool struct {
	on          bool
	queue       chan worker.IMsg
	sleep       int
	workerCount int
	started     int
	mtx         *sync.RWMutex
	logger      *zap.SugaredLogger
	msgHandlers map[string]worker.MsgHandler
}

func NewWorkerPool(queueSize, sleep, workerCount int, logger *zap.SugaredLogger) *WorkerPool {
	return &WorkerPool{
		on:          true,
		logger:      logger,
		mtx:         &sync.RWMutex{},
		sleep:       sleep,
		workerCount: workerCount,
		queue:       make(chan worker.IMsg, queueSize),
		msgHandlers: map[string]worker.MsgHandler{},
	}
}

func (wp *WorkerPool) TryPut(task worker.IMsg) error {
	// this closes the window that queue can be full after checking
	wp.mtx.Lock()
	defer wp.mtx.Unlock()

	if len(wp.queue) == cap(wp.queue) {
		return worker.ErrFull
	}
	wp.queue <- task
	return nil
}

func (wp *WorkerPool) Start() {
	wp.mtx.Lock()
	defer wp.mtx.Unlock()

	wp.on = true
	for wp.started < wp.workerCount {
		go wp.startWorker()
		wp.started++
	}
}

func (wp *WorkerPool) Stop() {
	wp.mtx.Lock()
	defer wp.mtx.Unlock()

	// TODO: avoid sending and panic
	close(wp.queue)
	wp.on = false
	for wp.started > 0 {
		wp.logger.Errorf(
			fmt.Sprintf(
				"%d workers (sleep %d second) still in working/sleeping",
				wp.sleep,
				wp.started,
			),
		)
		time.Sleep(time.Duration(1) * time.Second)
	}
}

func (wp *WorkerPool) startWorker() {
	var err error

	// TODO: make it stateful
	for wp.on {
		func() {
			defer func() {
				if p := recover(); p != nil {
					wp.logger.Errorf("worker panic: %s", p)
				}
			}()


			msg, ok := <-wp.queue
			if !ok {
				return
			}
			
			headers := msg.Headers()
			msgType, ok := headers[MsgTypeKey]
			if !ok {
				wp.logger.Errorf("msg type not found: %v", headers)
				return
			}

			handler, ok := wp.msgHandlers[msgType]
			if !ok {
				wp.logger.Errorf("no handler for the message type: %s", msgType)
				return
			}

			if err = handler(msg); err != nil {
				wp.logger.Errorf("async task(%s) failed: %s", msgType, err)
			}
		}()

		time.Sleep(time.Duration(wp.sleep) * time.Second)
	}

	wp.started--
}

func (wp *WorkerPool) AddHandler(msgType string, handler worker.MsgHandler) {
	// existing task type will be overwritten
	wp.msgHandlers[msgType] = handler
}

func (wp *WorkerPool) DelHandler(msgType string) {
	delete(wp.msgHandlers, msgType)
}
