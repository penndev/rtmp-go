package pubsub

import "sync"

type Subscription struct {
	topic *Topic
	ch    chan any

	mu     sync.Mutex
	closed bool

	// Filter: return false to drop this message. nil means accept all.
	Filter func(any) bool
}

func NewSubscription(topic *Topic) *Subscription {
	return &Subscription{
		topic: topic,
		ch:    make(chan any, 64),
	}
}

func (s *Subscription) Chan() <-chan any {
	return s.ch
}

func (s *Subscription) Write(msg any) {
	s.ch <- msg
}

func (s *Subscription) Close() {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return
	}
	s.closed = true
	close(s.ch)
	s.mu.Unlock()

	if s.topic == nil {
		return
	}
	s.topic.mu.Lock()
	delete(s.topic.subs, s)
	s.topic.mu.Unlock()
}
