# Compliance API — 12-Factor AI CLI Checker

Static + runtime checker that validates CLI tools against the
[12-factor AI CLI spec](../README.md).

## Quick Start

### Go

```go
import "hop.top/kit/go/core/compliance"

report, err := compliance.Run(binaryPath, toolspecPath)
fmt.Print(compliance.FormatReport(report, "text"))
```

### TypeScript

```ts
import { run, formatReport } from "@hop-top/kit/compliance";

const report = run(binaryPath, toolspecPath);
console.log(formatReport(report, "text"));
```

### Python

```python
from hop_top_kit.compliance import run, format_report

report = run(binary_path, toolspec_path)
print(format_report(report, "text"))
```

### Via spaced CLI

```bash
spaced compliance              # full check (static + runtime)
spaced compliance --static     # static only
spaced compliance --format json
```

## What Gets Checked

### Static Checks (toolspec YAML)

| # | Factor             | What's checked                              |
|---|--------------------|---------------------------------------------|
| 1 | Self-Describing    | `commands` array non-empty, all named       |
| 2 | Structured I/O     | >= 1 command has `output_schema`             |
| 4 | Contracts & Errors | mutating commands have `contract` fields     |
| 5 | Preview            | mutating commands have `preview_modes`       |
| 6 | Idempotency        | `contract.idempotent` declared               |
| 7 | State Transparency | `state_introspection.config_commands` exists |
| 8 | Safe Delegation    | dangerous commands have `safety` block       |
|11 | Evolution          | `schema_version` is set                      |
|12 | Auth Lifecycle     | `auth_commands` in state_introspection       |

Factors 3 (Stream Discipline), 9 (Observable Ops), 10 (Provenance)
are skipped in static-only mode.

### Runtime Checks (binary execution)

| # | Factor             | What's checked                                |
|---|--------------------|-----------------------------------------------|
| 1 | Self-Describing    | `--help` exits 0, contains COMMANDS/USAGE     |
| 2 | Structured I/O     | read command `--format json` returns valid JSON|
| 3 | Stream Discipline  | stdout has data, stderr has no JSON            |
| 4 | Contracts & Errors | `--bogus-arg` causes non-zero exit             |
| 5 | Preview            | mutating command `--dry-run` exits 0           |
| 7 | State Transparency | `config show` exits 0                          |
| 8 | Safe Delegation    | dangerous commands have safety metadata        |
|10 | Provenance         | JSON output has `_meta` field                  |
|11 | Evolution          | `--version` exits 0                            |
|12 | Auth Lifecycle     | `auth status` exits 0 (or skip if no auth)     |

## CI Integration

```bash
# Fail CI if not fully compliant
spaced compliance --format json | jq -e '.score == 12'

# Or in Go tests
go test ./compliance/... -v
```

## Fixing Failing Factors

Each `CheckResult` includes a `suggestion` field with actionable
fix instructions. Common fixes:

- **F1 Self-Describing**: add `commands` array to toolspec
- **F2 Structured I/O**: add `output_schema` to read commands
- **F4 Contracts**: add `contract.idempotent` + `side_effects`
- **F5 Preview**: add `preview_modes: [--dry-run]`
- **F6 Idempotency**: declare `contract.idempotent: true/false`
- **F7 State Transparency**: add `state_introspection.config_commands`
- **F8 Safe Delegation**: add `safety.requires_confirmation`
- **F11 Evolution**: set `schema_version` in toolspec root
- **F12 Auth Lifecycle**: add `state_introspection.auth_commands`

## API Reference

All three ports expose identical APIs:

- `RunStatic(toolspecPath)` — static checks only
- `RunRuntime(binaryPath, toolspecPath)` — runtime checks only
- `Run(binaryPath, toolspecPath)` — both; empty binary = static only
- `FormatReport(report, format)` — render as "text" or "json"

Status values: `pass`, `fail`, `skip`, `warn`
