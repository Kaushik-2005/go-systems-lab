package main

import (
	"fmt"

	"pub-sub-broker/broker"
)

func main() {
	news := broker.New()

	readerA := news.Subscribe("news", 4)
	readerB := news.Subscribe("news", 4)
	eventsA := make([]broker.Event, 0)
	eventsB := make([]broker.Event, 0)

	for number, body := range []string{"one", "two", "three"} {
		event := broker.Event{ID: fmt.Sprintf("event-%d", number+1), Body: body}
		delivered := news.Publish("news", event)
		eventsA = append(eventsA, readAvailable(readerA)...)
		eventsB = append(eventsB, readAvailable(readerB)...)
		printState(event, delivered, eventsA, eventsB)
	}

	fmt.Printf("unsubscribe A: success=%t\n", news.Unsubscribe(readerA))
	event := broker.Event{ID: "event-4", Body: "four"}
	delivered := news.Publish("news", event)
	eventsB = append(eventsB, readAvailable(readerB)...)
	printState(event, delivered, eventsA, eventsB)

	slow := news.Subscribe("alerts", 1)
	firstDelivered, firstDropped := news.PublishNonBlocking("alerts", broker.Event{ID: "alert-1", Body: "first"})
	secondDelivered, secondDropped := news.PublishNonBlocking("alerts", broker.Event{ID: "alert-2", Body: "second"})
	fmt.Printf("non-blocking policy: alert-1 delivered=%d dropped=%d\n", firstDelivered, firstDropped)
	fmt.Printf("non-blocking policy: alert-2 delivered=%d dropped=%d\n", secondDelivered, secondDropped)
	fmt.Printf("slow subscriber received: %v\n", readAvailable(slow))
}

func readAvailable(subscription *broker.Subscription) []broker.Event {
	events := make([]broker.Event, 0)
	for {
		select {
		case event, ok := <-subscription.Messages():
			if !ok {
				return events
			}
			events = append(events, event)
		default:
			return events
		}
	}
}

func printState(event broker.Event, delivered int, eventsA, eventsB []broker.Event) {
	fmt.Printf("published %s (%q) to %d subscribers\n", event.ID, event.Body, delivered)
	fmt.Printf("  A has: %s\n", formatEvents(eventsA))
	fmt.Printf("  B has: %s\n\n", formatEvents(eventsB))
}

func formatEvents(events []broker.Event) string {
	formatted := make([]string, 0, len(events))
	for _, event := range events {
		formatted = append(formatted, fmt.Sprintf("%s(%s)", event.ID, event.Body))
	}
	return fmt.Sprintf("[%s]", join(formatted, ", "))
}

func join(values []string, separator string) string {
	result := ""
	for index, value := range values {
		if index > 0 {
			result += separator
		}
		result += value
	}
	return result
}
