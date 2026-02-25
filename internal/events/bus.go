package events

import (
	"context"
	"sync"
	"time"
)

type Event struct {
	Type      string
	Payload   any
	CreatedAt time.Time
}

type Consumer interface {
	Handle(ctx context.Context, e Event) error
}

type Bus struct {
	ch   chan Event
	stop chan struct{}
	wg   sync.WaitGroup
}

func NewBus(buffer int) *Bus {
	if buffer <= 0 {
		buffer = 64
	}
	return &Bus{
		ch:   make(chan Event, buffer),
		stop: make(chan struct{}),
	}
}

func (b *Bus) Publish(ctx context.Context, e Event) {
	select {
	case b.ch <- e:
	default:
		// buffer full: drop event in this toy simulation
	}
}

func (b *Bus) StartWorker(ctx context.Context, c Consumer) {
	b.wg.Add(1)
	go func() {
		defer b.wg.Done()
		for {
			select {
			case <-b.stop:
				return
			case e := <-b.ch:
				_ = c.Handle(ctx, e)
			}
		}
	}()
}

func (b *Bus) Stop() {
	close(b.stop)
	b.wg.Wait()
}