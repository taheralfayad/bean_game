package gameevents

import "sync"

type Broker struct {
	mu     sync.RWMutex
	topics map[string]map[chan Event]struct{}
	closed bool
}

func NewBroker() *Broker {
	return &Broker{topics: make(map[string]map[chan Event]struct{})}
}

func (b *Broker) Subscribe(topic string) (<-chan Event, func()) {
	client := make(chan Event)

	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		close(client)
		return client, func() {}
	}

	if b.topics[topic] == nil {
		b.topics[topic] = make(map[chan Event]struct{})
	}
	b.topics[topic][client] = struct{}{}
	b.mu.Unlock()

	var once sync.Once
	unsubscribe := func() {
		once.Do(func() {
			b.mu.Lock()
			defer b.mu.Unlock()

			delete(b.topics[topic], client)
			if len(b.topics[topic]) == 0 {
				delete(b.topics, topic)
			}

			close(client)
		})
	}

	return client, unsubscribe
}

func (b *Broker) Publish(event Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return
	}

	for client := range b.topics[event.Topic] {
		select {
		case client <- event:
		default:
		}
	}
}

func (b *Broker) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return
	}
	b.closed = true

	for _, clients := range b.topics {
		for client := range clients {
			close(client)
		}
	}

	b.topics = make(map[string]map[chan Event]struct{})
}
