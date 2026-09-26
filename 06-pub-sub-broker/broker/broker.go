package broker

import "sync"

// Event is one publication delivered to subscribers.
type Event struct {
	ID   string
	Body string
}

// Subscription is one subscriber's private stream of messages.
type Subscription struct {
	id       int
	topic    string
	messages chan Event

	mu     sync.Mutex
	closed bool
}

// Messages returns the channel owned by this subscription.
func (s *Subscription) Messages() <-chan Event {
	return s.messages
}

// Broker routes each published message to every subscriber of its topic.
type Broker struct {
	mu     sync.RWMutex
	nextID int
	topics map[string]map[int]*Subscription
}

func New() *Broker {
	return &Broker{topics: make(map[string]map[int]*Subscription)}
}

// Subscribe creates a buffered subscription for a topic.
func (b *Broker) Subscribe(topic string, buffer int) *Subscription {
	if buffer < 1 {
		buffer = 1
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.nextID++
	subscription := &Subscription{
		id:       b.nextID,
		topic:    topic,
		messages: make(chan Event, buffer),
	}

	if b.topics[topic] == nil {
		b.topics[topic] = make(map[int]*Subscription)
	}
	b.topics[topic][subscription.id] = subscription

	return subscription
}

// Publish sends the event to every current subscriber of topic.
// It blocks if a subscriber's buffer is full; this makes backpressure visible.
func (b *Broker) Publish(topic string, event Event) int {
	subscribers := b.subscribers(topic)

	delivered := 0
	for _, subscription := range subscribers {
		if subscription.deliver(event) {
			delivered++
		}
	}

	return delivered
}

// PublishNonBlocking sends to subscribers whose buffers have room.
// A full subscriber is skipped so one slow subscriber cannot block the publisher.
func (b *Broker) PublishNonBlocking(topic string, event Event) (delivered, dropped int) {
	for _, subscription := range b.subscribers(topic) {
		if subscription.tryDeliver(event) {
			delivered++
		} else {
			dropped++
		}
	}
	return delivered, dropped
}

// Unsubscribe removes a subscriber and closes its message stream.
func (b *Broker) Unsubscribe(subscription *Subscription) bool {
	b.mu.Lock()
	topicSubscribers := b.topics[subscription.topic]
	if _, ok := topicSubscribers[subscription.id]; !ok {
		b.mu.Unlock()
		return false
	}
	delete(topicSubscribers, subscription.id)
	if len(topicSubscribers) == 0 {
		delete(b.topics, subscription.topic)
	}
	b.mu.Unlock()

	subscription.close()
	return true
}

func (s *Subscription) deliver(event Event) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return false
	}
	s.messages <- event
	return true
}

func (s *Subscription) tryDeliver(event Event) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return false
	}
	select {
	case s.messages <- event:
		return true
	default:
		return false
	}
}

func (b *Broker) subscribers(topic string) []*Subscription {
	b.mu.RLock()
	defer b.mu.RUnlock()

	subscribers := make([]*Subscription, 0, len(b.topics[topic]))
	for _, subscription := range b.topics[topic] {
		subscribers = append(subscribers, subscription)
	}
	return subscribers
}

func (s *Subscription) close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	s.closed = true
	close(s.messages)
}
