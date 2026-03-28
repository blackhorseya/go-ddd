package event

import "context"

// EventHandler is a function that processes a domain event.
type EventHandler func(c context.Context, event DomainEvent) error

// EventBus defines the contract for publishing and subscribing to domain events.
// Infrastructure layer provides concrete implementations (in-memory, Kafka, NATS, etc.).
type EventBus interface {
	// Publish sends one or more domain events to all registered handlers.
	Publish(c context.Context, events ...DomainEvent) error

	// Subscribe registers a handler for a specific event type.
	// The eventType should match DomainEvent.EventType() (e.g. "order.created").
	Subscribe(eventType string, handler EventHandler) error
}
