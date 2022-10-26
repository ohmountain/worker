package worker

import (
	"context"
	"log"
	"sync"
	"time"
)

// TickerWorker
// 创建一个定时器的worker
// @param ticker time.Ticker, 定时器
// @param runner RunnerFunc， 定时运行的函数
func NewTickerWorker(ticker time.Ticker, runner RunnerFunc) *_tickerWorker {
	return &_tickerWorker{
		mu:     sync.Mutex{},
		status: Stopped,
		stop:   make(chan struct{}, 1),
		runner: runner,
		ticker: ticker,
	}
}

// _tickerWorker
// 用于定时器任务的worker
type _tickerWorker struct {
	mu sync.Mutex

	stop   chan struct{}
	status Status
	runner RunnerFunc

	ticker time.Ticker
}

func (c *_tickerWorker) setStatus(status Status) {
	if c.status != status {
		c.mu.Lock()
		c.status = status
		c.mu.Unlock()
	}
}

func (c *_tickerWorker) run(ctx context.Context) {
	for {
		select {

		case <-ctx.Done():
			{
				log.Printf("stop from ctx")
				c.ticker.Stop()
				c.Stop()
			}

		case <-c.stop:
			log.Printf("stop from chan")
			return

		case <-c.ticker.C:
			{
				if c.status != Running {
					continue
				}
				go c.runner()
			}
		}
	}
}

func (c *_tickerWorker) Status() Status {
	return c.status
}

func (c *_tickerWorker) Run(ctx context.Context) {
	if c.status == Running {
		return
	}

	c.setStatus(Running)
	go c.run(ctx)
}

func (c *_tickerWorker) Pause() error {
	if c.status != Running {
		return ERR_NOT_RUNNING
	}

	c.setStatus(Paused)

	return nil
}

func (c *_tickerWorker) Resume() error {
	if c.status != Paused {
		return ERR_NOT_PAUSED
	}

	c.setStatus(Running)

	return nil
}

func (c *_tickerWorker) Stop() error {
	if c.status != Running {
		return ERR_NOT_RUNNING
	}

	c.setStatus(Stopped)
	log.Printf("Send stop chan")
	c.stop <- struct{}{}
	log.Printf("Send stop chan done")

	return nil
}
