#!/usr/bin/env bats
# Tests for .githooks/pre-push scoped lint/test logic.
#
# Validates that the hook correctly detects changed languages and
# constructs the right make/go targets. Does NOT run actual builds.
#
# Run: bats .github/tests/pre-push-hook.bats
# Or:  make test-hook

HOOK=".githooks/pre-push"

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

# Extract the language-detection block and evaluate it against a fake
# CHANGED variable, printing which HAS_* vars are set.
detect_languages() {
    local changed="$1"
    CHANGED="$changed"
    HAS_GO=$(echo "$CHANGED" | grep -E '\.go$' || true)
    HAS_TS=$(echo "$CHANGED" | grep -E '^sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/ts/' || true)
    HAS_PY=$(echo "$CHANGED" | grep -E '^sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/py/' || true)
    HAS_DOCS=$(echo "$CHANGED" | grep -E '\.(md)$' || true)
    HAS_CLI=$(echo "$CHANGED" | grep -E '^cli/' || true)
}

# Extract Go packages from changed files (mirrors hook logic).
go_pkgs_from() {
    local changed="$1"
    echo "$changed" | grep -E '\.go$' \
        | xargs -I{} dirname {} \
        | sort -u \
        | sed 's|^|./|' \
        | tr '\n' ' '
}

# ---------------------------------------------------------------------------
# Structure
# ---------------------------------------------------------------------------

@test "hook file exists and is executable" {
    [ -f "$HOOK" ]
    [ -x "$HOOK" ]
}

@test "hook is valid sh syntax" {
    bash -n "$HOOK"
}

@test "hook reads stdin (remote ref protocol)" {
    # pre-push hooks receive lines on stdin; the script must read them.
    grep -q 'while read' "$HOOK"
}

@test "hook has SHA cache skip logic" {
    grep -q 'pre-push-last-sha' "$HOOK"
}

# ---------------------------------------------------------------------------
# Language detection
# ---------------------------------------------------------------------------

@test "detect: Go only" {
    detect_languages "bus/bus.go
bus/bus_test.go"
    [ -n "$HAS_GO" ]
    [ -z "$HAS_TS" ]
    [ -z "$HAS_PY" ]
    [ -z "$HAS_DOCS" ]
}

@test "detect: TypeScript only" {
    detect_languages "sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/ts/src/cli.ts
sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/ts/src/cli.test.ts"
    [ -z "$HAS_GO" ]
    [ -n "$HAS_TS" ]
    [ -z "$HAS_PY" ]
    [ -z "$HAS_DOCS" ]
}

@test "detect: Python only" {
    detect_languages "sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/py/hop_top_kit/cli.py
sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/py/tests/test_cli.py"
    [ -z "$HAS_GO" ]
    [ -z "$HAS_TS" ]
    [ -n "$HAS_PY" ]
    [ -z "$HAS_DOCS" ]
}

@test "detect: docs only" {
    detect_languages "README.md
docs/plans/foo.md"
    [ -z "$HAS_GO" ]
    [ -z "$HAS_TS" ]
    [ -z "$HAS_PY" ]
    [ -n "$HAS_DOCS" ]
}

@test "detect: mixed Go + TS + docs" {
    detect_languages "bus/bus.go
sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/ts/src/bus.ts
CHANGELOG.md"
    [ -n "$HAS_GO" ]
    [ -n "$HAS_TS" ]
    [ -z "$HAS_PY" ]
    [ -n "$HAS_DOCS" ]
}

@test "detect: CLI triggers parity" {
    detect_languages "cli/root.go
cli/completion/complete.go"
    [ -n "$HAS_GO" ]
    [ -n "$HAS_CLI" ]
}

@test "detect: non-CLI Go does not trigger parity" {
    detect_languages "bus/bus.go
kv/tidb/tidb.go"
    [ -n "$HAS_GO" ]
    [ -z "$HAS_CLI" ]
}

# ---------------------------------------------------------------------------
# Go package extraction
# ---------------------------------------------------------------------------

@test "go_pkgs: single file" {
    result=$(go_pkgs_from "bus/bus.go")
    [[ "$result" == *"./bus"* ]]
}

@test "go_pkgs: multiple files same package" {
    result=$(go_pkgs_from "bus/bus.go
bus/bus_test.go")
    # Should deduplicate to single ./bus
    [ "$(echo "$result" | xargs -n1 | wc -l | tr -d ' ')" = "1" ]
    [[ "$result" == *"./bus"* ]]
}

@test "go_pkgs: multiple packages" {
    result=$(go_pkgs_from "bus/bus.go
kv/tidb/tidb.go
identity/jwt.go")
    [[ "$result" == *"./bus"* ]]
    [[ "$result" == *"./kv/tidb"* ]]
    [[ "$result" == *"./identity"* ]]
}

@test "go_pkgs: nested paths" {
    result=$(go_pkgs_from "llm/router/controller.go
secret/file/file.go")
    [[ "$result" == *"./llm/router"* ]]
    [[ "$result" == *"./secret/file"* ]]
}

@test "go_pkgs: ignores non-go files" {
    result=$(go_pkgs_from "README.md
sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/ts/src/cli.ts
bus/bus.go")
    [[ "$result" == *"./bus"* ]]
    [[ "$result" != *"./ts"* ]]
    [[ "$result" != *"README"* ]]
}
