# ADR-0003: Output and Hints Pipeline

## Status

Accepted

## Context

CLIs serve two audiences simultaneously: humans in terminals
and machines consuming structured data. Mixing help text,
progress indicators, and next-step suggestions into the primary
data stream breaks both `jq` pipelines and shell scripts.

Prior art (docker, gh, kubectl) splits human chrome from
machine data, but the split varies per tool. kit needs a
single convention enforced across Go, TypeScript, and Python
implementations via the shared `cli` and `output` packages.

Key tensions:

- Beginners need guidance; power users need silence.
- Structured output (JSON/YAML) must be grep/jq-safe.
- Debug channels must exist without polluting either stream.
- Cross-language parity: same flags, same behavior.

## Decision

### Stream separation

stdout carries data exclusively. stderr carries everything
else: hints, logs, progress, diagnostics. This follows POSIX
convention and lets `cmd | jq` work without filtering.

Implementation: `cli.Channel()` returns an `io.Writer` bound
to stderr with a `[name]` prefix when the named stream is
enabled via `--stream`; otherwise returns `io.Discard`.

### Data format (--format)

`--format table|json|yaml` controls data serialization on
stdout. Default is `table` (human-aligned columns via
`text/tabwriter`). Table rendering uses `table:""` struct
tags; JSON/YAML pass through standard encoders.

`output.RegisterFlags()` adds the flag and binds to viper's
`"format"` key. `output.Render()` dispatches by format.

### Hints (--no-hints)

Hints are contextual next-step suggestions rendered after
primary output (e.g. "Run `hop version` to verify.").

`output.HintSet` is a concurrency-safe registry mapping
command names to `Hint` structs. Each hint carries an optional
`Condition` func gating relevance at render time.

`output.RenderHints()` writes active hints to stderr only
when ALL conditions hold:

1. Format is `table` (not JSON/YAML).
2. `--no-hints` flag is false.
3. `hints.enabled` config key is not false.
4. `HOP_QUIET_HINTS` env is not truthy.
5. `--quiet` flag is false.
6. Output target is a TTY.

This progressive-disclosure stack lets each layer suppress
without affecting the others:
- CI sets `HOP_QUIET_HINTS=1` globally.
- Scripts pass `--no-hints` per invocation.
- Config file disables persistently.
- Pipe detection auto-suppresses.

### Named streams (--stream)

`cli.RegisterStream()` declares a named debug channel on a
command. `--stream fetch,cache` enables those channels. Each
enabled stream writes `[name] ...` prefixed lines to stderr.
Disabled streams route to `io.Discard` (zero allocation).

Streams appear in a `STREAMS` help section appended to the
command's usage template.

### Verbosity (-V)

Stackable count flag: `-V` = debug, `-VV` = trace. Stored as
`Root.verboseCount`. `--quiet` overrides all verbosity.

### Noise-reduction flags

| Flag | Scope | Effect |
|-------------|----------|--------------------------------------|
| `--format`  | data     | serialization shape                  |
| `--no-hints`| hints    | suppress next-step suggestions       |
| `--quiet`   | all      | suppress non-essential output         |
| `--no-color`| display  | disable ANSI escapes                 |
| `-V` / `-VV`| logging  | increase log detail                  |
| `--stream`  | debug    | enable named diagnostic channels     |

### Cross-language parity

All kit CLIs (Go, TS, Python) implement identical flags with
identical semantics. The Go `cli` and `output` packages are
the reference; TS/Python ports mirror the flag names, defaults,
env vars, and suppression rules.

## Consequences

### Enables

- **Scriptability**: stdout is always machine-safe; no grep
  filtering required.
- **Progressive disclosure**: new users see hints; automation
  sees nothing extra.
- **Granular debug**: named streams avoid log-level blasting;
  enable only the channels you need.
- **Consistent UX**: one flag vocabulary across all kit CLIs
  regardless of implementation language.

### Constrains

- Commands must never write human text to stdout. All chrome
  (hints, progress, banners) goes to stderr.
- New output formats require updating `output.Render()` and
  all language ports simultaneously.
- Hint conditions run at render time; expensive checks must
  be pre-computed and stored behind a `*bool` flag (see
  `RegisterUpgradeHints` pattern).
- Table rendering depends on `table:""` struct tags; adding
  a column means updating the struct, not a template.
