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
			c.runner(data)
		}
	}

}

func (c *_eventWorker[T]) Status() Status {
	return c.status
}

func (c *_eventWorker[T]) Run(ctx context.Context) {
	if c.status == Running {
		return
	}

	c.setStatus(Running)
	go c.run(ctx)
}

func (c *_eventWorker[T]) Pause() error {
	if c.status != Running {
		return ERR_NOT_RUNNING
	}

	c.setStatus(Paused)

	return nil
}

func (c *_eventWorker[T]) Resume() error {
	if c.status != Paused {
		return ERR_NOT_PAUSED
	}

	c.setStatus(Running)

	return nil
}

func (c *_eventWorker[T]) Stop() error {
	if c.status != Running {
		return ERR_NOT_RUNNING
	}

	c.setStatus(Stopped)
	c.stop <- struct{}{}

	return nil
}

func (c *_eventWorker[T]) Invoke(data T) error {
	if c.status != Running {
		return ERR_NOT_RUNNING
	}

	c.mu.Lock()
	c.receiver <- data
	c.mu.Unlock()

	return nil
}
