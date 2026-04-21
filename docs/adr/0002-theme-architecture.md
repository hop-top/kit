# ADR-0002: Theme Architecture

## Status

Accepted

## Context

hop-top CLIs need consistent visual identity across:

- Help output (commands, flags, titles, arguments)
- Status lines, spinners, error/success messages
- TUI components (bubbletea views, glamour markdown)

Requirements:

- Brand accent per tool (hex override) without forking
  the theme system
- Perceptually balanced semantic colors (error, muted,
  success) that work on light and dark terminals
- Named palettes for quick switching (dev vs production
  aesthetics)
- Cross-language parity: TS and Python tools use the same
  hex values via `contracts/parity/parity.json`

## Decision

**CharmTone for semantics + brand Palette for identity.**

### Palette struct

Two named palettes ship with kit:

| Palette | Command (primary) | Flag (secondary) |
|---------|-------------------|------------------|
| Neon    | `#7ED957` grass   | `#FF00FF` pink   |
| Dark    | `#C1FF72` lime    | `#FF66C4` pink   |

`Palette` holds two `color.Color` values: `Command` and
`Flag`. These drive fang's `ColorScheme` for help rendering
and are exposed on `Theme.Accent` / `Theme.Secondary`.

### CharmTone

`charmbracelet/x/exp/charmtone` provides perceptually
balanced named colors. kit uses three:

- `charmtone.Squid` — muted/subtle text
- `charmtone.Cherry` — error
- `charmtone.Guac` — success

CharmTone handles light/dark terminal adaptation internally.
kit does not maintain its own WCAG contrast logic.

### Theme struct

Built by `buildTheme(accent string)`:

1. Start with Neon palette
2. If `Config.Accent` is non-empty, override `Command` color
3. Derive semantic colors from CharmTone
4. Build pre-built lipgloss styles: `Title` (bold white),
   `Subtle` (muted), `Bold`

Injected once at `Root` level — all subcommands access
`root.Theme` without re-building.

### Brand color scheme

`brandColorScheme` passes Neon palette colors into fang's
`ColorScheme`:

- `cs.Title` = white
- `cs.Command` / `cs.Program` = Neon.Command
- `cs.Flag` = Neon.Flag
- `cs.Argument` = `#B5E89B`
- `cs.DimmedArgument` = `#8ABF6E`

Registered via `fang.WithColorSchemeFunc(brandColorScheme)`
in `Execute`.

### Cross-language parity

`contracts/parity/parity.json` defines section order, status
symbols, spinner frames, verbosity mapping. Theme hex
values are code-level constants today; a `colors` key in
`parity.json` is the planned next step for full TS/Python
color parity.

## Consequences

### Enables

- New tool overrides accent via `Config{Accent: "#..."}`;
  all help, status, TUI output adapts automatically
- Neon/Dark palettes available for direct use outside the
  theme system (e.g. `cli.Neon.Command` in custom renderers)
- CharmTone semantic colors are terminal-profile-aware —
  no per-tool contrast tuning needed
- Single `Theme` on `Root` — no global state, testable with
  different accents in parallel

### Constrains

- Only two brand colors per palette (Command + Flag); adding
  a third requires `Palette` struct change
- CharmTone is experimental (`x/exp/charmtone`); API may
  shift with Charm upstream
- `brandColorScheme` hardcodes Neon values for fang — does
  not respect `Config.Accent` for flag/argument colors (only
  `Theme.Accent` reflects the override)
- Color parity with TS/Python is manual until `parity.json`
  gains a `colors` section
