# aim

> **Incubating** — this package is developed in
> [hop-top/kit](https://github.com/hop-top/kit/tree/main/incubator/aim).
> Submit issues, PRs, and discussions there.

AI model registry and cache backed by [models.dev](https://models.dev).

Discover, filter, and inspect models across providers with local
caching, ETag-based conditional refresh, and a structured query
language.

## Install

```
go get hop.top/aim
```

## Library

### Registry

```go
reg, _ := aim.NewRegistry(
    aim.WithSource(aim.NewModelsDevSource()),
    aim.WithRegistryTTL(1 * time.Hour),
)

// all models
all := reg.Models(aim.Filter{})

// filtered
tools := reg.Models(aim.Filter{ToolCall: ptr(true)})

// single lookup
m, ok := reg.Get("anthropic", "claude-sonnet-4-5-20250514")

// providers
provs := reg.Providers()
```

### Cache

Disk-backed cache with TTL and ETag conditional-GET:

```go
src := aim.NewModelsDevSource()
cache, _ := aim.NewCache(src,
    aim.WithCacheDir("~/.cache/aim"),
    aim.WithTTL(6 * time.Hour),
)
```

### Source Options

```go
aim.NewModelsDevSource(
    aim.WithURL("https://custom-mirror.example/api.json"),
    aim.WithHTTPClient(customClient),
    aim.WithTimeout(10 * time.Second),
    aim.WithMaxResponseSize(100 << 20),
)
```

### Query Language

```go
f, _ := aim.ParseQuery("anthropic tool_call:true input:text")
models := reg.Models(f)
```

Filter fields: `Provider`, `Family`, `Input`, `Output`,
`ToolCall`, `Reasoning`, `OpenWeights`, `Query` (free-text).

## CLI

Mounts as a `cobra.Command` via `aim.Cmd()`.

```
models [query]              Browse AI model catalog
models list [query]         List models (default)
models show <provider> <id> Show model details
models providers            List providers
models query <string>       Query models (alias for list)
models refresh              Force-refresh model cache
```

### Flags

| Scope | Flag | Description |
|-------|------|-------------|
| Global | `--format` | Output format: `table`, `json` |
| `list` | `--provider` | Filter by provider ID |
| `list` | `--family` | Filter by model family |
| `list` | `--input` | Filter by input modality |
| `list` | `--output` | Filter by output modality |
| `list` | `--tool-call` | Require tool-call support |
| `list` | `--reasoning` | Require reasoning support |
| `list` | `--open-weights` | Require open-weights |

## Contents

- [testdata/](testdata/README.md): test fixtures and schemas.

## License

MIT
