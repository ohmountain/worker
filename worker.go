// 这个文件定义给了一些常用worker的定义
// 这些worker包括:
// 1. Worker
//    Worker是一个基本的规范和约束,Worker与父协程使用context通讯,允许父协程通过context停止它.
//    Worker做了一些约束:
//    它们是:‘状态’(Status())、运行(Run(ctx context.Context))、暂停(Pause())、恢复(Resume())、停止(Stop()).
//    需要注意的是,通过Run(ctx)中的ctx和Stop()都可以停止掉Worker,区别在于,如果父协程管理多个Worker,那么可以统一通过context停止掉所有的Worker;而主动调用Stop()只能停止掉当前的Worker.
//
// 2. TickerWorker
//    TickerWorker实现Worker的基本约束,创建它时需要一个计时器和执行函数,它重复会在计时器的到期时触发这个执行函数,直到被停止.
//    TickerWorker被调用Pause后,计时器不会重制,但是计时器到期时不会触发执行函数.
//
// 3. EventWorker
//    EventWorker是一个事件/数据驱动的Worker,它实现Worker的所有约束,同时新增了Invoke约束.
//    一般来说,一个EventWorker只能处理一种数据类型,除非指定处理的数据类型时any,同时它只允许拥有一个处理函数.
//    当父协程调用Invoke时,会触发执行函数.
//    如果想要将原始数据分发给其他函数,可以在执行函数中进行分发;这种情况,最好使用PubSubWorker.
// 4. PubSubWorker
//    PubSubWorker是一个发布订阅Worker，它实现Worker的所有约束，同时添加了Pub/Sub/UnSub三个约束.
//    PubSubWorker类似EventWorker，PubSubWorker通过Sun函数可以拥有多个执行函数，同时可以在程序生命周期内退出订阅.

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
	// 它可以通过父协程运行
	Run(ctx context.Context)

	// Pause
	// 它可以通过父协程进行暂停
	Pause() error

	// Resume
	// 它可以通过父协程恢复运行
	Resume() error

	// Stop
	// 它可以通过父协程停止
	Stop() error
}

// TickerWorker
// 计时器worker
type TickerWorker interface {
	Worker
}

// EventWorker
// 事件性地处理任务的worker
// @type T 为处理任务数据的类型
// 一个事件只能拥有一个处理函数
type EventWorker[T any] interface {
	Worker

	// Invoke
	Invoke(data T) error
}

// PubSubWorker
// 发布订阅的worker
// 一个发布,多个订阅
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

// RunnerFunc
// 无入参的执行函数，通常作为TickerWorker的执行函数
type RunnerFunc func()

// RunnerFuncT
// 有入参的执行函数，在创建Worker时必须指定参数类型
type RunnerFuncT[T any] func(data T)
