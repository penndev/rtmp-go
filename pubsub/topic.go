package pubsub

import (
	"sync"

	"github.com/penndev/rtmp/rtmp/stream"
)

type Topic struct {
	broker *Broker
	name   string

	mu     sync.RWMutex
	subs   map[*Subscription]struct{}
	closed bool

	// process
	BeforePublish func(any)
	OnClose       func()
}

func (t *Topic) Publish(msg any) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.closed {
		return
	}
	if t.BeforePublish != nil {
		t.BeforePublish(msg)
	}
	for s := range t.subs {
		if s.Filter != nil && !s.Filter(msg) {
			continue
		}
		select {
		case s.ch <- msg:
		default:
			// slow subscriber: drop
		}
	}
}

func (t *Topic) Subscribe() stream.Subscriber {
	s := NewSubscription(t)
	t.Attach(s)
	return s
}

func (t *Topic) Attach(s *Subscription) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		s.mu.Lock()
		if !s.closed {
			s.closed = true
			close(s.ch)
		}
		s.mu.Unlock()
		return
	}
	t.subs[s] = struct{}{}
}

func (t *Topic) Close() {
	t.mu.Lock()
	if t.closed {
		t.mu.Unlock()
		return
	}
	t.closed = true
	onClose := t.OnClose
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

	if t.broker != nil {
		t.broker.mu.Lock()
		if t.broker.topics[t.name] == t {
			delete(t.broker.topics, t.name)
		}
		t.broker.mu.Unlock()
	}
	if onClose != nil {
		onClose()
	}
}
