package queue

import (
	"sync"

	"message-queue/message"
)

// Queue stores messages in FIFO order.
type Queue struct {
	mu       sync.Mutex
	messages []message.Message
}

func New() *Queue {
	return &Queue{}
}

func (q *Queue) Enqueue(item message.Message) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.messages = append(q.messages, item)
}

func (q *Queue) Dequeue() (message.Message, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.messages) == 0 {
		return message.Message{}, false
	}

	item := q.messages[0]
	q.messages[0] = message.Message{}
	q.messages = q.messages[1:]
	return item, true
}

func (q *Queue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.messages)
}
