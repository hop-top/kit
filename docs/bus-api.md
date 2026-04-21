# Bus API Reference

> In-process pub/sub for CLI hooking. Events published to
> dot-separated topics; subscribers filter via MQTT-style
> wildcard patterns. Zero external dependencies (Go).

## Event

Standard envelope for all bus messages.

| Field       | Type      | Description                        |
|-------------|-----------|------------------------------------|
| `Topic`     | `Topic`   | dot-separated path, e.g.          |
|             | (string)  | `"llm.request.start"`             |
| `Source`    | `string`  | emitter id, e.g. `"llm.client"`  |
| `Timestamp` | `time`    | creation time (auto-set by        |
|             |           | `NewEvent`)                        |
| `Payload`  | `any`     | event-specific data                |

### Creating Events

| Language | Function                                     |
|----------|----------------------------------------------|
| Go       | `bus.NewEvent(topic, source, payload)`       |
| TS       | `createEvent(topic, source, payload)`        |
| Python   | `create_event(topic, source, payload)`       |

Timestamp set automatically to current time.

## Bus

Pub/sub hub. Create, subscribe, publish, close.

### Creating a Bus

| Language | Function                | Returns    |
|----------|-------------------------|------------|
| Go       | `bus.New()`             | `Bus`      |
| TS       | `createBus()`           | `Bus`      |
| Python   | `create_bus()`          | `Bus`      |

### Bus Interface (Go)

```go
type Bus interface {
    Publish(ctx context.Context, e Event) error
    Subscribe(pattern string, h Handler) Unsubscribe
    SubscribeAsync(pattern string, h AsyncHandler) Unsubscribe
    Close(ctx context.Context) error
}
```

### Subscribe

```go
unsub := bus.Subscribe("llm.*", func(ctx, e) error {
    // handle event
    return nil
})

// later
unsub()  // removes subscription
```

Returns an `Unsubscribe` function. Call it to remove the
subscription.

### Publish

Delivers event to all matching subscribers:

1. Sync handlers run in registration order
2. First sync error vetoes -- remaining handlers skipped
3. Async handlers launch after all sync handlers succeed

```go
err := b.Publish(ctx, bus.NewEvent(
    "llm.request", "client", payload,
))
```

### Handler Types (Go only)

| Type           | Signature                             |
|----------------|---------------------------------------|
| `Handler`      | `func(ctx, Event) error` -- sync,    |
|                | blocks publisher, can veto            |
| `AsyncHandler` | `func(ctx, Event)` -- goroutine,     |
|                | never blocks publisher                |

TS and Python: all handlers are async by default; sync
veto via returned promise rejection / raised exception.

## Topic Patterns

MQTT-style wildcards on dot-separated segments:

| Pattern          | Matches                              |
|------------------|--------------------------------------|
| `llm.request`    | exact: `llm.request` only           |
| `llm.*`          | one segment: `llm.request`,         |
|                  | `llm.response`; NOT `llm.req.start` |
| `llm.#`          | zero+ trailing: `llm`,              |
|                  | `llm.request`, `llm.request.start`  |

`#` must be the last segment. `*` matches exactly one.

## Sinks

Side-effect processors (logging, metrics, tracing). Errors
never block publish or handler delivery.

### Sink Interface

```go
type Sink interface {
    Drain(ctx context.Context, e Event) error
    Close() error
}
```

### Built-in Sinks

| Sink          | Output                                    |
|---------------|-------------------------------------------|
| `StdoutSink`  | human-readable to stdout                  |
|               | format: `[timestamp] topic source: payload`|
| `JSONLSink`   | newline-delimited JSON to writer/file     |

```go
// stdout sink
sink := bus.NewStdoutSink()

// JSONL to writer
sink := bus.NewJSONLSink(w)

// JSONL to file (owns handle; Close closes file)
sink, err := bus.NewJSONLSinkFile("/tmp/events.jsonl")
```

### TeeBus

Wraps a Bus and fans published events to sinks. Sink errors
reported via `ErrFunc` callback, never block publisher.

```go
tee := bus.NewTeeBus(b, []bus.Sink{jsonlSink}, onErr)
tee.Publish(ctx, event)  // bus + all sinks
```

## Lifecycle

### Close

Stops accepting new publishes (`ErrBusClosed` returned).
Waits for in-flight async handlers, respecting context
deadline.

```go
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()
err := b.Close(ctx)  // blocks until async done or timeout
```

## Future

Adapter interface for cross-process event delivery
(see `T-0529`). Current in-memory bus will become one
adapter; others (NATS, Redis PubSub) pluggable via same
`Bus` interface.

## Cross-Language Parity

| Feature           | Go     | TS       | Python   |
|-------------------|--------|----------|----------|
| Event type        | yes    | planned  | planned  |
| Bus create        | yes    | planned  | planned  |
| Subscribe         | yes    | planned  | planned  |
| MQTT wildcards    | yes    | planned  | planned  |
| Sync handlers     | yes    | n/a      | n/a      |
| Async handlers    | yes    | default  | default  |
| Sinks (Tee)       | yes    | planned  | planned  |
| Close / drain     | yes    | planned  | planned  |
