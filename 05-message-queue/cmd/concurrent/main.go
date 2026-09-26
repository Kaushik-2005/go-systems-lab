package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"message-queue/broker"
)

func main() {
	const producers = 3
	const messagesPerProducer = 10
	const consumers = 2
	totalMessages := producers * messagesPerProducer

	work := broker.New(2*time.Second, 3)

	var producerGroup sync.WaitGroup
	var produced atomic.Int64
	for producer := 0; producer < producers; producer++ {
		producerGroup.Add(1)
		go func(producer int) {
			defer producerGroup.Done()
			for messageNumber := 0; messageNumber < messagesPerProducer; messageNumber++ {
				work.Publish(fmt.Sprintf("producer-%d message-%d", producer, messageNumber))
				produced.Add(1)
			}
		}(producer)
	}
	producerGroup.Wait()

	var consumerGroup sync.WaitGroup
	var consumed atomic.Int64
	for consumer := 0; consumer < consumers; consumer++ {
		consumerGroup.Add(1)
		go func() {
			defer consumerGroup.Done()
			for consumed.Load() < int64(totalMessages) {
				delivery, ok := work.Consume()
				if !ok {
					continue
				}
				work.Ack(delivery.Message.ID)
				consumed.Add(1)
			}
		}()
	}
	consumerGroup.Wait()

	fmt.Printf("producers=%d consumers=%d produced=%d consumed=%d pending=%d in_flight=%d\n", producers, consumers, produced.Load(), consumed.Load(), work.PendingLen(), work.InFlightLen())
}
