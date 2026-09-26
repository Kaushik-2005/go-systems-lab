package broker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"message-queue/message"
	"message-queue/queue"
)

// Delivery is a message handed to a consumer but not acknowledged yet.
type Delivery struct {
	Message message.Message
}

type inFlightMessage struct {
	message  message.Message
	deadline time.Time
}

// Broker coordinates producers, consumers, and acknowledgements.
type Broker struct {
	mu                sync.Mutex
	nextID            int
	pending           *queue.Queue
	dead              *queue.Queue
	inflight          map[string]inFlightMessage
	visibilityTimeout time.Duration
	retryLimit        int
}

func New(visibilityTimeout time.Duration, retryLimit int) *Broker {
	return &Broker{
		pending:           queue.New(),
		dead:              queue.New(),
		inflight:          make(map[string]inFlightMessage),
		visibilityTimeout: visibilityTimeout,
		retryLimit:        retryLimit,
	}
}

func (b *Broker) Publish(body string) message.Message {
	b.mu.Lock()
	b.nextID++
	item := message.Message{
		ID:   fmt.Sprintf("message-%d", b.nextID),
		Body: body,
	}
	b.mu.Unlock()

	b.pending.Enqueue(item)
	return item
}

func (b *Broker) Consume() (Delivery, bool) {
	item, ok := b.pending.Dequeue()
	if !ok {
		return Delivery{}, false
	}
	item.Attempts++

	b.mu.Lock()
	b.inflight[item.ID] = inFlightMessage{
		message:  item,
		deadline: time.Now().Add(b.visibilityTimeout),
	}
	b.mu.Unlock()
	return Delivery{Message: item}, true
}

func (b *Broker) Ack(id string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, ok := b.inflight[id]; !ok {
		return false
	}
	delete(b.inflight, id)
	return true
}

func (b *Broker) StartVisibilityMonitor(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		return
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				b.RequeueExpired()
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (b *Broker) RequeueExpired() int {
	now := time.Now()
	expired := make([]message.Message, 0)

	b.mu.Lock()
	for id, delivery := range b.inflight {
		if !now.Before(delivery.deadline) {
			expired = append(expired, delivery.message)
			delete(b.inflight, id)
		}
	}
	b.mu.Unlock()

	for _, item := range expired {
		if item.Attempts >= b.retryLimit {
			b.dead.Enqueue(item)
		} else {
			b.pending.Enqueue(item)
		}
	}
	return len(expired)
}

func (b *Broker) PendingLen() int {
	return b.pending.Len()
}

func (b *Broker) InFlightLen() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.inflight)
}

func (b *Broker) DeadLetterLen() int {
	return b.dead.Len()
}

func (b *Broker) ConsumeDeadLetter() (message.Message, bool) {
	return b.dead.Dequeue()
}
