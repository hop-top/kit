// Package bus provides pub/sub event delivery for CLI hooking.
//
// Events are published to topics; subscribers receive events matching
// their topic pattern. Wildcards follow MQTT conventions:
//   - `*` matches exactly one segment
//   - `#` matches zero or more trailing segments
//
// # Adapters
//
// Bus supports pluggable adapters via [WithAdapter]:
//   - MemoryAdapter: in-process, goroutine-based async (default)
//   - SQLiteAdapter: cross-process via shared SQLite + polling
//   - NetworkAdapter: cross-machine via WebSocket
//
// # Network Adapter
//
// NetworkAdapter bridges local bus instances over WebSocket for
// cross-machine event delivery. Usage:
//
//	b := bus.New()
//	na := bus.NewNetworkAdapter(b,
//	    bus.WithOriginID("node-1"),
//	    bus.WithFilter(bus.TopicFilter{Allow: []string{"task.*"}}),
//	    bus.WithAuth(&bus.StaticTokenAuth{Token_: "secret"}),
//	)
//	na.Connect(ctx, "ws://peer:8080/bus")
//
// Or use the [WithNetwork] option for auto-connect on construction:
//
//	b := bus.New(bus.WithNetwork("ws://peer:8080/bus"))
//
// To accept inbound peers, mount the handler on an HTTP server:
//
//	http.Handle("/bus", na.Handler())
//
// Features:
//   - Exponential backoff reconnect (100ms base, 30s cap, jitter)
//   - Topic filtering: deny-first, then allow (glob patterns)
//   - Loop prevention via origin tracking
//   - Auth handshake (JWT or static token via [Authenticator])
//   - JSON text frames compatible with standard tooling
package bus
