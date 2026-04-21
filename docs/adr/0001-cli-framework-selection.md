# ADR-0001: CLI Framework Selection

## Status

Accepted

## Context

kit provides a shared Go CLI layer for all hop-top tools.
Requirements:

- Subcommand trees with git-style dispatch
- Shell completion (bash/zsh/fish/powershell)
- Styled help, errors, man pages (brand-consistent)
- Config file + env var + flag binding with precedence
- Cross-language parity: TS (commander) and Python
  (click/typer) implementations mirror the same contract
  via `contracts/parity/parity.json`

Candidates evaluated:

| Framework     | Strengths              | Gap for kit            |
|---------------|------------------------|------------------------|
| cobra         | Industry standard      | No styled output       |
| urfave/cli    | Simpler API            | No fang integration    |
| kong          | Struct tags            | Smaller ecosystem      |
| raw flag/pflag| Zero deps              | Everything hand-rolled |

## Decision

**cobra + fang + viper** — three libraries, each owning one
concern:

### cobra (github.com/spf13/cobra)

Root command factory (`cli.New`) builds a `cobra.Command`
with subcommand groups, hidden management commands,
`--help-all` / `--help-<group>` flags, and alias annotation.
Cobra owns dispatch, arg validation, and completion generation.

Chosen because:

- De facto Go CLI standard; patterns translate directly to
  commander (TS) and click (Python)
- Built-in shell completion subcommand — placed in management
  group, hidden from default `--help`
- Subcommand groups (`cobra.Group`) used for COMMANDS vs
  MANAGEMENT vs custom sections
- `PersistentFlags` + `PersistentPreRunE` chain cleanly with
  functional options (`WithAPI`, `WithIdentity`, `WithPeers`)

### fang (charm.land/fang/v2)

`fang.Execute` wraps cobra's `Execute`, injecting:

- Styled help/error rendering via `fang.ColorScheme`
- Version output (`fang.WithVersion`)
- Brand color scheme override (`fang.WithColorSchemeFunc`)
- Man page generation

kit's `brandColorScheme` overrides fang defaults with Neon
palette colors for commands, flags, arguments.

Chosen because:

- Only library that bolts styled output onto cobra without
  replacing it
- Charm ecosystem alignment (lipgloss, bubbletea, glamour
  already in `go.mod`)
- Color scheme function pattern lets kit inject brand palette
  at root level

### viper (github.com/spf13/viper)

Each `Root` holds a dedicated `viper.Viper` instance. Flags
registered on cobra are bound via `v.BindPFlag`:

- `--quiet`, `--no-color`, `--format`, `--no-hints`
- Tool-specific globals from `Config.Globals`
- Stackable `--verbose` / `-V` count flag

Chosen because:

- Canonical cobra companion; flag-to-config binding is
  one-liner per flag
- Supports config file, env var, and flag precedence out
  of the box
- Per-instance (not global singleton) via `viper.New()` —
  safe for testing and multi-root scenarios

## Consequences

### Enables

- New tool = `cli.New(cfg)` + add subcommands; zero
  boilerplate for help, completion, version, color
- `Config.Disable` suppresses built-in flags per tool;
  `Config.Globals` adds tool-specific flags — both without
  forking the factory
- `parity.json` section order and symbols enforced across
  Go/TS/Python; Go relies on fang defaults, others mirror
- Functional options (`WithAPI`, `WithIdentity`, `WithPeers`)
  compose cleanly via cobra's `PersistentPreRunE` chain

### Constrains

- Help layout locked to fang's template; custom section
  rendering requires fang upstream changes
- Three transitive dep trees (cobra, fang, viper) — acceptable
  given Charm ecosystem overlap
- cobra's completion subcommand must be placed post-init
  (`InitDefaultCompletionCmd` in `Execute`) to land in the
  correct group
- viper global state avoided by using `viper.New()`, but
  third-party libs that call `viper.Get*` won't see kit's
  instance
