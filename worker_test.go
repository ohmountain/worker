package worker

import (
	"context"
	"testing"
	"time"
)

func TestEventWorker(t *testing.T) {
	var wk EventWorker[int64] = NewEventWorker(func(data int64) {
		t.Logf("Worker processed: %#v", time.Unix(data, 0).Format(time.RFC1123))
	})

	ctx, cf := context.WithTimeout(context.Background(), time.Minute*1)
	defer cf()

	wk.Run(ctx)

	go func() {
		ticker := time.NewTicker(time.Second * 1)
		timer := time.NewTimer(time.Second * 60)

		for {
			select {
			case <-ticker.C:
				wk.Invoke(time.Now().Unix())
			case <-timer.C:
				ticker.Stop()
				timer.Stop()
				return
			}
		}
	}()

	time.Sleep(62 * time.Second)
}

func TestTickerWorker(t *testing.T) {
	now := time.Now().Unix()
	ctx, cf := context.WithCancel(context.Background())
	var wk TickerWorker
	wk = NewTickerWorker(*time.NewTicker(1 * time.Second), func() {
		duration := time.Now().Unix() - now
		switch duration {
		case 3:
			wk.Pause()
			timer := time.NewTimer(2 * time.Second)
			<-timer.C
			wk.Resume()
		case 10:
			cf()
		default:
			t.Logf("D: %d", duration)
		}
	})

	wk.Run(ctx)
	<-ctx.Done()

	t.Logf("Done all")
}

func TestPubSubWorker(t *testing.T) {
	wk := NewPubSubWorker[string]()
	ctx, cf := context.WithCancel(context.Background())

	wk.Run(ctx)

	for i := 0; i < 10; i++ {
		func(x int) {
			wk.Sub(func(data string) {
				t.Logf("%d - Receive: %s", x, data)
			})
		}(i)
	}

	wk1 := NewTickerWorker(*time.NewTicker(2 * time.Second), func() { wk.Pub(time.Now().String()) })
	wk1.Run(ctx)

	time.Sleep(100 * time.Second)
	cf()
}
