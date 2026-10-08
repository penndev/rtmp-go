package stream

type Publisher interface {
	Publish(any)
	Close()
}

type Subscriber interface {
	Chan() <-chan any
	Close()
}

type Stream interface {
	Publish(string) (Publisher, error)
	Play(string) (Subscriber, error)
}
