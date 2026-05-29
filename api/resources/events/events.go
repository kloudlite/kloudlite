package events

import (
	"context"
	"sync"
)

type DirtyReason string

const (
	DirtyReasonPendingCreate DirtyReason = "pending_create_confirmation"
	DirtyReasonPendingPatch  DirtyReason = "pending_patch_confirmation"
	DirtyReasonPendingDelete DirtyReason = "pending_delete_confirmation"
)

type EventType string

const (
	EventInitialState EventType = "INITIAL_STATE"
	EventAdded        EventType = "ADDED"
	EventModified     EventType = "MODIFIED"
	EventDeleted      EventType = "DELETED"
	EventDirty        EventType = "DIRTY"
	EventClean        EventType = "CLEAN"
	EventResyncing    EventType = "RESYNCING"
	EventResynced     EventType = "RESYNCED"
)

type DirtyObject struct {
	Resource  string            `json:"resource"`
	Namespace string            `json:"namespace,omitempty"`
	Name      string            `json:"name"`
	Reason    DirtyReason       `json:"reason"`
	Labels    map[string]string `json:"labels,omitempty"`
}

type Event struct {
	Type           EventType
	Resource       string
	Namespace      string
	Name           string
	Object         any
	Dirty          *DirtyObject
	PreviousLabels map[string]string
}

type Broker struct {
	mu          sync.Mutex
	subscribers map[chan Event]*subscription
}

type subscription struct {
	closed bool
}

func NewBroker() *Broker {
	return &Broker{subscribers: map[chan Event]*subscription{}}
}

func (b *Broker) Subscribe(ctx context.Context) <-chan Event {
	ch := make(chan Event, 32)
	b.mu.Lock()
	b.subscribers[ch] = &subscription{}
	b.mu.Unlock()
	go func() {
		<-ctx.Done()
		b.unsubscribe(ch)
	}()
	return ch
}

func (b *Broker) Publish(event Event) {
	if b == nil {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	for ch := range b.subscribers {
		select {
		case ch <- event:
		default:
			b.closeLocked(ch)
		}
	}
}

func (b *Broker) unsubscribe(ch chan Event) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.closeLocked(ch)
}

func (b *Broker) closeLocked(ch chan Event) {
	sub, ok := b.subscribers[ch]
	if !ok || sub.closed {
		return
	}
	sub.closed = true
	delete(b.subscribers, ch)
	close(ch)
}
