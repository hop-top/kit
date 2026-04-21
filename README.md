# hop.top/kit

Shared multi-language library for the hop-top family of tools.
Follows same monorepo pattern as `hop-top/upgrade`.

> **Pre-1.0** — breaking changes expected between minor versions.

## Project Layout

- [cmd/](cmd/README.md): Entry points for binaries.
- [go/](go/README.md): Core polyglot library implementations.
- [contracts/](contracts/README.md): Shared schemas and protobufs.
- [sdk/](sdk/README.md): Language-specific client libraries.
- [engine/](engine/README.md): Sidecar product line components.
- [incubator/](incubator/README.md): Experimental and emerging packages.
- [templates/](templates/README.md): Scaffolding blueprints.
- [examples/](examples/README.md): Runnable integration samples.
- [docs/](docs/README.md): ADRs, plans, and stories.

## Packages

### Go (`hop.top/kit`)

| Package           | Purpose                                       |
|-------------------|-----------------------------------------------|
| `go/api`         | HTTP toolkit: router, middleware, resources, capabilities |
| `go/bus`         | Pub/sub with memory, SQLite, and network adapters |
| `go/cli`         | fang+cobra+viper root command factory         |
| `go/identity`    | Local-first Ed25519 identity, JWT, encryption |
| `go/config`      | Layered config loader (system->user->project->env) |
| `go/domain`      | Generic DDD building blocks (entity, repo, state machine, service) |
| `go/domain/version` | Append-only version DAG for entity history tracking         |
| `go/ext`         | Shared extensibility contract (capabilities)  |
| `go/ext/config`  | Config-driven feature toggling                |
| `go/ext/discover`| PATH-based external plugin discovery          |
| `go/ext/dispatch`| Git-style plugin→cobra subcommand bridge      |
| `go/ext/hook`    | Lifecycle hooks (delegates to kit/bus)         |
| `go/ext/registry`| init()-based plugin registration              |
| `go/llm`         | Provider-agnostic LLM client (multimodal, fallback, hooks)|
| `go/log`         | Viper-configured charm.land/log/v2 wrapper    |
| `go/markdown`    | Glamour v2 terminal markdown renderer         |
| `go/output`      | table/json/yaml renderer; owns `--format`     |
| `go/peer`        | Decentralized peer discovery, trust mesh, TOFU|
| `go/sync`        | Local-first multi-remote entity replication    |
| `go/sqlstore`    | Generic SQLite kv store with TTL              |
| `go/toolspec`    | Structured CLI tool knowledge base            |
| `go/tui`         | Pre-themed bubbletea/v2 components            |
| `go/upgrade`     | Self-upgrade check, download, and replace     |
| `go/uxp`         | AI CLI detection, project keys, doctor diags  |
| `go/util`        | Stdlib-only helpers: env, fingerprint, humanize, jsonl, must, ptr, retry, since, slug |
| `go/xdg`         | XDG path resolution (config/data/cache/state) |
| `go/aim`         | AI model registry (query, cache, multi-source) |
| `go/qmochi`      | Terminal charting: bar, column, line, sparkline, heatmap, braille |

### TypeScript (`@hop-top/kit`)

| Subpath               | Purpose                    |
|-----------------------|----------------------------|
| `@hop-top/kit/cli`    | Commander program factory  |
| `@hop-top/kit/output` | table/json/yaml renderer   |

### Python (`hop-top-kit`)

| Module                | Purpose               |
|-----------------------|-----------------------|
| `hop_top_kit.cli`     | Typer app factory     |
| `hop_top_kit.output`  | table/json/yaml renderer |

## CLI contract (all languages)

Every tool built with `go/cli` behaves identically:

| Behaviour         | Detail                                           |
|-------------------|--------------------------------------------------|
| No `help` subcmd  | Hidden; only `-h`/`--help` flag at any level     |
| No `completion`   | Hidden or omitted entirely                       |
| `-v`/`--version`  | Root-only; handled by fang; `<tool> <version>`   |
| `--format`        | `table` (default) / `json` / `yaml` (kit/output) |
| `--quiet`         | Suppress non-essential output                    |
| `--no-color`      | Disable ANSI colour                              |
| Error messages    | To stderr; non-zero exit; no stack traces        |
| Theme accent      | Per-tool hex colour via `Config.Accent`          |

## Capabilities endpoint

Services built with `go/api` can self-describe via `/capabilities`:

```go
r := api.NewRouter(api.WithCapabilities("myapp", "1.0.0"))
r.Handle("GET", "/health", healthHandler)
r.MountResource("/api/widgets", widgetRouter, "create", "list", "get", "update", "delete")
// GET /capabilities → CapabilitySet JSON
```

The `cli.WithAPI` integration auto-wires capabilities when configured:

```go
root := cli.New(cli.Config{Name: "myapp", Version: "1.0.0", Short: "..."})
cli.WithAPI(cli.APIConfig{
    Addr: ":8080",
    Capabilities: &cli.CapabilitiesConfig{ServiceName: "myapp", Version: "1.0.0"},
})(root)
```

## Typical usage (Go)

```go
root := cli.New(cli.Config{
    Name:    "mytool",
    Version: "1.2.3",
    Short:   "does things",
    Accent:  "#E040FB",  // optional per-tool accent
})
root.Cmd.AddCommand(serveCmd(), listCmd())
if err := root.Execute(context.Background()); err != nil {
    os.Exit(1)
}
```

`root.Execute(ctx)` runs the cobra command through fang, which handles
`--version`, styled help, and error formatting. Subcommands read
`root.Viper` for `quiet`, `no-color`, and `format`.

## Adoption

### Go tools (tlc, rsx, ben, mdl, ...)

```
go get hop.top/kit@latest
```

- `cli.New(cli.Config{...})` -- fang+cobra+viper root command
- `root.Execute(ctx)` -- run CLI (replaces `root.Cmd.Execute()`)
- `xdg.ConfigDir("toolname")` -- replaces per-tool XDG path logic
- `config.Load(&cfg, config.Options{...})` -- replaces local loaders
- `sqlstore.Open(path, opts)` -- replaces local SQLite cache
- `output.Render(w, format, v)` -- replaces local renderers
- `log.New(v)` -- viper-configured structured logger
- `markdown.Render(src, noColor)` -- terminal markdown
- `tui.NewSpinner(...)` / `tui.NewProgress(...)` -- themed widgets

### TS tools (idx, eva-pkg, ...)

```
pnpm add @hop-top/kit
```

- `createCLI({name, version, description})` -- replaces Commander setup

### Python tools (eva, eva-ee, ...)

```
pip install hop-top-kit
```

- `create_app(name=..., version=..., help=...)` -- replaces Typer setup

## Migrating from v0

### Import paths

charm.land v2 replaces charmbracelet v1:

```
# before
github.com/charmbracelet/bubbletea
github.com/charmbracelet/lipgloss

# after
charm.land/bubbletea/v2
charm.land/lipgloss/v2
```

### CLI wiring

`go/cli` now uses fang for execution:

```go
// before (v0)
root := cli.New(cli.Config{...})
root.Cmd.Execute()

// after (v1)
root := cli.New(cli.Config{...})
root.Execute(context.Background())
```

### --format flag

Previously each tool registered its own `--format` flag.
Now `go/output.RegisterFlags` handles it -- transparent to callers.
Tools that used `output.Render` need no changes.

### --version flag

Previously each tool wired `--version` via cobra.
Now fang handles it -- transparent. Just set `Config.Version`.

## Dev environment

Three ways to get a fully configured environment:

| Method | Command |
|--------|---------|
| Nix flakes | `nix develop -f .devcontainer/flake.nix` |
| Devbox | `devbox shell -c .devcontainer/devbox.json` |
| Dev Containers | Open in VS Code or Codespaces |

All paths use `.devcontainer/flake.nix` as source of truth. Pinned: Go 1.26,
Node 22, Python 3.13, golangci-lint, eslint, ruff, go-task, git,
gh CLI.

### AI coding tools

Optional — never baked into the dev image:

```
task ai:setup          # interactive multi-select
task ai:claude         # install Claude Code
task ai:copilot        # install GitHub Copilot CLI
```

Or create `.ai-tools` (one tool name per line) for auto-install
on container create.

### CI

CI reuses the same Dockerfile via `.github/actions/dev-env/`.
Docker layers cached; nix stage rebuilds only when flake files
change.

## Adding a new package

Keep packages small and independently importable.
Minimal intra-kit dependencies:

- `go/tui` → `go/cli` (Theme), `go/upgrade` (Badge)
- `go/ext/hook` → `go/bus` (pub/sub transport)
- `go/ext/dispatch` → `go/ext/discover`, `cobra` (opt-in)
- `go/llm` → `go/bus` (lifecycle event hooks)
- `go/upgrade` → `go/xdg` (path resolution), `go/output` (CLI formatting)
- `go/wizard/wizardtui` → `go/wizard` (engine), `go/cli` (Theme), `go/tui` (Spinner)
- `go/cli` does **not** import `go/ext`

All other packages are independently importable.

## Multi-Protocol API

The `api/` and `rpc/` packages provide a unified approach to
exposing the same `Service[T]` over REST, WebSocket, and
ConnectRPC. See `examples/multiprotocol/` for a runnable server.

### REST + OpenAPI

```go
r := api.NewRouter(api.WithOpenAPI(api.OpenAPIConfig{
    Title: "My API", Version: "1.0.0",
}))
r.Mount("/widgets", api.ResourceRouter[Widget](svc,
    `api.WithHumaAPI[Widget](api.HumaAPI`(r)`, "/widgets")`,
))
```

OpenAPI 3.1 spec served at `/openapi.json`. Uses
[huma](https://huma.rocks) for typed operation registration.

### WebSocket

```go
hub := api.NewHub()
go hub.Run(ctx)
r.Handle("GET", "/ws", api.WSHandler(hub))
```

Topic-based pub/sub; clients subscribe via JSON messages.
Glob patterns (`*`, `**`) supported. Use `BusAdapter` to
bridge an event bus to WS clients.

### ConnectRPC

```go
rpcSrv := rpc.NewServer(
    rpc.WithInterceptors(rpc.RequestIDInterceptor()),
)
path, handler := rpc.RPCResource[Widget](svc,
    connect.WithInterceptors(rpcSrv.Interceptors()...),
)
rpcSrv.Handle(path, handler)
rpc.ListenAndServe(ctx, ":8081", rpcSrv)
```

Generic CRUD proto — no per-entity codegen. Interceptors
mirror api/ middleware (auth, logging, recovery, request ID).

### Polyglot Clients

| Language   | Import                              |
|------------|-------------------------------------|
| Go (REST)  | `hop.top/kit/go/transport/api/client`            |
| Go (WS)    | `hop.top/kit/go/transport/api/client` (DialWS)   |
| Go (RPC)   | `hop.top/kit/go/transport/rpc/client`            |
| TypeScript | `@hop-top/kit/api`, `@hop-top/kit/rpc` |
| PHP        | `HopTop\Kit\Api\*`                  |
| Rust       | `hop_top_kit::api`                  |
