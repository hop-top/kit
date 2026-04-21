# kit — Agent Instructions

## Cross-Language Parity

kit is a polyglot shared-library: Go (primary), TypeScript, Python.
**Every user-facing feature must ship in all 3 languages.**

### Architecture

```
contracts/parity/parity.json   ← single source of truth for constants
Go:     cli/, log/, output/, tui/
TS:     sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/ts/src/          (flat files, vitest, CommonJS)
Python: sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/py/hop_top_kit/  (flat files, pytest, hatch)
```

### Parity contract

- `contracts/parity/parity.json` defines shared constants (symbols,
  spinner frames, help layout, section order). All languages
  load from this file — never hardcode.
- `cli/parity_test.go` (Go, `-tags parity`) validates that
  TS and Python CLIs produce identical help/flag output. Run
  via `go test -tags parity ./cli/...`.
- When adding a new built-in flag, constant, or behavioral
  contract: add it to `parity.json` first, then implement in
  all 3 languages, then add a parity test case.

### File conventions

| Concern | Go | TypeScript | Python |
|---------|-----|------------|--------|
| CLI framework | `cli/cli.go` | `sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/ts/src/cli.ts` | `sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/py/hop_top_kit/cli.py` |
| Logging | `log/log.go` | `sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/ts/src/log.ts` | `sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/py/hop_top_kit/log.py` |
| Output/format | `output/` | `sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/ts/src/output.ts` | `sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/py/hop_top_kit/output.py` |
| Stream writer | `cli/stream.go` | `sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/ts/src/stream.ts` | `sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/py/hop_top_kit/stream.py` |
| Tests | `*_test.go` | `*.test.ts` | `sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/py/tests/test_*.py` |

### Built-in global flags (all languages)

| Flag | Short | Viper key | Effect |
|------|-------|-----------|--------|
| `--quiet` | — | `quiet` | Suppress non-essential output (log → Warn) |
| `--verbose` | `-V` | `verbose` | Stackable: -V=debug, -VV=trace. Quiet overrides. |
| `--no-color` | — | `no-color` | Disable ANSI color |
| `--format` | — | `format` | Output format: table, json, yaml |
| `--no-hints` | — | `no-hints` | Suppress next-step hints |
| `--stream` | — | per-cmd | Enable named output streams (repeatable, comma-sep) |

### Named streams (per-command)

Commands register diagnostic streams at setup time. Users
toggle with `--stream <name>`. Channel returns `[name]`-prefixed
stderr writer when enabled, no-op when disabled.

```
// Go
cli.RegisterStream(cmd, "training", "Training progress")
w := cli.Channel(cmd, "training")

// TS
registerStream(cmd, "training", "Training progress")
const w = channel(cmd, "training")

// Python
register_stream("train", "training", "Training progress")
w = channel("train", "training")
```

Commands with streams show a STREAMS section in `--help`.
Stream contract defined in `parity.json` under `streams`.

### Adding a new cross-language feature

1. Design in `parity.json` if it involves constants/contracts
2. Implement Go first (primary; parity tests reference Go output)
3. Port to TS (`sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/ts/src/`) and Python (`sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/py/hop_top_kit/`)
4. Share test data via `testdata/` or `parity.json`
5. Add parity test case in `cli/parity_test.go`
6. Run full gate: `go test -tags parity ./cli/...`

### Dependencies

- Go: `cobra`, `viper`, `fang`, `charm.land/*` — see `go.mod`
- TS: `commander`, `pino`, `hono` — see `sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/ts/package.json`
- Python: `typer`, `structlog`, `rich`, `platformdirs` — see
  `sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/py/pyproject.toml`
- New deps: vet with `rsx analyze` before adding

## Build

- `Makefile` at root; `make test`, `make lint`
- Pre-push hook (`.githooks/pre-push`) runs parity + Go + TS +
  Python tests. All must pass before push.
- Worktrees: `pnpm install` + `uv sync` needed per worktree
  (symlinked node_modules can go stale)

## Conventions

- Go module: `hop.top/kit`
- Files <500 LOC; split when approaching
- Commits: Conventional Commits
- Tests: table-driven, `testify/assert` (Go), `vitest` (TS),
  `pytest` (Python)
