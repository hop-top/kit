---
title: "spaced — Unofficial SpaceX CLI Historian"
date: 2026-04-17
author: $USER
track: cli-ux-parity
tasks:
  - Build Go spaced binary (examples/spaced/go/)
  - Build TS spaced script (examples/spaced/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/ts/)
  - Build Python spaced script (examples/spaced/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/py/)
  - Implement shared data layer (missions, vehicles, people, competitors)
  - Implement parity test suite (cli/parity_test.go)
  - Build web demo (examples/spaced/web/)
---

## spaced — Unofficial SpaceX CLI Historian

## Purpose

Dual-role artifact:

1. **Parity test vehicle** — exercises every hop-top CLI contract behavior
   across Go, TS, and Python implementations of `kit/cli`. The parity test
   suite (`cli/parity_test.go`) spawns all three binaries and asserts
   identical contract behavior.

2. **Living example** — developers reference this when building kit-based
   CLIs. Satirical tone makes running tests enjoyable.

> Not affiliated with, endorsed by, or in any way authorized by SpaceX,
> Elon Musk, DOGE, NASA, the FAA, or the Starman mannequin currently
> past Mars. We would, however, accept a sponsorship. Cash, Starlink
> credits, or a ride on the next Crew Dragon all acceptable.
> GitHub Sponsors: <https://github.com/sponsors/hop-top>

---

## Contract Coverage Matrix

Every row is a hop-top CLI contract behavior. Every column is a test.

| Behavior                          | Commands exercised                     |
|-----------------------------------|----------------------------------------|
| `-h/--help` is flag not subcommand | `spaced --help`                       |
| `-v/--version` is flag not subcommand | `spaced --version`                 |
| `--format json/yaml/table`        | `spaced mission list --format json`   |
| `--quiet` suppresses extras       | `spaced launch <m> --quiet`           |
| `--no-color` strips ANSI          | `spaced mission list --no-color`      |
| Boolean flag (`--dry-run`)        | `spaced launch <m> --dry-run`         |
| Single-value flag (`--orbit`)     | `spaced launch <m> --orbit leo`       |
| Comma-list flag (`--payload`)     | `spaced launch <m> --payload cargo,crew` |
| Short flag (`-o`)                 | `spaced launch <m> -o report.json`    |
| Nested subcommands (2-deep)       | `spaced telemetry get <m>`            |
| Nested subcommands (3-deep)       | `spaced fleet vehicle inspect <v>`    |
| Positional argument               | `spaced mission inspect <name>`       |
| Unknown command → stderr + exit 1 | `spaced bogus`                        |
| Unknown arg → stderr + exit 1     | `spaced mission inspect bogus-mission`|
| Error to stderr, not stdout       | all error cases                       |
| No `help` subcommand              | `spaced help` → error                 |
| No `completion` subcommand        | `spaced completion` → error           |
| Version string format             | `spaced --version` → `spaced <ver>`  |

---

## Command Surface

```
spaced [--format] [--quiet] [--no-color] [--version] [--help]

  mission list                         List all missions
  mission inspect <name>               Deep-dive a mission
  mission search --query <q>           Search missions by keyword

  launch <mission>                     Launch a mission
    --payload <p1,p2,...>              Comma-list: cargo, crew, starlink, ...
    --orbit <orbit>                    Single: leo, geo, lunar, helio, tbd
    --dry-run                          Boolean: simulate only
    -o, --output <file>               Short flag: write report to file

  abort <mission>                      Abort a mission
    --reason <string>                  Required reason string

  telemetry get <mission>              Fetch live(ish) telemetry
    --format json|yaml|table          Exercises --format

  countdown <mission>                  T-minus display (spinner/progress)

  fleet list                           List vehicle fleet
  fleet vehicle inspect <name>         Inspect one vehicle
    --systems <s1,s2,...>             Comma-list: propulsion,landing,...

  starship status                      Starship program overview
  starship history                     All integrated test flights

  elon status                          Current Elon Musk status report
  ipo status                           SpaceX IPO watch

  competitor compare <name>            Compare SpaceX vs competitor
    --metric cost,launches,crewed     Comma-list of metrics
```

---

## Data Model

All data hardcoded in each language. No network calls. Works offline.

```
Mission {
  slug        string          // url-safe key e.g. "crew-dragon-demo2"
  name        string          // display name
  vehicle     string          // vehicle slug
  date        string          // ISO8601
  outcome     enum            // SUCCESS | RUD | PARTIAL | SCRUBBED
  orbit       string
  payload     string[]
  crew        string[]
  notes       string          // satirical one-liner
  elon_quote  string          // real or plausible quote
  market_mood string          // emoji + pithy market sentiment
}

Vehicle {
  slug          string
  name          string
  status        string        // active | retired | in_development
  first_flight  string
  launches      int
  reuse_record  string
  description   string        // satirical
  systems       map[string]SystemInfo
}

Competitor {
  slug        string
  name        string
  founded     int
  assessment  string[]        // rotating satirical lines
  metrics     map[string]string
}
```

---

## Randomness / Surprise Mechanic

Selected fields rotate on each run using a **seeded pool**:

- `notes`, `market_mood`, `elon_quote` — pool of 3–5 per mission
- Seed: `time.Now().UnixNano() % len(pool)` (Go), `Date.now() % pool.length` (TS),
  `int(time.time()) % len(pool)` (Python)
- Same mission can yield different commentary run-to-run
- This is intentional: every `spaced mission inspect starman` is a
  shareable screenshot moment

---

## Directory Layout

```
examples/spaced/
  go/
    main.go              # Go binary entry point
    data/
      missions.go        # hardcoded mission data
      vehicles.go        # hardcoded vehicle data
      competitors.go     # hardcoded competitor data
    cmd/
      mission.go         # mission subcommands
      launch.go          # launch command
      abort.go           # abort command
      telemetry.go       # telemetry subcommands
      countdown.go       # countdown command
      fleet.go           # fleet subcommands
      starship.go        # starship subcommands
      elon.go            # elon status
      ipo.go             # ipo status
      competitor.go      # competitor compare
  sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/ts/
    spaced.ts            # TS entry point (Node + browser)
    data.ts              # shared data layer
    commands/            # one file per command group
  sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/py/
    spaced.py            # Python entry point
    data.py              # shared data layer
    commands/            # one file per command group
  web/
    index.html           # browser terminal demo
    terminal.ts          # thin wrapper: routes input → spaced TS commands
    style.css            # terminal aesthetic
```

---

## Parity Constants — Single Source of Truth

**Rule: any value that must be identical across languages lives in `contracts/parity/parity.json`.
No language may hardcode it.**

This includes: status symbols, spinner frames, animation rune set, timing intervals,
default widths, and any other rendering constant that parity depends on.

### File location

```
contracts/parity/parity.json   ← canonical; edit here only
contracts/parity/parity.go     ← Go: go:embed + exported Values var
sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/ts/src/tui/parity.ts     ← TS: readFileSync relative to import.meta.url
sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/py/hop_top_kit/tui/      ← Python: json.loads + pathlib relative to __file__
```

### When adding a new TUI primitive

Before hardcoding any constant, ask: *could another language need to match this?*
If yes — add it to `parity.json` first, then read from it in all language impls.

Default answer: **yes**. Opt out only if the constant is implementation-specific
(e.g., a Rich-only spinner type that has no equivalent elsewhere).

### Tests

`contracts/parity/parity_test.go` — Go tests that pin exact symbol values and validate
schema completeness. Add a test whenever a new key is added to `parity.json`.

---

## Parity Test Suite

Location: `cli/parity_test.go` (Go test, build tag `parity`)

Strategy: `TestMain` builds all three binaries into `t.TempDir()` once,
then each `Test*` spawns all three and asserts same contract behavior.

```
TestMain              builds go/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/ts/py binaries
TestVersion           --version → "spaced <ver>", exit 0, all three
TestHelp              --help → contains name+desc, no subcommand named "help"
TestHelpNoColor       --help --no-color → no ANSI escapes
TestFormat_JSON       mission list --format json → valid JSON array
TestFormat_YAML       mission list --format yaml → valid YAML
TestFormat_Table      mission list → table with headers
TestQuiet             launch <m> --quiet → no spinner/progress lines
TestNoColor           mission list --no-color → no \x1b escapes
TestUnknownCommand    spaced bogus → exit 1, output on stderr
TestUnknownMission    mission inspect bogus → exit 1, stderr message
TestDryRun            launch <m> --dry-run → contains "DRY RUN", exit 0
TestCommaList         launch <m> --payload cargo,crew → both in output
TestShortFlag         launch <m> -o /tmp/out.json → file written
TestPositionalArg     mission inspect starman → contains "Starman"
TestNestedCmd         telemetry get crew-dragon-demo2 → exit 0
TestDeepNestedCmd     fleet vehicle inspect falcon9 → exit 0
TestNoHelpSubcmd      spaced help → exit 1 (not a subcommand)
TestNoCompletionSubcmd spaced completion → exit 1
```

---

## Version String

All three binaries must output identically:

```
spaced 0.1.0
```

Footer (shown in --help only):

```
Not affiliated with, endorsed by, or in any way authorized by SpaceX,
Elon Musk, DOGE, NASA, the FAA, or the Starman mannequin currently past Mars.
We would, however, accept a sponsorship. Cash, Starlink credits, or a ride
on the next Crew Dragon all acceptable. GitHub Sponsors: https://github.com/sponsors/hop-top
```

---

## Web Demo

`examples/spaced/web/` — static HTML + TS bundled with `tsup` or `esbuild`.

- Imports `spaced.ts` command handlers directly (same code as CLI)
- Fake terminal: `<input>` captures keystrokes, output rendered in `<pre>`
- Opening animation: plays curated sequence on load
  (`spaced mission list` → `spaced mission inspect starman` → `spaced elon status`)
- CTA below terminal: `brew install spaced` (or install instructions)
- Footer: the disclaimer, verbatim

No framework. Vanilla TS + CSS. Ships as a single `index.html` + bundled JS.
