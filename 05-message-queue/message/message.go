package message

// Message is the unit of work stored by the queue.
type Message struct {
	ID       string
	Body     string
	Attempts int
}
