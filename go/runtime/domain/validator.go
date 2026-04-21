package domain

import "context"

// Validator validates an entity before persistence operations.
type Validator[T Entity] interface {
	// Validate checks entity invariants. Returns ErrValidation (or a
	// wrapped form) when the entity is invalid.
	Validate(ctx context.Context, entity T) error
}
