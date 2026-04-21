// Package hook provides a thread-safe lifecycle hook system.
//
// Extensions subscribe handlers to named hooks; the bus dispatches
// them in priority order (lower values run first).
//
// Internally delegates to kit/bus for pub/sub transport while
// maintaining priority ordering and the DispatchAll contract.
package hook

import (
	"context"
	"sort"
	"sync"

	"hop.top/kit/go/runtime/bus"
)

// Hook is a named lifecycle event.
type Hook string

// Common lifecycle hooks.
const (
	BeforeInit  Hook = "before_init"
	AfterInit   Hook = "after_init"
	BeforeClose Hook = "before_close"
	AfterClose  Hook = "after_close"
	BeforeRun   Hook = "before_run"
	AfterRun    Hook = "after_run"
)

// topicPrefix namespaces hook events on the bus.
const topicPrefix = "hook."

// Handler is a function invoked when a hook fires.
type Handler func(ctx context.Context, payload any) error

// entry pairs a handler with its priority.
type entry struct {
	handler  Handler
	priority int
}

// Bus is a thread-safe hook dispatcher backed by kit/bus.
type Bus struct {
	mu       sync.RWMutex
	handlers map[Hook][]entry
	inner    bus.Bus
}

// NewBus returns a ready-to-use Bus.
func NewBus() *Bus {
	return &Bus{
		handlers: make(map[Hook][]entry),
		inner:    bus.New(),
	}
}

// Subscribe registers a handler for the given hook.
// Lower priority values run first.
func (b *Bus) Subscribe(hook Hook, handler Handler, priority int) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.handlers[hook] = append(b.handlers[hook], entry{
		handler:  handler,
		priority: priority,
	})
	sort.SliceStable(b.handlers[hook], func(i, j int) bool {
		return b.handlers[hook][i].priority < b.handlers[hook][j].priority
	})
}

// Dispatch runs all handlers for hook in priority order.
// It stops and returns the first non-nil error.
// If ctx is already canceled, Dispatch returns ctx.Err() immediately.
func (b *Bus) Dispatch(ctx context.Context, hook Hook, payload any) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	b.mu.RLock()
	entries := make([]entry, len(b.handlers[hook]))
	copy(entries, b.handlers[hook])
	b.mu.RUnlock()

	// Notify cross-cutting observers asynchronously so they cannot
	// block or veto hook handler execution.
	b.notify(ctx, hook, payload)

	for _, e := range entries {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := e.handler(ctx, payload); err != nil {
			return err
		}
	}
	return nil
}

// DispatchAll runs all handlers for hook in priority order,
// collecting every error instead of stopping on the first.
// Context cancellation is checked before each handler.
func (b *Bus) DispatchAll(ctx context.Context, hook Hook, payload any) []error {
	if err := ctx.Err(); err != nil {
		return []error{err}
	}

	b.mu.RLock()
	entries := make([]entry, len(b.handlers[hook]))
	copy(entries, b.handlers[hook])
	b.mu.RUnlock()

	b.notify(ctx, hook, payload)

	var errs []error
	for _, e := range entries {
		if err := ctx.Err(); err != nil {
			errs = append(errs, err)
			return errs
		}
		if err := e.handler(ctx, payload); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

// Handlers returns the number of handlers registered for hook.
func (b *Bus) Handlers(hook Hook) int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.handlers[hook])
}

// notify publishes a hook event to the inner bus asynchronously.
// Observer errors and latency never affect hook handler execution.
func (b *Bus) notify(ctx context.Context, hook Hook, payload any) {
	go func() {
		_ = b.inner.Publish(ctx, bus.NewEvent(
			bus.Topic(topicPrefix+string(hook)), "ext.hook", payload,
		))
	}()
}

// Inner returns the underlying kit/bus instance for cross-cutting
// observers that want to subscribe to hook events via topic patterns
// (e.g. "hook.#" for all lifecycle events). Observers should use
// SubscribeAsync or ensure sync handlers are non-blocking and never
// return errors.
func (b *Bus) Inner() bus.Bus {
	return b.inner
}

// ---- package-level default bus ----

var defaultBus = NewBus()

// Default returns the package-level Bus.
func Default() *Bus { return defaultBus }

// Subscribe registers a handler on the default bus.
func Subscribe(hook Hook, handler Handler, priority int) {
	defaultBus.Subscribe(hook, handler, priority)
}

// Dispatch fires a hook on the default bus.
func Dispatch(ctx context.Context, hook Hook, payload any) error {
	return defaultBus.Dispatch(ctx, hook, payload)
}

// DispatchAll fires a hook on the default bus, collecting all errors.
func DispatchAll(ctx context.Context, hook Hook, payload any) []error {
	return defaultBus.DispatchAll(ctx, hook, payload)
}

// Handlers returns the handler count on the default bus.
func Handlers(hook Hook) int {
	return defaultBus.Handlers(hook)
}
