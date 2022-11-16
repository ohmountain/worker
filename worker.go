package worker

import (
	"context"
	"errors"
)

type Status uint8

const (
	Stopped Status = iota
	Running
	Paused
)

var ERR_IS_RUNNING = errors.New("worker is running")
var ERR_NOT_RUNNING = errors.New("worker is not running")
var ERR_IS_PAUSED = errors.New("worker is paused")
var ERR_NOT_PAUSED = errors.New("worker is not paused")

// Worker
// 定义一个worker的规范
type Worker interface {

	// Status
	// 返回它的当前状态
	Status() Status

	// Run
	// 它可以通过总线运行
	Run(ctx context.Context)

	// Pause
	// 它可以通过总线进行暂停
	Pause() error

	// Resume
	// 它可以通过总线恢复运行
	Resume() error

	// Stop
	// 它可以通过总线停止
	Stop() error
}

type TickerWorker interface {
	Worker
}

// EventWorker
// 事件性地处理任务的worker
// @type T 为处理任务数据的类型
type EventWorker[T any] interface {
	Worker

	Invoke(data T) error
}

type PubSubWorker[T any] interface {
	Worker

	// Pub
	// 推送订阅数据
	Pub(data T) error

	// Sub
	// 订阅数据
	Sub(fn RunnerFuncT[T]) string

	// UnSub
	// 停止订阅
	UnSub(sub string)
}

type RunnerFunc func()
type RunnerFuncT[T any] func(data T)
