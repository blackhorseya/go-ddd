// Package persistence provides repository implementations for the Identity bounded context.
//
// Each storage backend gets its own sub-package (e.g. postgres/, memory/).
// Implementations must satisfy the repository interfaces defined in the domain layer.
//
// Use compile-time interface checks:
//
//	var _ user.Repository = (*UserRepository)(nil)
package persistence
