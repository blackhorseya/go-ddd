// Domain events for the Credential aggregate.
//
// Events are defined here but NOT emitted by entity methods.
// Event publication is the responsibility of the application layer (use cases).
//
// Event type naming convention: "credential.{past_tense_verb}"
package credential

import "github.com/blackhorseya/go-ddd/internal/shared/domain/event"

// CredentialRegistered is published when a new credential is created.
type CredentialRegistered struct {
	event.BaseEvent

	Email string
}

// NewCredentialRegistered creates a CredentialRegistered event.
func NewCredentialRegistered(credID, email string) CredentialRegistered {
	return CredentialRegistered{
		BaseEvent: event.NewBaseEvent(credID, "credential", "credential.registered"),
		Email:     email,
	}
}

// CredentialActivated is published when a credential is activated.
type CredentialActivated struct {
	event.BaseEvent
}

// NewCredentialActivated creates a CredentialActivated event.
func NewCredentialActivated(credID string) CredentialActivated {
	return CredentialActivated{
		BaseEvent: event.NewBaseEvent(credID, "credential", "credential.activated"),
	}
}

// CredentialSuspended is published when a credential is suspended.
type CredentialSuspended struct {
	event.BaseEvent

	Reason string
}

// NewCredentialSuspended creates a CredentialSuspended event.
func NewCredentialSuspended(credID, reason string) CredentialSuspended {
	return CredentialSuspended{
		BaseEvent: event.NewBaseEvent(credID, "credential", "credential.suspended"),
		Reason:    reason,
	}
}
