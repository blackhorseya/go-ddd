// Domain events for the User aggregate.
//
// Events are defined here but NOT emitted by entity methods.
// Event publication is the responsibility of the application layer (use cases).
//
// Event type naming convention: "user.{past_tense_verb}"
package user

import "github.com/blackhorseya/go-ddd/internal/shared/domain/event"

// UserRegistered is published when a new user is created.
type UserRegistered struct {
	event.BaseEvent

	Email string
	Name  string
}

// NewUserRegistered creates a UserRegistered event.
func NewUserRegistered(userID, email, name string) UserRegistered {
	return UserRegistered{
		BaseEvent: event.NewBaseEvent(userID, "user", "user.registered"),
		Email:     email,
		Name:      name,
	}
}

// UserActivated is published when a user account is activated.
type UserActivated struct {
	event.BaseEvent
}

// NewUserActivated creates a UserActivated event.
func NewUserActivated(userID string) UserActivated {
	return UserActivated{
		BaseEvent: event.NewBaseEvent(userID, "user", "user.activated"),
	}
}

// UserSuspended is published when a user account is suspended.
type UserSuspended struct {
	event.BaseEvent

	Reason string
}

// NewUserSuspended creates a UserSuspended event.
func NewUserSuspended(userID, reason string) UserSuspended {
	return UserSuspended{
		BaseEvent: event.NewBaseEvent(userID, "user", "user.suspended"),
		Reason:    reason,
	}
}
