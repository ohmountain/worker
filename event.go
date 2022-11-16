package worker

import (
	"context"
	"sync"
)

// _eventWorker
// 用于基于事件任务的worker
type _eventWorker[T any] struct {
	mu sync.Mutex

	stop   chan struct{}
	status Status

	receiver chan T
	runner   RunnerFuncT[T]
}

// NewEventWorker
// 创建一个事件/数据驱动的Worker
// @pparam runner RunnerFunc[T]，
// 在创建这个Worker时必须指定Worker接受的数据类型，即Worker可以处理那种数据类型
func NewEventWorker[T any](runner RunnerFuncT[T]) *_eventWorker[T] {
	return &_eventWorker[T]{
		mu:       sync.Mutex{},
		status:   Stopped,
		stop:     make(chan struct{}, 1),
		receiver: make(chan T),
		runner:   runner,
	}
}

func (c *_eventWorker[T]) setStatus(status Status) {
	if c.status != status {
		c.mu.Lock()
		c.status = status
		c.mu.Unlock()
	}
}

func (c *_eventWorker[T]) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			c.Stop()
		case <-c.stop:
			return
		case data := <-c.receiver:
			func() {
				// recover if panic
				defer func() {
					recover()
				}()
				c.runner(data)
			}()
		}
	}
}

// Status
// 返回当前Worker的状态
func (c *_eventWorker[T]) Status() Status {
	return c.status
}

// Run
// 运行当前Worker
// 注意：这是异步的，
// 如果要停止这个Worker，可以使用ctx来停止，也可注意主动调用Stop来停止
func (c *_eventWorker[T]) Run(ctx context.Context) {
	if c.status == Running {
		return
	}

	c.setStatus(Running)
	go c.run(ctx)
}

// Pause
// 暂停当前Workekr
func (c *_eventWorker[T]) Pause() error {
	if c.status != Running {
		return ERR_NOT_RUNNING
	}

	c.setStatus(Paused)

	return nil
}

// Resume
// 从暂停状态中恢复运行
func (c *_eventWorker[T]) Resume() error {
	if c.status != Paused {
		return ERR_NOT_PAUSED
	}

	c.setStatus(Running)

	return nil
}

// Resume
// 停止掉Worker
func (c *_eventWorker[T]) Stop() error {
	if c.status != Running {
		return ERR_NOT_RUNNING
	}

	c.setStatus(Stopped)
	c.stop <- struct{}{}

	return nil
}

// Invoke
// 向Worker的处理函数发送数据
func (c *_eventWorker[T]) Invoke(data T) error {
	if c.status != Running {
		return ERR_NOT_RUNNING
	}

	c.mu.Lock()
	c.receiver <- data
	c.mu.Unlock()

	return nil
}
