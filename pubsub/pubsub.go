package pubsub

import "sync"

type Message struct {
	Data interface{}
}

type Broker struct {
	mu     sync.Mutex
	topics map[string]*Topic
}

type Topic struct {
	broker *Broker
	name   string

	mu     sync.RWMutex
	subs   map[*Subscription]struct{}
	closed bool
}

type Subscription struct {
	topic *Topic
	ch    chan *Message

	mu     sync.Mutex
	closed bool
}

func New() *Broker {
	return &Broker{topics: make(map[string]*Topic)}
}

func (b *Broker) Topic(name string) *Topic {
	b.mu.Lock()
	defer b.mu.Unlock()
	if t, ok := b.topics[name]; ok {
		return t
	}
	t := &Topic{
		broker: b,
		name:   name,
		subs:   make(map[*Subscription]struct{}),
	}
	b.topics[name] = t
	return t
}

func (t *Topic) Publish(msg *Message) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.closed {
		return
	}
	for s := range t.subs {
		select {
		case s.ch <- msg:
		default:
			// slow subscriber: drop
		}
	}
}

func (t *Topic) Subscribe() *Subscription {
	t.mu.Lock()
	defer t.mu.Unlock()
	s := &Subscription{
		topic: t,
		ch:    make(chan *Message, 64),
	}
	if t.closed {
		close(s.ch)
		s.closed = true
		return s
	}
	t.subs[s] = struct{}{}
	return s
}

func (s *Subscription) Chan() <-chan *Message {
	return s.ch
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

func (t *Topic) Close() {
	t.mu.Lock()
	if t.closed {
		t.mu.Unlock()
		return
	}
	t.closed = true
	subs := make([]*Subscription, 0, len(t.subs))
	for s := range t.subs {
		subs = append(subs, s)
	}
	t.subs = nil
	t.mu.Unlock()

	for _, s := range subs {
		s.mu.Lock()
		if !s.closed {
			s.closed = true
			close(s.ch)
		}
		s.mu.Unlock()
	}

	if t.broker == nil {
		return
	}
	t.broker.mu.Lock()
	if t.broker.topics[t.name] == t {
		delete(t.broker.topics, t.name)
	}
	t.broker.mu.Unlock()
}
