# Storage Abstractions

Kit provides 5 storage layers. Each targets a different access
pattern; they compose vertically where needed.

## Package List

| Package       | Type     | Backend                          |
|---------------|----------|----------------------------------|
| `kv/sqlite`   | kv       | Embedded SQLite (default)        |
| `kv/badger`   | kv       | Embedded Badger (high-throughput) |
| `kv/etcd`     | kv       | Distributed etcd cluster         |
| `kv/tidb`     | kv       | TiDB / MySQL-compatible          |
| `blob/local`  | blob     | Local filesystem                 |
| `blob/s3`     | blob     | AWS S3                           |
| `secret/env`  | secret   | Environment variables            |
| `secret/file` | secret   | Encrypted files on disk          |
| `secret/keyring`| secret | OS keychain                      |
| `secret/openbao`| secret | OpenBao / Vault                  |
| `secret/infisical`| secret| Infisical cloud/self-hosted     |
| `secret/memory`| secret  | In-memory (testing)              |

## Layers

### 1. kv.Store -- Key-Value

Interface: `Put` / `Get` / `Delete` / `List` / `Close`

- Values are `[]byte`; caller serializes
- Optional `TTLStore` extension for expiration
- Backend selection via factory:
  ```go
  store, err := kv.Open(kv.Config{
      Backend: "sqlite",   // sqlite | badger | etcd | tidb
      DSN:     "path.db",  // backend-specific connection string
      Table:   "kv",       // table/bucket name (where applicable)
  })
  ```
- Use when: caching, session state, config persistence, sync queue

### 2. blob.Store -- Object/Blob

Interface: `Put` / `Get` / `Delete` / `List` / `Exists`

- Streaming via `io.Reader` / `io.ReadCloser`; no full-buffer
  requirement
- Adapters: `blob/local` (filesystem), `blob/s3` (AWS S3)
- Backup integration: `blob.Store` serves as destination for
  automated backups -- any kv/sqldb data can be snapshotted to a
  blob store via the backup scheduler
- Use when: file storage, backups, large payloads, media

### 3. sqldb -- Shared SQLite Connection

`sqldb.Open()` -- shared connection management (not an interface).

- Opens with standard pragmas (WAL, busy_timeout, foreign_keys)
- Migration helper included
- Packages using sqldb:
  - `domain/sqlite` -- typed repository backend
  - `store` -- DocumentStore (kit serve)
  - `upgrade` -- migration version tracking (local mode)
  - `bus/sqlite` -- cross-process event bus
- Use when: any package needs raw SQL against the local database

### 4. secret.Store -- Secrets

Interface: `Get` / `List` / `Exists` (read-only)
Extended: `MutableStore` adds `Set` / `Delete`

- Values are `*Secret` with `Key`, `Value []byte`, `Metadata`
- Optional `Keeper` interface for encryption at rest
- Backend selection via factory:
  ```go
  store, err := secret.Open(secret.Config{
      Backend: "env",      // env | file | keyring | openbao
      Prefix:  "MYAPP_",   //   | infisical | memory
  })
  ```
- Use when: credentials, API keys, tokens, any sensitive value
- See [Secret Management Guide](secret-management-guide.md)

### 5. domain.Repository[T] -- Typed Entities

Generic CRUD: `Create` / `Get` / `List` / `Update` / `Delete`

- Typed entity operations with validation/auditing
- Backed by `sqldb` under the hood (via `domain/sqlite`)
- Use when: CRUD on domain objects with schema enforcement

## How They Compose

```
App code
  | uses
domain.Repository[T]  <- typed CRUD (highest level)
  | backed by
sqldb.Open()          <- shared SQLite connection

kv.Open(Config)       <- raw key-value (mid level)
  | dispatches to
kv/sqlite, kv/badger, kv/etcd, kv/tidb

blob.Store            <- object storage (files/backups)
  | adapters
blob/local, blob/s3

secret.Open(Config)   <- credentials / sensitive values
  | dispatches to
secret/env, secret/file, secret/keyring,
secret/openbao, secret/infisical, secret/memory

store.DocumentStore   <- kit serve's generic JSON store
  | backed by
sqldb.Open()
```

## Choosing the Right Abstraction

| Need                          | Use                    |
|-------------------------------|------------------------|
| Typed CRUD with validation    | `domain.Repository[T]` |
| Raw bytes by key              | `kv.Open(Config)`      |
| Files / large objects         | `blob.Store`           |
| Automated backups             | `blob.Store` as dest   |
| Credentials / API keys        | `secret.Open(Config)`  |
| Generic JSON documents        | `store.DocumentStore`  |
| Raw SQL (local)               | `sqldb.Open()` direct  |
