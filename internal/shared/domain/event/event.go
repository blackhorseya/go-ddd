// Package event provides the domain event infrastructure for inter-BC communication.
//
// Event type naming convention: "{aggregate}.{past_tense_verb}"
// Examples: "order.created", "user.registered", "payment.completed"
package event

import (
	"time"

	"github.com/google/uuid"
)

// DomainEvent represents a significant occurrence in the domain.
// All domain events must implement this interface.
type DomainEvent interface {
	// EventID returns the unique identifier of this event.
	EventID() string

	// EventType returns the type of the event (e.g. "order.created").
	EventType() string

	// OccurredAt returns the time when the event occurred.
	OccurredAt() time.Time

	// AggregateID returns the ID of the aggregate that produced this event.
	AggregateID() string

	// AggregateType returns the type of the aggregate (e.g. "order").
	AggregateType() string
}

// BaseEvent provides a reusable base implementation of DomainEvent.
// Embed this struct in concrete event types.
//
// Example:
//
//	type OrderCreated struct {
//	    event.BaseEvent
//	    UserID string
//	    Total  money.Money
//	}
//
//	func NewOrderCreated(orderID, userID string, total money.Money) OrderCreated {
//	    return OrderCreated{
//	        BaseEvent: event.NewBaseEvent(orderID, "order", "order.created"),
//	        UserID:    userID,
//	        Total:     total,
//	    }
//	}
type BaseEvent struct {
	id            string
	eventType     string
	occurredAt    time.Time
	aggregateID   string
	aggregateType string
}

// NewBaseEvent creates a new BaseEvent with a generated UUID and current timestamp.
func NewBaseEvent(aggregateID, aggregateType, eventType string) BaseEvent {
	return BaseEvent{
		id:            uuid.New().String(),
		eventType:     eventType,
		occurredAt:    time.Now(),
		aggregateID:   aggregateID,
		aggregateType: aggregateType,
	}
}

func (e BaseEvent) EventID() string       { return e.id }
func (e BaseEvent) EventType() string     { return e.eventType }
func (e BaseEvent) OccurredAt() time.Time { return e.occurredAt }
func (e BaseEvent) AggregateID() string   { return e.aggregateID }
func (e BaseEvent) AggregateType() string { return e.aggregateType }
