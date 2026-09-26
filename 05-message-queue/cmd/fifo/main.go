package main

import (
	"fmt"

	"message-queue/message"
	"message-queue/queue"
)

func main() {
	work := queue.New()
	for _, body := range []string{"first", "second", "third"} {
		work.Enqueue(message.Message{Body: body})
	}

	fmt.Printf("queued=%d\n", work.Len())
	for {
		item, ok := work.Dequeue()
		if !ok {
			break
		}
		fmt.Printf("delivered=%s remaining=%d\n", item.Body, work.Len())
	}
}
