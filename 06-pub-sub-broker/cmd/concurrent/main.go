package main

import (
	"fmt"
	"sync"

	"pub-sub-broker/broker"
)

func main() {
	news := broker.New()
	readerA := news.Subscribe("news", 20)
	readerB := news.Subscribe("news", 20)

	const publishers = 3
	const eventsPerPublisher = 5

	var group sync.WaitGroup
	for publisher := 1; publisher <= publishers; publisher++ {
		group.Add(1)
		go func(publisher int) {
			defer group.Done()
			for event := 1; event <= eventsPerPublisher; event++ {
				news.Publish("news", broker.Event{
					ID:   fmt.Sprintf("publisher-%d-event-%d", publisher, event),
					Body: fmt.Sprintf("from publisher %d", publisher),
				})
			}
		}(publisher)
	}

	group.Wait()

	receivedA := readAll(readerA)
	receivedB := readAll(readerB)
	expected := publishers * eventsPerPublisher
	fmt.Printf("publishers=%d events_each=%d expected=%d\n", publishers, eventsPerPublisher, expected)
	fmt.Printf("subscriber A received=%d\n", len(receivedA))
	fmt.Printf("subscriber B received=%d\n", len(receivedB))
	fmt.Printf("fan-out complete=%t\n", len(receivedA) == expected && len(receivedB) == expected)
}

func readAll(subscription *broker.Subscription) []broker.Event {
	events := make([]broker.Event, 0)
	for {
		select {
		case event := <-subscription.Messages():
			events = append(events, event)
		default:
			return events
		}
	}
}
