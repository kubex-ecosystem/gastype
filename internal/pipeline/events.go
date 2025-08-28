package pipeline

type Topic string

const (
	TopicNodeVisited Topic = "node.visited"
	TopicNodeChanged Topic = "node.changed"
	TopicNodeSkipped Topic = "node.skipped"
)

type Event struct {
	Topic   Topic
	Payload any
}

type EventBus struct{ subs map[Topic][]chan Event }

func NewEventBus() *EventBus { return &EventBus{subs: map[Topic][]chan Event{}} }

func (b *EventBus) Subscribe(t Topic, buf int) <-chan Event {
	ch := make(chan Event, buf)
	b.subs[t] = append(b.subs[t], ch)
	return ch
}

func (b *EventBus) Publish(e Event) {
	if ls, ok := b.subs[e.Topic]; ok {
		for _, ch := range ls {
			select {
			case ch <- e:
			default:
			} // non-blocking
		}
	}
}
