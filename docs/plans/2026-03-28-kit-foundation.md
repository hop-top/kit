# hop-top/kit Foundation Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan
> task-by-task.

**Goal:** Build `hop-top/kit` — a shared multi-language library providing config loading, XDG path
resolution, SQLite storage, output rendering, upgrade checking, and opinionated CLI setup for the
hop-top family of tools (tlc, rsx, ben, eva, idx, xray, ctxt, etc.).

**Architecture:** Monorepo following the `hop-top/upgrade` pattern — one repo, three language
roots (`go/`, `sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/ts/`, `sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/py/`), each published independently. No language root imports another.
Every sub-package is independently importable. No tool-specific logic leaks in. Tools adopt
incrementally by swapping imports on their own schedule.

**Published as:**
- Go: `hop.top/kit` (e.g. `hop.top/kit/xdg`, `hop.top/kit/cli`)
- TypeScript/CJS: `@hop/kit-ts` (subpath exports: `@hop/kit-sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/ts/cli`, `@hop/kit-sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/ts/output`)
- ESM: `@hop/kit` (subpath exports: `@hop/kit/cli`, `@hop/kit/output`)
- Python: `hop-kit` (submodules: `hop_kit.cli`, `hop_kit.output`)

**Tech Stack:**
- Go: stdlib, `modernc.org/sqlite`, `gopkg.in/yaml.v3`, `github.com/spf13/cobra`,
  `github.com/spf13/viper`, `github.com/stretchr/testify`
- TypeScript: Commander v11, `tsup` for bundling, `vitest` for tests
- Python: Typer, Click, `pytest`

---

## Repo structure

```
kit/
  hops/
    main/                    ← Go module root (hop.top/kit)
      go.mod
      justfile
      xdg/
      config/
      sqlstore/
      output/
      upgrade/
      cli/
    sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/ts/                      ← TypeScript/CJS (@hop/kit-ts)
      package.json
      tsconfig.json
      src/
        cli.ts
        output.ts
    js/                      ← ESM (@hop/kit)
      package.json
      cli.js
      output.js
    sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/py/                      ← Python (hop-kit)
      pyproject.toml
      hop_kit/
        cli.py
        output.py
```

---

## What kit provides (v0.1 scope)

| Package | Go | TS | Python | What it replaces |
|---------|----|----|--------|-----------------|
| `xdg` | ✓ | — | — | Per-tool XDG path logic in tlc, rsx, ben |
| `config` | ✓ | — | — | system→user→project→env loader in tlc, rsx |
| `sqlstore` | ✓ | — | — | SQLite open+migrate+CRUD in rsx, tlc, ben |
| `output` | ✓ | ✓ | ✓ | table/json/yaml renderer in all CLIs |
| `upgrade` | ✓ | — | — | re-export of `hop.top/upgrade` |
| `cli` | ✓ | ✓ | ✓ | cobra+viper / Commander / Typer setup |

Out of scope for v0.1: auth, plugin host, TUI primitives, registry client.

---

## Consistent CLI behaviour (all languages)

Every hop-top CLI built with `kit/cli` must behave identically:

| Behaviour | Detail |
|-----------|--------|
| No `help` subcommand | Hidden; only `-h` / `--help` flag at any level |
| No `completion` subcommand | Hidden or omitted entirely |
| `-v` / `--version` | Root-only flag; not a subcommand; prints `<tool> <version>` to stdout |
| `--format` global flag | `table` (default) / `json` / `yaml` |
| `--quiet` global flag | Suppress non-essential output |
| `--no-color` global flag | Disable ANSI colour |
| Error messages | To stderr; non-zero exit; no stack traces |
| `--help` on unknown cmd | Print help for nearest parent command |

---

## Task 1: Repo bootstrap (Go)

**Files:**
- Create: `hops/main/go.mod`
- Create: `hops/main/justfile`
- Create: `hops/main/.gitignore`

**Step 1: Initialise worktree and module**

```bash
cd ~/.w/ideacrafterslabs/kit
/usr/bin/git hop add main
cd hops/main
go mod init hop.top/kit
go get github.com/stretchr/testify@latest
```

Expected: `go.mod` with `module hop.top/kit`.

**Step 2: Create justfile**

```makefile
test:
    go test ./...

lint:
    go vet ./...

check: lint test
```

**Step 3: Create .gitignore**

```
*.test
*.out
/dist/
```

**Step 4: Commit**

```bash
git add go.mod go.sum justfile .gitignore
git commit -m "chore: bootstrap hop.top/kit Go module"
```

---

## Task 2: `kit/xdg` — XDG path resolution (Go)

**Files:**
- Create: `hops/main/xdg/xdg.go`
- Create: `hops/main/xdg/xdg_test.go`

**Context:** tlc `internal/config/paths.go` and rsx `internal/config/config.go` both implement
the same XDG fallback logic independently. This package generalises it: caller passes the tool
name, gets back the right OS-native path.

**Step 1: Write the failing tests**

```go
// xdg/xdg_test.go
package xdg_test

import (
    "os"
    "path/filepath"
    "strings"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "hop.top/kit/go/core/xdg"
)

func TestConfigDir_XDGOverride(t *testing.T) {
    t.Setenv("XDG_CONFIG_HOME", "/tmp/xdg-cfg")
    dir, err := xdg.ConfigDir("mytool")
    require.NoError(t, err)
    assert.Equal(t, "/tmp/xdg-cfg/mytool", dir)
}

func TestDataDir_XDGOverride(t *testing.T) {
    t.Setenv("XDG_DATA_HOME", "/tmp/xdg-data")
    dir, err := xdg.DataDir("mytool")
    require.NoError(t, err)
    assert.Equal(t, "/tmp/xdg-data/mytool", dir)
}

func TestCacheDir_XDGOverride(t *testing.T) {
    t.Setenv("XDG_CACHE_HOME", "/tmp/xdg-cache")
    dir, err := xdg.CacheDir("mytool")
    require.NoError(t, err)
    assert.Equal(t, "/tmp/xdg-cache/mytool", dir)
}

func TestStateDir_XDGOverride(t *testing.T) {
    t.Setenv("XDG_STATE_HOME", "/tmp/xdg-state")
    dir, err := xdg.StateDir("mytool")
    require.NoError(t, err)
    assert.Equal(t, "/tmp/xdg-state/mytool", dir)
}

func TestConfigDir_FallbackContainsTool(t *testing.T) {
    os.Unsetenv("XDG_CONFIG_HOME")
    dir, err := xdg.ConfigDir("mytool")
    require.NoError(t, err)
    assert.True(t, strings.HasSuffix(dir, filepath.Join("mytool")),
        "expected path to end with tool name, got: %s", dir)
}
```

**Step 2: Run tests to verify they fail**

```bash
go test ./xdg/... -v
```

Expected: FAIL — package not found.

**Step 3: Implement `xdg/xdg.go`**

```go
package xdg

import (
    "fmt"
    "os"
    "path/filepath"
    "runtime"
)

func ConfigDir(tool string) (string, error) {
    if v := os.Getenv("XDG_CONFIG_HOME"); v != "" {
        return filepath.Join(v, tool), nil
    }
    dir, err := os.UserConfigDir()
    if err != nil {
        return "", fmt.Errorf("resolve config dir: %w", err)
    }
    return filepath.Join(dir, tool), nil
}

func DataDir(tool string) (string, error) {
    if v := os.Getenv("XDG_DATA_HOME"); v != "" {
        return filepath.Join(v, tool), nil
    }
    home, err := os.UserHomeDir()
    if err != nil {
        return "", fmt.Errorf("resolve home dir: %w", err)
    }
    switch runtime.GOOS {
    case "darwin":
        return filepath.Join(home, "Library", "Application Support", tool), nil
    case "windows":
        if local := os.Getenv("LocalAppData"); local != "" {
            return filepath.Join(local, tool), nil
        }
        return "", fmt.Errorf("%%LocalAppData%% not set")
    default:
        return filepath.Join(home, ".local", "share", tool), nil
    }
}

func CacheDir(tool string) (string, error) {
    if v := os.Getenv("XDG_CACHE_HOME"); v != "" {
        return filepath.Join(v, tool), nil
    }
    dir, err := os.UserCacheDir()
    if err != nil {
        return "", fmt.Errorf("resolve cache dir: %w", err)
    }
    return filepath.Join(dir, tool), nil
}

func StateDir(tool string) (string, error) {
    if v := os.Getenv("XDG_STATE_HOME"); v != "" {
        return filepath.Join(v, tool), nil
    }
    home, err := os.UserHomeDir()
    if err != nil {
        return "", fmt.Errorf("resolve home dir: %w", err)
    }
    switch runtime.GOOS {
    case "darwin":
        return filepath.Join(home, "Library", "Application Support", tool, "state"), nil
    case "windows":
        if local := os.Getenv("LocalAppData"); local != "" {
            return filepath.Join(local, tool, "state"), nil
        }
        return "", fmt.Errorf("%%LocalAppData%% not set")
    default:
        return filepath.Join(home, ".local", "state", tool), nil
    }
}

// MustEnsure returns dir after calling os.MkdirAll; panics on any error.
func MustEnsure(dir string, err error) string {
    if err != nil {
        panic(err)
    }
    if err := os.MkdirAll(dir, 0o750); err != nil {
        panic(err)
    }
    return dir
}
```

**Step 4: Run tests to verify they pass**

```bash
go test ./xdg/... -v
```

Expected: PASS (5 tests).

**Step 5: Commit**

```bash
git add xdg/
git commit -m "feat(xdg): add XDG path resolution"
```

---

## Task 3: `kit/config` — layered config loader (Go)

**Files:**
- Create: `hops/main/config/loader.go`
- Create: `hops/main/config/loader_test.go`

**Context:** tlc and rsx each independently implement system→user→project→env config loading.
This package provides that as a generic loader; callers supply their own struct + env-override fn.

**Step 1: Write the failing tests**

```go
// config/loader_test.go
package config_test

import (
    "os"
    "path/filepath"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "gopkg.in/yaml.v3"
    "hop.top/kit/go/core/config"
)

type testCfg struct {
    Name  string `yaml:"name"`
    Debug bool   `yaml:"debug"`
    Port  int    `yaml:"port"`
}

func writeYAML(t *testing.T, dir, filename string, v any) string {
    t.Helper()
    require.NoError(t, os.MkdirAll(dir, 0o750))
    data, err := yaml.Marshal(v)
    require.NoError(t, err)
    path := filepath.Join(dir, filename)
    require.NoError(t, os.WriteFile(path, data, 0o644))
    return path
}

func TestLoader_UserOverridesDefault(t *testing.T) {
    userDir := t.TempDir()
    writeYAML(t, userDir, "config.yaml", map[string]any{"name": "user-name"})
    var cfg testCfg
    cfg.Name = "default"
    cfg.Port = 8080
    err := config.Load(&cfg, config.Options{
        UserConfigPath: filepath.Join(userDir, "config.yaml"),
    })
    require.NoError(t, err)
    assert.Equal(t, "user-name", cfg.Name)
    assert.Equal(t, 8080, cfg.Port)
}

func TestLoader_ProjectOverridesUser(t *testing.T) {
    userDir, projDir := t.TempDir(), t.TempDir()
    writeYAML(t, userDir, "config.yaml", map[string]any{"name": "user"})
    writeYAML(t, projDir, "config.yaml", map[string]any{"name": "project"})
    var cfg testCfg
    err := config.Load(&cfg, config.Options{
        UserConfigPath:    filepath.Join(userDir, "config.yaml"),
        ProjectConfigPath: filepath.Join(projDir, "config.yaml"),
    })
    require.NoError(t, err)
    assert.Equal(t, "project", cfg.Name)
}

func TestLoader_EnvOverride(t *testing.T) {
    t.Setenv("MY_DEBUG", "true")
    var cfg testCfg
    err := config.Load(&cfg, config.Options{
        EnvOverride: func(c any) {
            if os.Getenv("MY_DEBUG") == "true" {
                c.(*testCfg).Debug = true
            }
        },
    })
    require.NoError(t, err)
    assert.True(t, cfg.Debug)
}

func TestLoader_MissingFilesAreSkipped(t *testing.T) {
    var cfg testCfg
    cfg.Name = "default"
    err := config.Load(&cfg, config.Options{
        UserConfigPath:    "/nonexistent/config.yaml",
        ProjectConfigPath: "/also/nonexistent.yaml",
    })
    require.NoError(t, err)
    assert.Equal(t, "default", cfg.Name)
}
```

**Step 2: Run tests to verify they fail**

```bash
go test ./config/... -v
```

Expected: FAIL.

**Step 3: Add yaml dependency**

```bash
go get gopkg.in/yaml.v3
```

**Step 4: Implement `config/loader.go`**

```go
package config

import (
    "errors"
    "fmt"
    "os"

    "gopkg.in/yaml.v3"
)

type Options struct {
    SystemConfigPath  string
    UserConfigPath    string
    ProjectConfigPath string
    EnvOverride       func(cfg any)
}

func Load(dst any, opts Options) error {
    for _, path := range []string{
        opts.SystemConfigPath, opts.UserConfigPath, opts.ProjectConfigPath,
    } {
        if path == "" {
            continue
        }
        if err := mergeFile(dst, path); err != nil && !errors.Is(err, os.ErrNotExist) {
            return fmt.Errorf("load config %s: %w", path, err)
        }
    }
    if opts.EnvOverride != nil {
        opts.EnvOverride(dst)
    }
    return nil
}

func mergeFile(dst any, path string) error {
    data, err := os.ReadFile(path)
    if err != nil {
        return err
    }
    return yaml.Unmarshal(data, dst)
}
```

**Step 5: Run tests to verify they pass**

```bash
go test ./config/... -v
```

Expected: PASS (4 tests).

**Step 6: Commit**

```bash
git add config/
git commit -m "feat(config): add layered config loader"
```

---

## Task 4: `kit/sqlstore` — SQLite store (Go)

**Files:**
- Create: `hops/main/sqlstore/store.go`
- Create: `hops/main/sqlstore/store_test.go`

**Context:** rsx `internal/cache/sqlite.go` has the cleanest version of this pattern.
Generalised here: open+migrate, generic JSON kv store, optional TTL, `DB()` escape hatch
for tool-specific queries.

**Step 1: Write failing tests**

```go
// sqlstore/store_test.go
package sqlstore_test

import (
    "context"
    "path/filepath"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "hop.top/kit/go/storage/sqlstore"
)

func TestStore_PutGet(t *testing.T) {
    s, err := sqlstore.Open(filepath.Join(t.TempDir(), "test.db"), sqlstore.Options{})
    require.NoError(t, err)
    defer s.Close()
    ctx := context.Background()
    require.NoError(t, s.Put(ctx, "key1", map[string]string{"hello": "world"}))
    var out map[string]string
    found, err := s.Get(ctx, "key1", &out)
    require.NoError(t, err)
    assert.True(t, found)
    assert.Equal(t, "world", out["hello"])
}

func TestStore_MissingKey(t *testing.T) {
    s, err := sqlstore.Open(filepath.Join(t.TempDir(), "test.db"), sqlstore.Options{})
    require.NoError(t, err)
    defer s.Close()
    var out map[string]string
    found, err := s.Get(context.Background(), "missing", &out)
    require.NoError(t, err)
    assert.False(t, found)
}

func TestStore_TTLExpiry(t *testing.T) {
    s, err := sqlstore.Open(filepath.Join(t.TempDir(), "test.db"), sqlstore.Options{
        TTL: 10 * time.Millisecond,
    })
    require.NoError(t, err)
    defer s.Close()
    ctx := context.Background()
    require.NoError(t, s.Put(ctx, "ttl-key", "value"))
    time.Sleep(20 * time.Millisecond)
    var out string
    found, err := s.Get(ctx, "ttl-key", &out)
    require.NoError(t, err)
    assert.False(t, found)
}

func TestStore_PutOverwrites(t *testing.T) {
    s, err := sqlstore.Open(filepath.Join(t.TempDir(), "test.db"), sqlstore.Options{})
    require.NoError(t, err)
    defer s.Close()
    ctx := context.Background()
    require.NoError(t, s.Put(ctx, "k", "first"))
    require.NoError(t, s.Put(ctx, "k", "second"))
    var out string
    found, err := s.Get(ctx, "k", &out)
    require.NoError(t, err)
    assert.True(t, found)
    assert.Equal(t, "second", out)
}
```

**Step 2: Run tests to verify they fail**

```bash
go test ./sqlstore/... -v
```

Expected: FAIL.

**Step 3: Add SQLite dependency**

```bash
go get modernc.org/sqlite
```

**Step 4: Implement `sqlstore/store.go`**

```go
package sqlstore

import (
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "time"

    _ "modernc.org/sqlite"
)

type Options struct {
    TTL        time.Duration
    MigrateSQL string
}

type Store struct {
    db   *sql.DB
    opts Options
}

func Open(path string, opts Options) (*Store, error) {
    if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
        return nil, fmt.Errorf("create db dir: %w", err)
    }
    db, err := sql.Open("sqlite", path)
    if err != nil {
        return nil, fmt.Errorf("open sqlite: %w", err)
    }
    s := &Store{db: db, opts: opts}
    if err := s.migrate(); err != nil {
        return nil, fmt.Errorf("migrate: %w", err)
    }
    return s, nil
}

func (s *Store) migrate() error {
    _, err := s.db.Exec(`create table if not exists kv (
      key       text primary key,
      stored_at text not null,
      payload   text not null
    );`)
    if err != nil {
        return err
    }
    if s.opts.MigrateSQL != "" {
        _, err = s.db.Exec(s.opts.MigrateSQL)
    }
    return err
}

func (s *Store) Put(ctx context.Context, key string, v any) error {
    data, err := json.Marshal(v)
    if err != nil {
        return fmt.Errorf("marshal: %w", err)
    }
    _, err = s.db.ExecContext(ctx,
        `insert into kv (key, stored_at, payload) values (?, ?, ?)
         on conflict(key) do update set stored_at=excluded.stored_at, payload=excluded.payload`,
        key, time.Now().UTC().Format(time.RFC3339), string(data))
    return err
}

func (s *Store) Get(ctx context.Context, key string, dst any) (bool, error) {
    var storedAt, payload string
    err := s.db.QueryRowContext(ctx,
        `select stored_at, payload from kv where key = ?`, key).Scan(&storedAt, &payload)
    if err == sql.ErrNoRows {
        return false, nil
    }
    if err != nil {
        return false, err
    }
    if s.opts.TTL > 0 {
        ts, err := time.Parse(time.RFC3339, storedAt)
        if err != nil || time.Since(ts) > s.opts.TTL {
            return false, nil
        }
    }
    return true, json.Unmarshal([]byte(payload), dst)
}

func (s *Store) DB() *sql.DB  { return s.db }
func (s *Store) Close() error  { return s.db.Close() }
```

**Step 5: Run tests to verify they pass**

```bash
go test ./sqlstore/... -v
```

Expected: PASS (4 tests).

**Step 6: Commit**

```bash
git add sqlstore/
git commit -m "feat(sqlstore): add generic SQLite kv store with TTL"
```

---

## Task 5: `kit/output` — table/json/yaml renderer (Go)

**Files:**
- Create: `hops/main/output/renderer.go`
- Create: `hops/main/output/renderer_test.go`

**Step 1: Write failing tests**

```go
// output/renderer_test.go
package output_test

import (
    "bytes"
    "encoding/json"
    "strings"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "gopkg.in/yaml.v3"
    "hop.top/kit/go/console/output"
)

type row struct {
    Name  string `json:"name"  yaml:"name"  table:"Name"`
    Score int    `json:"score" yaml:"score" table:"Score"`
}

func TestRender_JSON(t *testing.T) {
    var buf bytes.Buffer
    require.NoError(t, output.Render(&buf, output.JSON, row{Name: "xray", Score: 95}))
    var got row
    require.NoError(t, json.Unmarshal(buf.Bytes(), &got))
    assert.Equal(t, "xray", got.Name)
}

func TestRender_YAML(t *testing.T) {
    var buf bytes.Buffer
    require.NoError(t, output.Render(&buf, output.YAML, row{Name: "grep", Score: 42}))
    var got row
    require.NoError(t, yaml.Unmarshal(buf.Bytes(), &got))
    assert.Equal(t, "grep", got.Name)
}

func TestRender_Table_ContainsValues(t *testing.T) {
    var buf bytes.Buffer
    require.NoError(t, output.Render(&buf, output.Table, []row{
        {Name: "xray", Score: 95},
        {Name: "grep", Score: 40},
    }))
    out := buf.String()
    assert.True(t, strings.Contains(out, "xray"))
    assert.True(t, strings.Contains(out, "95"))
}

func TestRender_UnknownFormat(t *testing.T) {
    var buf bytes.Buffer
    assert.Error(t, output.Render(&buf, "xml", row{}))
}
```

**Step 2: Run tests to verify they fail**

```bash
go test ./output/... -v
```

**Step 3: Implement `output/renderer.go`**

```go
package output

import (
    "encoding/json"
    "fmt"
    "io"
    "reflect"
    "strings"
    "text/tabwriter"

    "gopkg.in/yaml.v3"
)

type Format = string

const (
    JSON  Format = "json"
    YAML  Format = "yaml"
    Table Format = "table"
)

func Render(w io.Writer, format Format, v any) error {
    switch format {
    case JSON:
        enc := json.NewEncoder(w)
        enc.SetIndent("", "  ")
        return enc.Encode(v)
    case YAML:
        return yaml.NewEncoder(w).Encode(v)
    case Table:
        return renderTable(w, v)
    default:
        return fmt.Errorf("unknown output format %q (valid: json, yaml, table)", format)
    }
}

func renderTable(w io.Writer, v any) error {
    tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
    defer tw.Flush()
    rv := reflect.ValueOf(v)
    if rv.Kind() == reflect.Slice {
        if rv.Len() == 0 {
            return nil
        }
        headers, _ := tableFields(rv.Index(0).Type())
        fmt.Fprintln(tw, strings.Join(headers, "\t"))
        for i := range rv.Len() {
            _, idxs := tableFields(rv.Index(i).Type())
            row := make([]string, len(idxs))
            for j, idx := range idxs {
                row[j] = fmt.Sprintf("%v", rv.Index(i).Field(idx))
            }
            fmt.Fprintln(tw, strings.Join(row, "\t"))
        }
        return nil
    }
    headers, idxs := tableFields(rv.Type())
    fmt.Fprintln(tw, strings.Join(headers, "\t"))
    row := make([]string, len(idxs))
    for j, idx := range idxs {
        row[j] = fmt.Sprintf("%v", rv.Field(idx))
    }
    fmt.Fprintln(tw, strings.Join(row, "\t"))
    return nil
}

func tableFields(t reflect.Type) (headers []string, indices []int) {
    if t.Kind() == reflect.Ptr {
        t = t.Elem()
    }
    for i := range t.NumField() {
        if tag := t.Field(i).Tag.Get("table"); tag != "" && tag != "-" {
            headers = append(headers, tag)
            indices = append(indices, i)
        }
    }
    return
}
```

**Step 4: Run tests to verify they pass**

```bash
go test ./output/... -v
```

Expected: PASS (4 tests).

**Step 5: Commit**

```bash
git add output/
git commit -m "feat(output): add table/json/yaml renderer"
```

---

## Task 6: `kit/upgrade` — re-export hop.top/upgrade (Go)

**Files:**
- Create: `hops/main/upgrade/upgrade.go`

**Step 1: Add dependency and check API**

```bash
go get hop.top/upgrade@latest
go doc hop.top/upgrade   # note exported symbols
```

**Step 2: Implement `upgrade/upgrade.go`**

```go
// Package upgrade re-exports hop.top/upgrade for consistent kit imports.
package upgrade

import upstream "hop.top/upgrade"

// Check is re-exported from hop.top/upgrade. Safe to call at CLI startup.
var Check = upstream.Check
```

**Step 3: Build and commit**

```bash
go build ./upgrade/...
git add upgrade/
git commit -m "feat(upgrade): re-export hop.top/upgrade"
```

---

## Task 7: `kit/cli` — opinionated CLI setup (Go, cobra+viper)

**Files:**
- Create: `hops/main/cli/cli.go`
- Create: `hops/main/cli/cli_test.go`

**Context:** Every hop-top Go CLI uses cobra+viper but each wires it differently.
`kit/cli` provides `New(cfg Config) *cobra.Command` that returns a root command
pre-configured with the standard hop-top CLI contract: no help/completion subcommands,
`-h/--help` flag only, `-v/--version` root-only flag, global `--format/--quiet/--no-color`.

**Step 1: Add dependencies**

```bash
go get github.com/spf13/cobra@latest
go get github.com/spf13/viper@latest
```

**Step 2: Write failing tests**

```go
// cli/cli_test.go
package cli_test

import (
    "bytes"
    "strings"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "hop.top/kit/go/console/cli"
)

func root() *cli.Root {
    return cli.New(cli.Config{
        Name:    "mytool",
        Version: "1.2.3",
        Short:   "A test tool",
    })
}

func TestCLI_VersionFlag(t *testing.T) {
    r := root()
    var buf bytes.Buffer
    r.Cmd.SetOut(&buf)
    r.Cmd.SetArgs([]string{"-v"})
    err := r.Cmd.Execute()
    require.NoError(t, err)
    assert.Contains(t, buf.String(), "mytool 1.2.3")
}

func TestCLI_NoHelpSubcommand(t *testing.T) {
    r := root()
    for _, sub := range r.Cmd.Commands() {
        assert.NotEqual(t, "help", sub.Name(),
            "help must not appear as a subcommand")
        assert.NotEqual(t, "completion", sub.Name(),
            "completion must not appear as a subcommand")
    }
}

func TestCLI_HelpFlagExists(t *testing.T) {
    r := root()
    f := r.Cmd.Flags().Lookup("help")
    require.NotNil(t, f, "expected -h/--help flag on root command")
}

func TestCLI_GlobalFlagsExist(t *testing.T) {
    r := root()
    pf := r.Cmd.PersistentFlags()
    assert.NotNil(t, pf.Lookup("format"))
    assert.NotNil(t, pf.Lookup("quiet"))
    assert.NotNil(t, pf.Lookup("no-color"))
}

func TestCLI_UnknownCommand_ShowsHelp(t *testing.T) {
    r := root()
    var buf bytes.Buffer
    r.Cmd.SetErr(&buf)
    r.Cmd.SetArgs([]string{"unknowncmd"})
    _ = r.Cmd.Execute()
    assert.True(t, strings.Contains(buf.String(), "unknown command") ||
        strings.Contains(buf.String(), "Usage:"),
        "expected help or error on unknown command")
}
```

**Step 3: Run tests to verify they fail**

```bash
go test ./cli/... -v
```

Expected: FAIL — package not found.

**Step 4: Implement `cli/cli.go`**

```go
package cli

import (
    "fmt"

    "github.com/spf13/cobra"
    "github.com/spf13/viper"
)

// Config holds the tool identity for root command construction.
type Config struct {
    Name    string // binary name, e.g. "tlc"
    Version string // semver, e.g. "1.2.3"
    Short   string // one-line description for --help
}

// Root wraps the cobra root command and viper instance.
type Root struct {
    Cmd    *cobra.Command
    Viper  *viper.Viper
    Config Config
}

// New returns a Root pre-configured to the hop-top CLI contract:
//   - no help or completion subcommands (only -h/--help flag)
//   - -v/--version root-only flag (prints "<name> <version>")
//   - persistent global flags: --format, --quiet, --no-color
//   - errors go to stderr; usage printed on error
func New(cfg Config) *Root {
    v := viper.New()

    cmd := &cobra.Command{
        Use:   cfg.Name,
        Short: cfg.Short,
        // SilenceUsage prevents cobra printing usage on every error;
        // we print it only for unknown commands (handled below).
        SilenceUsage:  true,
        SilenceErrors: true,
    }

    // Hide the default help command; -h/--help flag remains.
    cmd.SetHelpCommand(&cobra.Command{Hidden: true})

    // Disable completion subcommand entirely.
    cmd.CompletionOptions.DisableDefaultCmd = true

    // -v / --version root-only flag.
    var showVersion bool
    cmd.Flags().BoolVarP(&showVersion, "version", "v", false,
        fmt.Sprintf("Print %s version and exit", cfg.Name))
    cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
        if showVersion {
            fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", cfg.Name, cfg.Version)
            // Signal callers to exit 0 cleanly.
            return cobra.ErrSubCommandRequired // reused as sentinel; handled in Run
        }
        return nil
    }
    // Execute version print before any run logic.
    origPersistentPreRunE := cmd.PersistentPreRunE
    cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
        if f := cmd.Root().Flags().Lookup("version"); f != nil && f.Value.String() == "true" {
            fmt.Fprintf(cmd.Root().OutOrStdout(), "%s %s\n", cfg.Name, cfg.Version)
            return nil
        }
        if origPersistentPreRunE != nil {
            return origPersistentPreRunE(cmd, args)
        }
        return nil
    }

    // Global persistent flags bound to viper.
    pf := cmd.PersistentFlags()
    pf.String("format", "table", "Output format: table, json, yaml")
    pf.Bool("quiet", false, "Suppress non-essential output")
    pf.Bool("no-color", false, "Disable ANSI colour")
    _ = v.BindPFlag("format", pf.Lookup("format"))
    _ = v.BindPFlag("quiet", pf.Lookup("quiet"))
    _ = v.BindPFlag("no-color", pf.Lookup("no-color"))

    return &Root{Cmd: cmd, Viper: v, Config: cfg}
}
```

**Step 5: Run tests to verify they pass**

```bash
go test ./cli/... -v
```

Expected: PASS (5 tests).

**Step 6: Commit**

```bash
git add cli/
git commit -m "feat(cli): add opinionated cobra+viper CLI setup"
```

---

## Task 8: `@hop/kit-sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/ts/cli` — Commander setup (TypeScript)

**Files:**
- Create: `hops/main/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/ts/package.json`
- Create: `hops/main/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/ts/tsconfig.json`
- Create: `hops/main/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/ts/src/cli.ts`
- Create: `hops/main/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/ts/src/cli.test.ts`

**Context:** Mirrors `@hop/upgrade-sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/ts/commander` pattern already established in
`upgrade/hops/main/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/ts/src/commander.ts`. `createCLI()` returns a pre-configured
Commander program with the same hop-top contract as the Go `cli.New()`.

**Step 1: Bootstrap TS package**

```bash
cd hops/main/ts
```

Create `package.json`:

```json
{
  "name": "@hop/kit-ts",
  "version": "0.1.0",
  "description": "Shared CLI utilities for hop-top tools — TypeScript/CJS edition",
  "type": "commonjs",
  "main": "./dist/index.js",
  "types": "./dist/index.d.ts",
  "exports": {
    ".": {
      "require": "./dist/index.js",
      "types": "./dist/index.d.ts"
    },
    "./cli": {
      "require": "./dist/cli.js",
      "types": "./dist/cli.d.ts"
    },
    "./output": {
      "require": "./dist/output.js",
      "types": "./dist/output.d.ts"
    }
  },
  "scripts": {
    "build": "tsup src/cli.ts src/output.ts --format cjs --dts",
    "test": "vitest run"
  },
  "devDependencies": {
    "@types/node": "^20",
    "commander": "^11",
    "tsup": "^8",
    "typescript": "^5",
    "vitest": "^1"
  },
  "peerDependencies": {
    "commander": ">=11"
  },
  "engines": { "node": ">=20" },
  "license": "MIT"
}
```

Create `tsconfig.json`:

```json
{
  "compilerOptions": {
    "target": "ES2020",
    "module": "CommonJS",
    "declaration": true,
    "outDir": "./dist",
    "strict": true,
    "esModuleInterop": true
  },
  "include": ["src"]
}
```

```bash
pnpm install
```

**Step 2: Write failing tests**

```typescript
// src/cli.test.ts
import { describe, it, expect } from 'vitest';
import { createCLI } from './cli';

describe('createCLI', () => {
  it('exposes -v/--version flag, not a command', () => {
    const program = createCLI({ name: 'mytool', version: '1.2.3', description: 'A tool' });
    const versionOpt = program.opts();
    // version is an option on root, not a subcommand
    const subNames = program.commands.map(c => c.name());
    expect(subNames).not.toContain('version');
    expect(subNames).not.toContain('help');
    expect(subNames).not.toContain('completion');
  });

  it('has --format, --quiet, --no-color global options', () => {
    const program = createCLI({ name: 'mytool', version: '1.2.3', description: 'A tool' });
    const optNames = program.options.map(o => o.long);
    expect(optNames).toContain('--format');
    expect(optNames).toContain('--quiet');
    expect(optNames).toContain('--no-color');
  });

  it('--format defaults to table', () => {
    const program = createCLI({ name: 'mytool', version: '1.2.3', description: 'A tool' });
    program.parse(['node', 'mytool']);
    expect(program.opts().format).toBe('table');
  });
});
```

**Step 3: Run tests to verify they fail**

```bash
pnpm test
```

Expected: FAIL — `./cli` not found.

**Step 4: Implement `src/cli.ts`**

```typescript
import { Command } from 'commander';

export interface CLIConfig {
  name: string;
  version: string;
  description: string;
}

/**
 * Creates a Commander program pre-configured to the hop-top CLI contract:
 * - No help or completion subcommands (only -h/--help flag)
 * - -v/--version root-only option (not a subcommand)
 * - Global options: --format, --quiet, --no-color
 * - Errors to stderr; usage shown on unknown command
 */
export function createCLI(cfg: CLIConfig): Command {
  const program = new Command(cfg.name)
    .description(cfg.description)
    .version(cfg.version, '-v, --version', `Print ${cfg.name} version and exit`)
    .helpOption('-h, --help', 'Display help')
    // Hide the default help command; -h/--help flag remains.
    .addHelpCommand(false)
    .option('--format <fmt>', 'Output format: table, json, yaml', 'table')
    .option('--quiet', 'Suppress non-essential output', false)
    .option('--no-color', 'Disable ANSI colour', false)
    .showHelpAfterError(true);

  return program;
}
```

**Step 5: Run tests to verify they pass**

```bash
pnpm test
```

Expected: PASS (3 tests).

**Step 6: Build and commit**

```bash
pnpm build
cd ../..   # back to hops/main
git add sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/ts/
git commit -m "feat(cli/ts): add Commander CLI setup for @hop/kit-ts"
```

---

## Task 9: `hop_kit.cli` — Typer setup (Python)

**Files:**
- Create: `hops/main/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/py/pyproject.toml`
- Create: `hops/main/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/py/hop_kit/__init__.py`
- Create: `hops/main/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/py/hop_kit/cli.py`
- Create: `hops/main/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/py/tests/test_cli.py`

**Context:** eva and eva-ee are Python. `hop_kit.cli` provides `create_app()` returning
a Typer app pre-configured to the hop-top contract. Typer hides its own `--install-completion`
and `--show-completion` flags; `kit/cli` also disables them and adds the standard global options.

**Step 1: Bootstrap Python package**

Create `pyproject.toml`:

```toml
[build-system]
requires = ["hatchling"]
build-backend = "hatchling.build"

[project]
name = "hop-kit"
version = "0.1.0"
description = "Shared CLI utilities for hop-top tools — Python edition"
requires-python = ">=3.11"
dependencies = ["typer>=0.12"]

[project.optional-dependencies]
dev = ["pytest>=8", "typer[all]>=0.12"]

[tool.hatch.build.targets.wheel]
packages = ["hop_kit"]
```

```bash
cd hops/main/py
pip install -e ".[dev]"
```

**Step 2: Write failing tests**

```python
# tests/test_cli.py
from typer.testing import CliRunner
from hop_kit.cli import create_app

runner = CliRunner()

def test_version_flag():
    app = create_app(name="mytool", version="1.2.3", help="A tool")
    result = runner.invoke(app, ["--version"])
    assert result.exit_code == 0
    assert "mytool 1.2.3" in result.output

def test_no_help_subcommand():
    app = create_app(name="mytool", version="1.2.3", help="A tool")
    # Typer exposes commands; none should be named 'help' or 'completion'
    from typer.main import get_command
    cmd = get_command(app)
    subcommand_names = list(cmd.commands.keys()) if hasattr(cmd, 'commands') else []
    assert "help" not in subcommand_names
    assert "completion" not in subcommand_names

def test_format_default():
    app = create_app(name="mytool", version="1.2.3", help="A tool")

    @app.command()
    def run(format: str = "table"):
        print(format)

    result = runner.invoke(app, ["run"])
    assert "table" in result.output
```

**Step 3: Run tests to verify they fail**

```bash
pytest tests/ -v
```

Expected: FAIL — `hop_kit.cli` not found.

**Step 4: Implement `hop_kit/cli.py`**

```python
"""
hop_kit.cli — opinionated Typer app factory for hop-top tools.

Usage:
    from hop_kit.cli import create_app
    app = create_app(name="mytool", version="1.2.3", help="Does things")

    @app.command()
    def run(format: str = "table", quiet: bool = False):
        ...
"""
from __future__ import annotations

import typer
from typing import Optional


def create_app(
    *,
    name: str,
    version: str,
    help: str,
) -> typer.Typer:
    """
    Return a Typer app pre-configured to the hop-top CLI contract:
    - --version root-only flag (not a subcommand); prints '<name> <version>'
    - No install-completion / show-completion flags
    - --format / --quiet / --no-color available as callback params for subcommands
    """
    app = typer.Typer(
        name=name,
        help=help,
        add_completion=False,   # disables --install-completion / --show-completion
        no_args_is_help=True,
    )

    @app.callback(invoke_without_command=True)
    def _root(
        ctx: typer.Context,
        ver: Optional[bool] = typer.Option(
            None, "-v", "--version",
            help=f"Print {name} version and exit",
            is_eager=True,
        ),
    ) -> None:
        if ver:
            typer.echo(f"{name} {version}")
            raise typer.Exit()
        if ctx.invoked_subcommand is None:
            typer.echo(ctx.get_help())
            raise typer.Exit()

    return app
```

**Step 5: Run tests to verify they pass**

```bash
pytest tests/ -v
```

Expected: PASS (3 tests).

**Step 6: Commit**

```bash
cd ../..
git add sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/py/
git commit -m "feat(cli/py): add Typer CLI setup for hop-kit Python"
```

---

## Task 10: README, full test gate, and v0.1.0 tag

**Files:**
- Create: `hops/main/README.md`

**Step 1: Update justfile for all languages**

```makefile
test-go:
    go test ./...

test-ts:
    cd ts && pnpm test

test-py:
    cd sdk/py && pytest tests/ -v

test: test-go test-ts test-py

lint-go:
    go vet ./...

check: lint-go test
```

**Step 2: Run full gate**

```bash
just check
```

Expected: all Go, TS, Python tests pass.

**Step 3: Write README.md**

```markdown
# hop.top/kit

Shared multi-language library for the hop-top family of tools.
Follows the same monorepo pattern as `hop-top/upgrade`.

## Packages

### Go (`hop.top/kit`)
| Package | Purpose |
|---------|---------|
| `kit/xdg` | XDG path resolution |
| `kit/config` | Layered config loader (system→user→project→env) |
| `kit/sqlstore` | Generic SQLite kv store with TTL |
| `kit/output` | table/json/yaml renderer |
| `kit/upgrade` | Re-export of hop.top/upgrade |
| `kit/cli` | cobra+viper root command factory |

### TypeScript (`@hop/kit-ts`)
| Subpath | Purpose |
|---------|---------|
| `@hop/kit-sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/ts/cli` | Commander program factory |
| `@hop/kit-sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/ts/output` | table/json/yaml renderer |

### Python (`hop-kit`)
| Module | Purpose |
|--------|---------|
| `hop_kit.cli` | Typer app factory |
| `hop_kit.output` | table/json/yaml renderer |

## CLI contract (all languages)

Every tool built with `kit/cli` must behave identically:
- `-h/--help` flag only; no `help` subcommand
- No `completion` subcommand
- `-v/--version` root-only flag; prints `<name> <version>`
- Global: `--format table|json|yaml`, `--quiet`, `--no-color`

## Adding a new package

Keep packages small and independently importable.
No intra-kit dependencies (no kit package imports another kit package).
```

**Step 4: Commit and tag**

```bash
git add README.md justfile
git commit -m "docs: README and full test gate for hop.top/kit v0.1"
git tag v0.1.0
```

---

## Adoption guide

After v0.1.0 is tagged, tools adopt incrementally:

**Go tools** (tlc, rsx, ben, mdl, etc.):
```bash
go get hop.top/kit@v0.1.0
```
- Replace local XDG path fns → `xdg.ConfigDir("toolname")` etc.
- Replace local config loader → `config.Load(&cfg, config.Options{...})`
- Replace local SQLite cache → `sqlstore.Open(path, opts)`
- Replace local output renderer → `output.Render(w, format, v)`
- Replace cobra+viper wiring → `cli.New(cli.Config{...})`

**TS tools** (eva-pkg, idx):
```bash
pnpm add @hop/kit-ts
```
- Replace Commander setup → `createCLI({name, version, description})`

**Python tools** (eva, eva-ee):
```bash
pip install hop-kit
```
- Replace Typer setup → `create_app(name=..., version=..., help=...)`
