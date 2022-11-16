package worker

import (
	"context"
	"runtime"
	"sync"
)

type _pubsubWorker[T any] struct {
	mu sync.Mutex

	stop   chan struct{}
	status Status

	// cur
	// 订阅计数器
	cur int

	// subs
	// 订阅者和订阅函数
	subs map[int]RunnerFuncT[T]

	// 允许并行运行的订阅函数
	paralles chan struct{}
}

// NewPubsubWorker
// 创建一个事件/数据驱动的Worker
// 在创建这个Worker时必须指定Worker接受的数据类型，即Worker可以处理那种数据类型
func NewPubSubWorker[T any]() *_pubsubWorker[T] {
	return &_pubsubWorker[T]{
		mu:       sync.Mutex{},
		status:   Stopped,
		stop:     make(chan struct{}, 1),
		cur:      0,
		subs:     make(map[int]RunnerFuncT[T]),
		paralles: make(chan struct{}, runtime.NumCPU()*2),
	}
}

func (c *_pubsubWorker[T]) setStatus(status Status) {
	if c.status != status {
		c.mu.Lock()
		c.status = status
		c.mu.Unlock()
	}
}

func (c *_pubsubWorker[T]) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			c.Stop()
		case <-c.stop:
			return
		}
	}

}

// pub
// 执行函数
// 这个执行函数通过通道限制同时执行的数量
func (c *_pubsubWorker[T]) pub(fn RunnerFuncT[T], data T, wg *sync.WaitGroup) {

	// recover 防止因为某个订阅函数导致所有订阅函数无法执行
	defer func() {
		recover()
		<-c.paralles
		wg.Done()
	}()

	c.paralles <- struct{}{}
	fn(data)
}

// Status
// 返回当前Worker的状态
func (c *_pubsubWorker[T]) Status() Status {
	return c.status
}

// Run
// 运行当前Worker
// 注意：这是异步的，
// 如果要停止这个Worker，可以使用ctx来停止，也可注意主动调用Stop来停止
func (c *_pubsubWorker[T]) Run(ctx context.Context) {
	if c.status == Running {
		return
	}

	c.setStatus(Running)
	go c.run(ctx)
}

// Pause
// 暂停当前Workekr
func (c *_pubsubWorker[T]) Pause() error {
	if c.status != Running {
		return ERR_NOT_RUNNING
	}

	c.setStatus(Paused)

	return nil
}

// Resume
// 从暂停状态中恢复运行
func (c *_pubsubWorker[T]) Resume() error {
	if c.status != Paused {
		return ERR_NOT_PAUSED
	}

	c.setStatus(Running)

	return nil
}

// Stop
// 停止掉Worker
func (c *_pubsubWorker[T]) Stop() error {
	if c.status != Running {
		return ERR_NOT_RUNNING
	}

	c.setStatus(Stopped)
	c.stop <- struct{}{}

	return nil
}

// Pub
// 发布消息
func (c *_pubsubWorker[T]) Pub(data T) error {
	if c.status != Running {
		return ERR_NOT_RUNNING
	}

	var wg = &sync.WaitGroup{}
	for _, fn := range c.subs {
		wg.Add(1)
		go c.pub(fn, data, wg)
	}
	wg.Wait()

	return nil
}

// Sub
// 订阅这个Worker的数据
func (c *_pubsubWorker[T]) Sub(fn RunnerFuncT[T]) int {
	c.mu.Lock()
	c.cur++
	c.subs[c.cur] = fn
	c.mu.Unlock()

	return c.cur
}

// 取消订阅这个Worker的数据
func (c *_pubsubWorker[T]) UnSub(sub int) {
	c.mu.Lock()
	delete(c.subs, sub)
	c.mu.Unlock()
}
