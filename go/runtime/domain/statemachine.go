package domain

import (
	"context"
	"fmt"
	"strings"
)

// State is a string label for entity states (e.g. "open", "closed").
type State string

// StateMachine enforces allowed state transitions and optionally publishes
// events for pre/post transition hooks via an EventPublisher.
type StateMachine struct {
	rules map[State][]State
	pub   EventPublisher
}

// NewStateMachine creates a state machine with the given transition rules.
// The publisher parameter is optional (nil disables event hooks).
func NewStateMachine(rules map[State][]State, pub EventPublisher) *StateMachine {
	return &StateMachine{rules: rules, pub: pub}
}

// TransitionError is returned when a state transition is not allowed.
// It carries the attempted from/to states and the list of valid targets.
type TransitionError struct {
	From    State
	To      State
	Allowed []State
}

// Error implements the error interface.
func (e *TransitionError) Error() string {
	if e.Allowed == nil {
		return fmt.Sprintf(
			"invalid state transition: no rules for state %q (attempted %q → %q)",
			e.From, e.From, e.To,
		)
	}
	allowed := make([]string, len(e.Allowed))
	for i, s := range e.Allowed {
		allowed[i] = string(s)
	}
	return fmt.Sprintf(
		"invalid state transition: %q → %q (allowed: [%s])",
		e.From, e.To, strings.Join(allowed, " "),
	)
}

// Is reports whether target matches ErrInvalidTransition.
func (e *TransitionError) Is(target error) bool {
	return target == ErrInvalidTransition
}

// AllowedFrom returns the valid target states from the given state.
// Returns nil if the state has no rules defined.
func (sm *StateMachine) AllowedFrom(state State) []State {
	allowed, ok := sm.rules[state]
	if !ok {
		return nil
	}
	cp := make([]State, len(allowed))
	copy(cp, allowed)
	return cp
}

// TransitionPayload is the event payload for state transitions.
type TransitionPayload struct {
	From  State
	To    State
	Force bool
}

// Transition validates and executes a state change from → to.
//
// When force is true the rules check is skipped.
//
// If an EventPublisher is configured:
//   - A synchronous "domain.state.pre-transition" event is published first;
//     any subscriber returning an error vetoes the transition.
//   - A best-effort "domain.state.post-transition" event fires in a goroutine
//     after success (non-blocking fire-and-forget).
func (sm *StateMachine) Transition(ctx context.Context, from, to State, force bool) error {
	if !force {
		allowed, ok := sm.rules[from]
		if !ok {
			return &TransitionError{From: from, To: to, Allowed: nil}
		}
		found := false
		for _, s := range allowed {
			if s == to {
				found = true
				break
			}
		}
		if !found {
			cp := make([]State, len(allowed))
			copy(cp, allowed)
			return &TransitionError{From: from, To: to, Allowed: cp}
		}
	}

	payload := TransitionPayload{From: from, To: to, Force: force}

	// Pre-transition: sync event (veto-able).
	if sm.pub != nil {
		if err := sm.pub.Publish(ctx, "domain.state.pre-transition", "domain.statemachine", payload); err != nil {
			return fmt.Errorf("pre-transition veto: %w", err)
		}
	}

	// Post-transition: fire-and-forget in a goroutine.
	if sm.pub != nil {
		go func() {
			_ = sm.pub.Publish(ctx, "domain.state.post-transition", "domain.statemachine", payload)
		}()
	}

	return nil
}
