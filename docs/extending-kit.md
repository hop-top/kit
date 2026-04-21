# Extending Kit

How to add new packages, ports, and cross-language features.

## Adding a Go package

1. Create a directory at root: `mypkg/`

2. Add `doc.go` with a package-level comment describing purpose
   and listing exported helpers:

   ```go
   // Package mypkg provides ...
   //
   // Helpers:
   //   - foo: does X
   //   - bar: does Y
   package mypkg
   ```

3. Add `example_test.go` with `Example` functions for every
   exported symbol. These serve as both docs and runnable tests:

   ```go
   package mypkg_test

   import (
       "fmt"
       "hop.top/kit/mypkg"
   )

   func ExampleFoo() {
       fmt.Println(mypkg.Foo())
       // Output: bar
   }
   ```

4. Add the package to the README.md table under
   **Packages > Go (`hop.top/kit`)**. Keep alphabetical order.

5. Verify:

   ```
   go test ./mypkg/... -count=1
   golangci-lint run ./mypkg/...
   ```

### Intra-kit dependencies

Keep packages independently importable. If your package needs
another kit package, document the edge in README.md under
"Adding a new package". Avoid circular imports.

## Adding a TypeScript port

1. Create `sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/ts/src/mypkg.ts` (implementation) and
   `sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/ts/src/mypkg.test.ts` (tests).

2. Add a subpath export to `sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/ts/package.json`:

   ```json
   "./mypkg": {
     "require": "./dist/mypkg.js",
     "types": "./dist/mypkg.d.ts"
   }
   ```

3. Ensure the build picks it up:

   ```
   cd ts && pnpm build
   ```

4. Run tests:

   ```
   cd ts && pnpm vitest run src/mypkg.test.ts
   ```

5. If the port mirrors a Go package with cross-language
   contracts, follow the parity.json pattern
   (see [Parity tests](#parity-tests)).

## Adding a Python port

1. Create `sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/py/hop_top_kit/mypkg.py` (implementation) and
   `sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/py/tests/test_mypkg.py` (tests).

2. Verify lint + tests:

   ```
   cd sdk/py && uv run ruff check . && uv run ruff format --check .
   cd sdk/py && uv run pytest tests/test_mypkg.py
   ```

3. If the port mirrors a Go package with cross-language
   contracts, follow the parity.json pattern
   (see [Parity tests](#parity-tests)).

## Conventions

**No unvetted deps.** Run `rsx analyze <owner/repo>` before
adding any external dependency. Block if risky.

**File size.** Keep files under ~500 LOC. Split into subpackages
when a file grows past that.

**TDD.** Write tests first, then implementation.

**doc.go + Examples required.** Every Go package needs a
`doc.go` and an `example_test.go` with `Example` functions
covering exported symbols.

**Doc line width.** Lines under 100 chars in markdown files.

**Build tag for parity.** Parity tests use `//go:build parity`
and run separately from unit tests:

```
go test -tags parity ./cli/... -timeout 300s -count=1
```

## Parity tests

Cross-language features use a shared JSON contract. Example:
`contracts/parity/parity.json` defines constants (spinner frames,
help layout sections, verbosity levels) that Go, TS, and
Python implementations must all load.

To add a parity contract:

1. Create `mypkg/parity/parity.json` with the shared schema.
2. Go: load values in `mypkg/parity/parity.go`, test in
   `mypkg/parity/parity_test.go`.
3. TS: load in `sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/ts/src/mypkg.ts`, test against same JSON.
4. Python: load in `sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/py/hop_top_kit/mypkg.py`, test likewise.
5. The CI parity stage runs after all per-language tests pass.

## CI expectations

CI runs in four stages (see `.github/workflows/ci.yml`):

### Stage 1: Lint

All linters must pass before tests run.

| Target | Command |
|--------|---------|
| Go | `make lint-go` (golangci-lint) |
| TypeScript | `make lint-ts` (eslint) |
| Python | `make lint-py` (ruff check + format) |
| Docs | `make lint-docs` (markdownlint + lychee) |

### Stage 2: Test

Per-language unit tests, gated on lint.

| Target | Command |
|--------|---------|
| Go | `go test ./...` |
| TypeScript | `pnpm vitest run` |
| Python | `uv run pytest` |

### Stage 3: Parity

Runs after all language tests pass:

```
go test -tags parity ./cli/... -timeout 300s
```

### Stage 4: Compliance

Builds the example binary and runs toolspec compliance.

## Pre-push hook

The `.githooks/pre-push` hook runs the full gate locally:

1. `make lint`
2. Parity tests
3. Go tests
4. TS tests
5. Python tests

All must pass before push succeeds.
