package domain

import (
	"context"
	"time"
)

// Option configures a Service via functional options.
type Option[T Entity] func(*Service[T])

// WithAudit attaches an audit repository to the service.
func WithAudit[T Entity](ar AuditRepository) Option[T] {
	return func(s *Service[T]) { s.audit = ar }
}

// WithPublisher attaches an EventPublisher for domain events.
func WithPublisher[T Entity](pub EventPublisher) Option[T] {
	return func(s *Service[T]) { s.pub = pub }
}

// WithValidation attaches a validator to the service.
func WithValidation[T Entity](v Validator[T]) Option[T] {
	return func(s *Service[T]) { s.validator = v }
}

// Service wraps a Repository with optional middleware: validation, auditing,
// event publishing.
type Service[T Entity] struct {
	repo      Repository[T]
	audit     AuditRepository
	pub       EventPublisher
	validator Validator[T]
}

// NewService creates a Service around the given repository.
// Additional middleware is configured via Option functions.
func NewService[T Entity](repo Repository[T], opts ...Option[T]) *Service[T] {
	s := &Service[T]{repo: repo}
	for _, o := range opts {
		o(s)
	}
	return s
}

// Create validates (if configured), persists, audits, and publishes.
func (s *Service[T]) Create(ctx context.Context, entity *T) error {
	if s.validator != nil {
		if err := s.validator.Validate(ctx, *entity); err != nil {
			return err
		}
	}
	if err := s.repo.Create(ctx, entity); err != nil {
		return err
	}
	id := (*entity).GetID()
	s.auditAction(ctx, id, "created")
	s.publishEvent(ctx, "domain.entity.created", id)
	return nil
}

// Get retrieves an entity by ID.
func (s *Service[T]) Get(ctx context.Context, id string) (*T, error) {
	return s.repo.Get(ctx, id)
}

// List retrieves entities matching the query.
func (s *Service[T]) List(ctx context.Context, q Query) ([]T, error) {
	return s.repo.List(ctx, q)
}

// Update validates (if configured), persists, audits, and publishes.
func (s *Service[T]) Update(ctx context.Context, entity *T) error {
	if s.validator != nil {
		if err := s.validator.Validate(ctx, *entity); err != nil {
			return err
		}
	}
	if err := s.repo.Update(ctx, entity); err != nil {
		return err
	}
	id := (*entity).GetID()
	s.auditAction(ctx, id, "updated")
	s.publishEvent(ctx, "domain.entity.updated", id)
	return nil
}

// Delete removes an entity, audits, and publishes.
func (s *Service[T]) Delete(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	s.auditAction(ctx, id, "deleted")
	s.publishEvent(ctx, "domain.entity.deleted", id)
	return nil
}

func (s *Service[T]) auditAction(ctx context.Context, entityID, action string) {
	if s.audit == nil {
		return
	}
	_ = s.audit.AddEntry(ctx, &AuditEntry{
		EntityID:  entityID,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Action:    action,
	})
}

// publishEvent does a best-effort publish of a domain event.
func (s *Service[T]) publishEvent(ctx context.Context, topic, entityID string) {
	if s.pub == nil {
		return
	}
	_ = s.pub.Publish(ctx, topic, "domain.service", map[string]string{"id": entityID})
}
