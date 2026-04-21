// Package toolspec defines a structured knowledge base for CLI tools.
//
// Types here are pure data with zero transitive dependencies. Sub-packages
// (sources/help, sources/completion) own their own deps and are imported
// separately.
//
// # Core Types
//
//   - [ToolSpec]: full spec for a CLI tool (commands, flags, errors, workflows)
//   - [Command]: command tree node with safety, contract, output schema, intent
//   - [SafetyLevel]: safe / caution / dangerous classification
//   - [Contract]: idempotency, side effects, pre-conditions
//
// # Registry
//
// [Registry] resolves specs from ordered [Source] implementations with
// optional caching. Sources are queried in order; results are merged via
// [Merge]:
//
//	reg := toolspec.NewRegistry(
//	    toolspec.WithSource(sources.Help{}),
//	    toolspec.WithCache(store),
//	)
//	spec, _ := reg.Resolve(ctx, "kubectl")
//
// # Capabilities
//
// [CapabilitySet] describes discoverable capabilities of a running service.
// Used by api.WithCapabilities to serve GET /capabilities:
//
//	cs := toolspec.NewCapabilitySet("myapp", "1.0.0")
//	cs.Add(toolspec.Capability{Name: "list-items", Type: "endpoint", Path: "/items"})
//	cs.Merge(otherSet)
//	data, _ := cs.JSON()
//
// Key functions: [NewCapabilitySet], [CapabilitySet.Add],
// [CapabilitySet.JSON], [CapabilitySet.Merge].
package toolspec
