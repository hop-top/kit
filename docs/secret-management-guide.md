# Secret Management Guide

## Overview

`secret/` provides a unified interface for reading and writing
application secrets (API keys, DB passwords, tokens) across
multiple backends. One import, swap providers via config.

Why it exists:

- Every secret provider has its own SDK, auth, error model
- Apps couple to a single provider; migration = rewrite
- Testing with real backends is slow, flaky, expensive
- kit's `secret.Store` interface eliminates all three problems

## Quick Start

```go
import "hop.top/kit/go/storage/secret/env"

store := env.New("MYAPP_")
s, err := store.Get(ctx, "db_password")
if err != nil { ... }
fmt.Println(string(s.Value))
```

Set `MYAPP_DB_PASSWORD=hunter2` in your shell; done.

## Backends

| Backend   | Package            | Use when                         |
|-----------|--------------------|----------------------------------|
| env       | `secret/env`       | Local dev, CI, 12-factor apps    |
| file      | `secret/file`      | Encrypted files, SOPS-style      |
| keyring   | `secret/keyring`   | Desktop apps, CLI tools           |
| openbao   | `secret/openbao`   | Production, team secrets          |
| infisical | `secret/infisical` | Cloud-native, SaaS teams          |
| memory    | `secret/memory`    | Testing                           |

### env

Reads `PREFIX + UPPER(key)` from environment. Slashes become
underscores: key `db/password` reads `MYAPP_DB_PASSWORD`.

```go
store := env.New("MYAPP_")
```

### file

Reads/writes secrets as individual files under a directory.
Supports optional encryption via `Keeper`.

```go
store := file.New("/etc/myapp/secrets", nil)       // plaintext
store := file.New("/etc/myapp/secrets", keeper)    // encrypted
```

### keyring

OS keychain (macOS Keychain, Windows Credential Manager,
Linux Secret Service). Good for CLI tools storing user tokens.

```go
store := keyring.New("myapp")
```

Note: `List` returns `ErrNotSupported` (OS limitation).

### openbao

HashiCorp Vault-compatible (OpenBao fork). Production use with
ACLs, audit, rotation.

```go
store := openbao.New("https://vault:8200", token, "secret")
```

### infisical

Cloud-hosted or self-hosted secret manager with REST API.

```go
store := infisical.New(baseURL, token, projectID, "production")
```

### memory

In-process map. Use in tests to avoid I/O or network.

```go
store := memory.New()
_ = store.Set(ctx, "api_key", []byte("test-value"))
```

## Encryption at Rest

The `file` adapter accepts a `Keeper` for transparent
encrypt-on-write / decrypt-on-read:

```go
import (
    "hop.top/kit/go/storage/secret/file"
    "hop.top/kit/go/storage/secret/local"
    "hop.top/kit/go/core/identity"
)

kp, _ := identity.LoadKeypair("~/.config/myapp/key")
keeper := local.NewKeeper(kp)
store := file.New("/etc/myapp/secrets", keeper)

// Writes NaCl secretbox-encrypted file
_ = store.Set(ctx, "db_password", []byte("hunter2"))

// Reads + decrypts transparently
s, _ := store.Get(ctx, "db_password")
```

Keeper interface (implement for KMS, age, etc.):

```go
type Keeper interface {
    Encrypt(ctx context.Context, plaintext []byte) ([]byte, error)
    Decrypt(ctx context.Context, ciphertext []byte) ([]byte, error)
}
```

## Configuration

`secret.Open(Config)` is the factory for config-driven backend
selection:

```go
import "hop.top/kit/go/storage/secret"

store, err := secret.Open(secret.Config{
    Backend: "env",
    Prefix:  "MYAPP_",
})
```

Config fields per backend:

| Field     | env | file | keyring | openbao | infisical |
|-----------|-----|------|---------|---------|-----------|
| Prefix    | x   |      |         |         |           |
| Dir       |     | x    |         |         |           |
| Service   |     |      | x       |         |           |
| Addr      |     |      |         | x       | x         |
| Token     |     |      |         | x       | x         |
| Mount     |     |      |         | x       |           |
| Project   |     |      |         |         | x         |
| Env       |     |      |         |         | x         |

Backends self-register via `secret.RegisterBackend`. Import the
adapter package for side-effect registration:

```go
import _ "hop.top/kit/go/storage/secret/env"  // registers "env" backend
```

## Testing

Use `memory` adapter for deterministic, fast tests:

```go
func TestMyService(t *testing.T) {
    store := memory.New()
    _ = store.Set(ctx, "api_key", []byte("fake-key"))

    svc := myservice.New(store)
    // ... assertions
}
```

For recorded integration tests, pair with xrr cassettes:

```go
func TestInfisicalIntegration(t *testing.T) {
    // Record mode: hits real Infisical, saves responses
    // Replay mode: serves saved responses, no network
    srv := xrr.Replay(t, "testdata/cassettes/infisical")
    store := infisical.New(srv.URL, "tok", "proj", "dev")

    s, err := store.Get(ctx, "db_password")
    assert.NoError(t, err)
    assert.Equal(t, "expected-value", string(s.Value))
}
```

## Cross-Language Access

TS/Python apps access secrets via kit's gRPC/HTTP serve layer:

```
GET /api/v1/secrets/{key}
Authorization: Bearer <service-token>
```

Response:

```json
{"key": "db_password", "value": "aHVudGVyMg=="}
```

Value is base64-encoded. Backend selection configured server-side;
clients are backend-agnostic.
