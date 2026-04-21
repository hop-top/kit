# CLI Parity Guide

kit enforces identical CLI behaviour across Go, TypeScript,
and Python. Every tool built with kit/cli (or its TS/Py
equivalents) must satisfy the same contract.

## Global Flags

| Flag | Purpose |
|------|---------|
| `-v, --version` | Print `<name> <version>` and exit |
| `-h, --help` | Show help for current command |
| `--format <fmt>` | Output format: table, json, yaml |
| `--quiet` | Suppress non-essential output |
| `--no-color` | Disable ANSI colour |
| `--help-all` | Show help including hidden groups |

## Help Subcommand

No `help` subcommand; only `-h`/`--help` flag. The `help`
command is hidden in all three languages.

## Completion

Disabled or hidden entirely. Tools ship completions via a
separate mechanism (not through the framework's default).

## Error Handling

- Errors print to stderr
- Non-zero exit code on error
- No stack traces in user-facing output

## Command Groups

Commands are organized into named groups. Groups control
how commands appear in `--help` output.

### Default Groups

| Group | ID | Visible | Purpose |
|-------|----|---------|---------|
| COMMANDS | `commands` | Yes | Primary user-facing commands |
| MANAGEMENT | `management` | No | Config, toolspec, diagnostics |

### Assigning Commands to Groups

Developers assign each subcommand to a group at
registration time. Unassigned commands default to the
COMMANDS group.

### Hidden Groups and `--help-all`

Groups with `Hidden: true` are excluded from default
`--help` output. The `--help-all` flag overrides this
filter, revealing all groups and their commands.

### Parity Requirement

All three languages must produce the same group layout:

- Same group IDs and titles
- Same commands in each group
- Same hidden/visible behaviour
- `--help-all` available in all languages

This ensures users see identical help output regardless
of which language a tool is built with.
