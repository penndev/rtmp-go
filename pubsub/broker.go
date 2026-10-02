package pubsub

import "sync"

type Broker struct {
	mu     sync.Mutex
	topics map[string]*Topic
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
