package main

import (
	"context"
	"fmt"
	"time"

	"message-queue/broker"
)

func main() {
	work := broker.New(200*time.Millisecond, 2)
	monitorContext, cancel := context.WithCancel(context.Background())
	defer cancel()
	work.StartVisibilityMonitor(monitorContext, 50*time.Millisecond)
	work.Publish("first")
	work.Publish("second")
	work.Publish("third")

	fmt.Printf("published: pending=%d in_flight=%d\n", work.PendingLen(), work.InFlightLen())

	first, _ := work.Consume()
	second, _ := work.Consume()
	third, _ := work.Consume()
	fmt.Printf("consumed: %s, %s, %s pending=%d in_flight=%d\n", first.Message.ID, second.Message.ID, third.Message.ID, work.PendingLen(), work.InFlightLen())

	fmt.Printf("ack %s: success=%t in_flight=%d\n", first.Message.ID, work.Ack(first.Message.ID), work.InFlightLen())
	fmt.Printf("ack %s: success=%t in_flight=%d\n", third.Message.ID, work.Ack(third.Message.ID), work.InFlightLen())
	fmt.Printf("unacked message remains: id=%s in_flight=%d\n", second.Message.ID, work.InFlightLen())
	time.Sleep(300 * time.Millisecond)
	redelivered, ok := work.Consume()
	fmt.Printf("after first timeout: redelivered=%t id=%s attempts=%d pending=%d in_flight=%d\n", ok, redelivered.Message.ID, redelivered.Message.Attempts, work.PendingLen(), work.InFlightLen())
	time.Sleep(300 * time.Millisecond)
	_, _ = work.Consume()
	time.Sleep(300 * time.Millisecond)
	deadLetterCount := work.DeadLetterLen()
	dead, deadLettered := work.ConsumeDeadLetter()
	fmt.Printf("after retry limit: dead_lettered=%t id=%s attempts=%d dead_letters=%d\n", deadLettered, dead.ID, dead.Attempts, deadLetterCount)
}
