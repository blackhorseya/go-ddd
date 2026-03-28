// Package valueobject contains shared value objects used across bounded contexts.
//
// Value objects are immutable, identity-less types that represent domain concepts.
// They must be compared by value, not by reference.
//
// Conventions:
//   - All fields are private; access through getters only
//   - Provide a constructor (NewXxx) that validates invariants
//   - Value objects are immutable; modification methods return new instances
//   - Implement Equals() for comparison when needed
//
// Examples of shared value objects: Money, Address, Email, PhoneNumber, DateRange.
//
// BC-specific value objects should be placed in {bc}/domain/valueobject/ instead.
package valueobject
