package watermill

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"

	"github.com/blackhorseya/go-ddd/internal/shared/domain/event"
)

// Compile-time check.
var _ event.EventBus = (*Bus)(nil)

// Bus implements event.EventBus using Watermill GoChannel (in-process pub/sub).
type Bus struct {
	pubsub   *gochannel.GoChannel
	router   *message.Router
	handlers map[string]event.EventHandler
	mu       sync.RWMutex
}

// NewBus creates a Watermill-backed EventBus.
func NewBus() (*Bus, error) {
	logger := watermill.NopLogger{}

	pubsub := gochannel.NewGoChannel(gochannel.Config{
		Persistent: false,
	}, logger)

	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		return nil, err
	}

	return &Bus{
		pubsub:   pubsub,
		router:   router,
		handlers: make(map[string]event.EventHandler),
	}, nil
}

// Publish serialises domain events and publishes them to the corresponding topic.
func (b *Bus) Publish(c context.Context, events ...event.DomainEvent) error {
	for _, evt := range events {
		payload, err := json.Marshal(eventEnvelope{
			EventID:       evt.EventID(),
			EventType:     evt.EventType(),
			AggregateID:   evt.AggregateID(),
			AggregateType: evt.AggregateType(),
			OccurredAt:    evt.OccurredAt().Unix(),
		})
		if err != nil {
			return err
		}

		msg := message.NewMessage(evt.EventID(), payload)
		msg.SetContext(c)

		if err := b.pubsub.Publish(evt.EventType(), msg); err != nil {
			return err
		}
	}

	return nil
}

// Subscribe registers a handler for a specific event type.
func (b *Bus) Subscribe(eventType string, handler event.EventHandler) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.handlers[eventType] = handler

	return nil
}

// Close shuts down the underlying pubsub.
func (b *Bus) Close() error {
	return b.pubsub.Close()
}

// eventEnvelope is the JSON structure published to Watermill topics.
type eventEnvelope struct {
	EventID       string `json:"event_id"`
	EventType     string `json:"event_type"`
	AggregateID   string `json:"aggregate_id"`
	AggregateType string `json:"aggregate_type"`
	OccurredAt    int64  `json:"occurred_at"`
}
