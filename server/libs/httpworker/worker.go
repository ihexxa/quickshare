package httpworker

import (
	"errors"
	"net/http"
	"runtime/debug"
	"time"
)

import (
	"github.com/ihexxa/quickshare/server/libs/logutil"
)

var (
	ErrWorkerNotFound = errors.New("worker not found")
	ErrTimeout        = errors.New("timeout")
)

type DoFunc func(http.ResponseWriter, *http.Request)

type Task struct {
	Ack chan error
	Do  DoFunc
	Res http.ResponseWriter
	Req *http.Request
}

type Workers interface {
	Put(*Task) bool
	IsInTime(ack chan error, msec time.Duration) error
}

type WorkerPool struct {
	queue   chan *Task
	size    int
	workers []*Worker
	log     logutil.LogUtil // TODO: should not pass log here
}

func NewWorkerPool(poolSize int, queueSize int, log logutil.LogUtil) Workers {
	queue := make(chan *Task, queueSize)
	workers := make([]*Worker, 0, poolSize)

	for i := 0; i < poolSize; i++ {
		worker := &Worker{
			Id:    uint64(i),
			queue: queue,
			log:   log,
		}

		go worker.Start()
		workers = append(workers, worker)
	}

	return &WorkerPool{
		queue:   queue,
		size:    poolSize,
		workers: workers,
		log:     log,
	}
}

func (pool *WorkerPool) Put(task *Task) bool {
	if len(pool.queue) >= pool.size {
		return false
	}

	pool.queue <- task
	return true
}

func (pool *WorkerPool) IsInTime(ack chan error, msec time.Duration) error {
	start := time.Now().UnixNano()
	timeout := make(chan error)

	go func() {
		time.Sleep(msec)
		timeout <- ErrTimeout
	}()

	select {
	case err := <-ack:
		if err == nil {
			pool.log.Printf(
				"finish cost: %d usec",
				(time.Now().UnixNano()-start)/1000,
			)
		} else {
			pool.log.Printf(
				"finish with error cost: %d usec",
				(time.Now().UnixNano()-start)/1000,
			)
		}
		return err
	case errTimeout := <-timeout:
		pool.log.Printf("timeout cost: %d usec", (time.Now().UnixNano()-start)/1000)
		return errTimeout
	}
}

type Worker struct {
	Id    uint64
	queue chan *Task
	log   logutil.LogUtil
}

func (worker *Worker) RecoverPanic() {
	if r := recover(); r != nil {
		worker.log.Printf("Recovered:%v stack: %v", r, debug.Stack())
		// restart worker and IsInTime will return timeout error for last task
		worker.Start()
	}
}

func (worker *Worker) Start() {
	defer worker.RecoverPanic()

	for {
		task := <-worker.queue
		if task.Do != nil {
			task.Do(task.Res, task.Req)
			task.Ack <- nil
		} else {
			task.Ack <- ErrWorkerNotFound
		}
	}
}

// ServiceFunc lets you return struct directly
type ServiceFunc func(http.ResponseWriter, *http.Request) interface{}
