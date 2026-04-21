---
id: security-operator
name: "Security Operator"
role: "Configures and manages secret backends for teams"
extends: platform-engineer
languages: [go]
---

## Context

DevOps/security engineer responsible for secret lifecycle:
provisioning backends, rotating credentials, enforcing access
policies, ensuring encryption at rest, maintaining audit trails.

## Needs

- Encryption at rest via Keeper interface (NaCl, KMS, etc.)
- Rotation support without app restarts
- Audit trail: who accessed what, when
- Access control: scope secrets per service/environment
- Compliance: FIPS, SOC2 evidence for secret storage
- Backend selection based on threat model (local vs. distributed)

## Pain points

- Teams hardcode secrets in env files committed to repos
- No unified rotation mechanism across providers
- Audit requires stitching logs from multiple systems
- Encrypted file secrets lack key management story
- Migrating between backends requires app code changes

## Success criteria

- `Keeper` encrypts all file-backed secrets transparently
- Backend swap via `secret.Open(Config)` -- no team code changes
- Access patterns visible in structured logs
- Rotation triggers re-fetch; apps never cache stale values
- Single config surface for all environments (dev/staging/prod)
