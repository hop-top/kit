---
id: secret-consumer
name: "Secret Consumer"
role: "Developer reading secrets in kit-based apps"
extends: go-toolmaker
languages: [go, typescript, python]
---

## Context

Application developer who needs credentials (API keys, DB
passwords, tokens) at runtime. Does not manage the secret
infrastructure -- just reads values through kit's unified API.

## Needs

- Simple `Get(ctx, key)` API; no backend-specific code
- Multiple backends swappable without app changes
- Local dev works without production infra (env vars or files)
- No vendor lock-in; swap OpenBao for Infisical without refactor
- Cross-language access (Go native, TS/Python via kit serve)

## Pain points

- Each secret provider has its own SDK, auth flow, error semantics
- Env vars don't support rotation or metadata
- Testing with real backends is slow and flaky
- 12-factor apps force env-only; can't use encrypted files easily
- No standard way to check if a secret exists before reading

## Success criteria

- 3 lines to read a secret in any supported language
- `memory.New()` in tests; `env.New()` in CI; same app code
- Swap backend via config, zero code change
- `secret.ErrNotFound` is the only error to handle for missing keys
