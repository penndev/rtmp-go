package pubsub

import "sync"

type Subscription struct {
	topic *Topic
	ch    chan *Message

	mu     sync.Mutex
	closed bool

	// Filter: return false to drop this message. nil means accept all.
	Filter func(*Message) bool
}

func (s *Subscription) Chan() <-chan *Message {
	return s.ch
}

func (s *Subscription) Write(msg *Message) {
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
