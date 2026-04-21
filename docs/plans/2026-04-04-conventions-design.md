# Conventions: Personas, Stories, Testing, Linting

**Scope:** Establish project conventions for personas, features, stories,
testing, and linting across all three kit language targets (Go, TS, Python).

**Applies to:** `hop-top/kit` and, by reference, all hop-top projects.

---

## 1. Personas

Structured persona files capture who uses kit and what they need.

**Directory:** `docs/personas/<slug>.md`

### Frontmatter

```yaml
---
id: <slug>
name: "<display name>"
role: "<one-line role>"
extends: <parent-slug>        # omit for base personas
languages: [go, ts, python]   # which kit targets they touch
---
```

### Body sections

- **Context** -- who they are, what they build
- **Needs** -- what they require from kit
- **Pain points** -- what breaks without kit
- **Success criteria** -- observable proof kit serves them

### Inheritance model

Base personas define broad roles. Language-specific and domain-specific
personas extend them via `extends:`.

```
cli-author (polyglot base)
  +-- go-toolmaker
  +-- ts-toolmaker
  +-- py-toolmaker

oss-contributor (base)
  +-- kit-contributor
```

| Slug | Role | Extends |
|------|------|---------|
| `cli-author` | Builds CLIs with kit in any language | -- |
| `go-toolmaker` | Builds Go CLIs (tlc, rsx, ben, mdl) | `cli-author` |
| `ts-toolmaker` | Builds TS tools (idx, eva-pkg) | `cli-author` |
| `py-toolmaker` | Builds Python tools (eva, eva-ee) | `cli-author` |
| `oss-contributor` | Contributes to any hop-top repo | -- |
| `kit-contributor` | Adds/modifies kit packages | `oss-contributor` |

---

## 2. Features & Stories

Features and stories live in separate directories with a many-to-many
relationship expressed via bidirectional frontmatter links.

### Features

**Directory:** `docs/features/FT-XXXX.md`

```yaml
---
id: FT-0001
title: "<short title>"
status: draft | active | deprecated
stories: [US-0001, US-0003]
personas: [cli-author, go-toolmaker]
created: 2026-04-04
---
```

Body:
- **Summary** -- one paragraph describing the feature
- **Acceptance criteria** -- checklist of observable outcomes

### Stories

**Directory:** `docs/stories/US-XXXX.md`

```yaml
---
id: US-0001
title: "<short title>"
persona: cli-author
features: [FT-0001, FT-0002]
status: draft | ready | implemented | verified
created: 2026-04-04
---
```

Body:
- **As a** `<persona>`, **I want** ..., **so that** ...
- **Acceptance criteria** -- checklist
- **Notes** -- optional context, constraints, edge cases

### ID scheme

Sequential counters per type: `FT-0001`, `US-0001`. Optional date prefix
when time-sensitivity matters: `FT-2026-0404-0001`.

Counters tracked in `docs/features/.counter` and `docs/stories/.counter`
(plain text, last used number).

### Link sync rule

Both sides must list each other. If `FT-0001.stories` includes `US-0003`,
then `US-0003.features` must include `FT-0001`. Keep links in sync
manually; automated enforcement is not yet implemented.

---

## 3. Testing

### Categories

Uniform across Go, TS, and Python:

| Category | Scope | Run target |
|----------|-------|------------|
| Unit | Single function/module | `task test-unit` |
| Integration | Cross-package or external deps | `task test-integration` |
| E2E | CLI invocation, full pipeline | `task test-e2e` |

### File naming (language-native)

| Category | Go | TS | Python |
|----------|----|----|--------|
| Unit | `*_test.go` | `*.test.ts` | `test_*.py` |
| Integration | `*_integration_test.go` | `*.integration.test.ts` | `test_*.py` + `@pytest.mark.integration` |
| E2E | `*_e2e_test.go` | `*.e2e.test.ts` | `test_*.py` + `@pytest.mark.e2e` |

Python uses standard `test_*.py` filenames with pytest markers for
category selection rather than filename conventions.

### Story traceability

Tests validating a story include its ID in the test name:

- Go: `func TestUS0001_ConfigLoadsUserLayer(t *testing.T)`
- TS: `test("US-0001: config loads user layer", ...)`
- Python: `def test_us0001_config_loads_user_layer():`

### Go integration tests

Use `//go:build integration` build tag. `go test ./...` skips them by
default; `task test-integration` passes `-tags integration`.

---

## 4. Linting

### Per-language tools and config

| Language | Tool | Config file | Location |
|----------|------|-------------|----------|
| Go | golangci-lint | `.golangci.yml` | repo root |
| TS | eslint | `eslint.config.mjs` | `sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/ts/` |
| Python | ruff | `ruff.toml` | `sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/py/` |

### Lint profiles

**Go** (golangci-lint): `govet`, `errcheck`, `staticcheck`, `unused`,
`gocritic`, `gofmt` enabled. `wsl`, `nlreturn`, `exhaustruct` disabled.

**TS** (eslint): typescript-eslint recommended + strict. No Prettier.

**Python** (ruff): `select = ["E", "F", "I", "UP", "B", "SIM"]` --
pyflakes, pycodestyle, isort, pyupgrade, bugbear, simplify. Line length
100.

---

## 5. Build System Selection

### Decision rule: Makefile vs Taskfile

Choose per-project based on fit, not habit.

| Choose | When |
|--------|------|
| **Makefile** | Single language; no cross-platform need; file-based targets benefit from Make's dependency graph; team already uses Make |
| **Taskfile** | Multi-language orchestration; cross-platform (Windows); logical task dependencies (not file-based); YAML readability preferred |

### kit: Taskfile

Kit is a multi-language monorepo with logical task orchestration. Make's
file-target model adds no value here.

### Structure

Single `Taskfile.yml` at repo root with all targets. Per-language tasks
use `dir:` to scope to the correct directory (e.g. `dir: ts`, `dir: py`).
Go tasks run from repo root.

### Unified targets

| Target | Effect |
|--------|--------|
| `task lint` | `go:lint`, `ts:lint`, `py:lint` in parallel |
| `task test` | `go:test`, `ts:test`, `py:test` in parallel |
| `task test-unit` | Unit only, all languages |
| `task test-integration` | Integration only, all languages |
| `task check` | `lint` then `test` (sequential, fail fast) |
| `task fix` | Auto-fix all three |

### Per-language targets

| Target | Go | TS | Python |
|--------|----|----|--------|
| `lint` | `golangci-lint run` | `eslint .` | `ruff check .` |
| `fix` | `golangci-lint run --fix` | `eslint --fix .` | `ruff check --fix .` |
| `test` | `go test ./...` | `vitest run` | `pytest` |
| `test-unit` | `go test -short ./...` | `vitest run (exclude integration/e2e)` | `pytest -m 'not integration and not e2e'` |

---

## Directory layout (new additions)

```
Taskfile.yml                            # all targets (lint/test/fix/check)
.golangci.yml                           # Go lint config
docs/
  personas/
    cli-author.md
    go-toolmaker.md
    ts-toolmaker.md
    py-toolmaker.md
    oss-contributor.md
    kit-contributor.md
  features/
    .counter
    FT-0001.md
  stories/
    .counter
    US-0001.md
  plans/
    2026-04-04-conventions-design.md    # this file
sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/ts/
  eslint.config.mjs                     # TS lint config
sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/py/
  ruff.toml                             # Python lint config
```
