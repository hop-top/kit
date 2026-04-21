# Engine Protocol Reference

This protocol is what makes kit apps language-agnostic peers.
A Go app using kit natively and a TS app using kit serve speak
the SAME sync/peer/bus wire protocol. This spec IS the interop
contract.

## Conventions

- Base URL: `http://localhost:<port>` (port from engine stdout)
- Content-Type: `application/json` for all request/response bodies
- Error format (all non-2xx responses):

```json
{"status": 404, "code": "not_found", "message": "document not found"}
```

| Status | Meaning             |
|--------|---------------------|
| 200    | OK                  |
| 201    | Created             |
| 204    | No Content (delete) |
| 400    | Bad Request         |
| 404    | Not Found           |
| 409    | Conflict            |
| 500    | Internal Error      |

---

## Documents

### Create Document

```
POST /:type/
```

**Request:**

| Header       | Value              |
|--------------|--------------------|
| Content-Type | application/json   |

```json
{
  "id": "string (optional, auto-generated if omitted)",
  "data": {}
}
```

**Response (201):**

```json
{
  "type": "string",
  "id": "string",
  "data": {},
  "created_at": "RFC3339",
  "updated_at": "RFC3339"
}
```

**curl:**

```sh
curl -X POST http://localhost:9090/notes/ \
  -H 'Content-Type: application/json' \
  -d '{"data":{"title":"Hello","body":"world"}}'
```

---

### List Documents

```
GET /:type/?limit=N&offset=N&sort=field&search=term
```

**Query params (all optional):**

| Param  | Type   | Default | Notes               |
|--------|--------|---------|---------------------|
| limit  | int    | 100     | max items returned  |
| offset | int    | 0       | pagination offset   |
| sort   | string | -updated_at | field, `-` = desc |
| search | string | —       | full-text search    |

**Response (200):**

```json
{
  "items": [
    {"type":"notes","id":"abc","data":{},"created_at":"...","updated_at":"..."}
  ],
  "total": 42
}
```

**curl:**

```sh
curl http://localhost:9090/notes/?limit=10&offset=0
```

---

### Get Document

```
GET /:type/:id
```

**Response (200):**

```json
{
  "type": "notes",
  "id": "abc",
  "data": {"title": "Hello"},
  "created_at": "2026-04-19T10:00:00Z",
  "updated_at": "2026-04-19T10:05:00Z"
}
```

**404** if not found.

**curl:**

```sh
curl http://localhost:9090/notes/abc
```

---

### Update Document

```
PUT /:type/:id
```

**Request:**

```json
{
  "data": {"title": "Updated"}
}
```

**Response (200):** full document with new `updated_at`.

**409** if concurrent write detected (optimistic locking).

**curl:**

```sh
curl -X PUT http://localhost:9090/notes/abc \
  -H 'Content-Type: application/json' \
  -d '{"data":{"title":"Updated"}}'
```

---

### Delete Document

```
DELETE /:type/:id
```

**Response:** 204 No Content.

**404** if not found.

**curl:**

```sh
curl -X DELETE http://localhost:9090/notes/abc
```

---

### Document History

```
GET /:type/:id/history
```

Returns version list (newest first).

**Response (200):**

```json
{
  "versions": [
    {
      "version": 3,
      "data": {},
      "timestamp": "RFC3339",
      "operation": "update"
    }
  ]
}
```

**curl:**

```sh
curl http://localhost:9090/notes/abc/history
```

---

### Revert Document

```
POST /:type/:id/revert
```

**Request:**

```json
{"version": 2}
```

**Response (200):** document at reverted state.

**409** if version does not exist.

**curl:**

```sh
curl -X POST http://localhost:9090/notes/abc/revert \
  -H 'Content-Type: application/json' \
  -d '{"version":2}'
```

---

## Sync

### Add Remote

```
POST /sync/remotes
```

**Request:**

```json
{
  "name": "string (required)",
  "url": "string (required, peer base URL)",
  "mode": "push | pull | both",
  "filter": "string (optional, entity type glob)"
}
```

**Response (201):**

```json
{
  "name": "peer-b",
  "url": "http://192.168.1.50:8080",
  "mode": "both",
  "filter": ""
}
```

**409** if name already exists.

**curl:**

```sh
curl -X POST http://localhost:9090/sync/remotes \
  -H 'Content-Type: application/json' \
  -d '{"name":"peer-b","url":"http://192.168.1.50:8080","mode":"both"}'
```

---

### Remove Remote

```
DELETE /sync/remotes/:name
```

**Response:** 204 No Content.

**curl:**

```sh
curl -X DELETE http://localhost:9090/sync/remotes/peer-b
```

---

### Sync Status

```
GET /sync/status
```

**Response (200):**

```json
{
  "remotes": [
    {
      "name": "peer-b",
      "connected": true,
      "last_sync": "RFC3339",
      "pending_diffs": 0,
      "last_error": null,
      "lag_ms": 120
    }
  ]
}
```

**curl:**

```sh
curl http://localhost:9090/sync/status
```

---

### Push Diffs (receive from peer)

```
POST /sync/push
```

Peer sends diffs TO this engine. Body is a JSON array of Diff
objects matching Go's `sync.Diff` struct exactly:

**Request:**

```json
[
  {
    "entity_id": "abc",
    "entity_type": "notes",
    "operation": 0,
    "before": null,
    "after": "{\"title\":\"Hello\"}",
    "timestamp": {
      "physical": 1713520000000000000,
      "logical": 1,
      "node_id": "peer-b-fingerprint"
    },
    "node_id": "peer-b-fingerprint"
  }
]
```

**Operation values:**

| Value | Meaning |
|-------|---------|
| 0     | Create  |
| 1     | Update  |
| 2     | Delete  |

**Response (200):**

```json
{"accepted": 1, "rejected": 0}
```

**curl:**

```sh
curl -X POST http://localhost:9090/sync/push \
  -H 'Content-Type: application/json' \
  -d '[{"entity_id":"abc","entity_type":"notes","operation":0,
       "after":"{\"title\":\"Hello\"}",
       "timestamp":{"physical":1713520000000000000,"logical":1,
       "node_id":"peer-b"},"node_id":"peer-b"}]'
```

---

### Pull Diffs (serve to peer)

```
GET /sync/pull?since_physical=N&since_logical=N&since_node=S
```

Returns diffs since the given HLC timestamp. Peers call this
to fetch changes they haven't seen yet.

**Query params:**

| Param         | Type   | Required | Notes                  |
|---------------|--------|----------|------------------------|
| since_physical| int64  | yes      | UnixNano wall clock    |
| since_logical | uint32 | yes      | logical counter        |
| since_node    | string | yes      | originating node ID    |

**Response (200):**

```json
[
  {
    "entity_id": "xyz",
    "entity_type": "notes",
    "operation": 1,
    "before": "{\"title\":\"Old\"}",
    "after": "{\"title\":\"New\"}",
    "timestamp": {
      "physical": 1713520100000000000,
      "logical": 0,
      "node_id": "local-fingerprint"
    },
    "node_id": "local-fingerprint"
  }
]
```

**curl:**

```sh
curl "http://localhost:9090/sync/pull?since_physical=0&since_logical=0&since_node=boot"
```

---

## Identity

### Get Identity

```
GET /identity
```

**Response (200):**

```json
{
  "public_key": "-----BEGIN PUBLIC KEY-----\n...\n-----END PUBLIC KEY-----",
  "fingerprint": "a1b2c3d4e5f67890"
}
```

**curl:**

```sh
curl http://localhost:9090/identity
```

---

### Verify Payload

```
POST /identity/verify
```

Verifies a JWT signed by any known peer's public key.

**Request:**

```json
{
  "token": "eyJhbGciOiJFZERTQSIs...",
  "expected_fingerprint": "string (optional)"
}
```

**Response (200):**

```json
{
  "valid": true,
  "payload": {"sub": "peer-a", "scope": "sync"},
  "signer": "a1b2c3d4e5f67890"
}
```

**Response (200, invalid):**

```json
{
  "valid": false,
  "error": "signature mismatch"
}
```

**curl:**

```sh
curl -X POST http://localhost:9090/identity/verify \
  -H 'Content-Type: application/json' \
  -d '{"token":"eyJhbGciOiJFZERTQSIs..."}'
```

---

## Peers

### List Peers

```
GET /peers
```

Returns all discovered peers with trust status.

**Response (200):**

```json
{
  "peers": [
    {
      "id": "a1b2c3d4e5f67890",
      "name": "laptop",
      "addrs": ["192.168.1.50:8080"],
      "trust": "trusted",
      "first_seen": "RFC3339",
      "last_seen": "RFC3339"
    }
  ]
}
```

Trust values: `unknown`, `pending_tofu`, `trusted`, `blocked`.

**curl:**

```sh
curl http://localhost:9090/peers
```

---

### Trust Peer

```
POST /peers/:id/trust
```

Promotes a `pending_tofu` or `unknown` peer to `trusted`.

**Response:** 204 No Content.

**404** if peer ID not found. **409** if peer already blocked.

**curl:**

```sh
curl -X POST http://localhost:9090/peers/a1b2c3d4e5f67890/trust
```

---

### Block Peer

```
POST /peers/:id/block
```

Sets peer to `blocked`. Blocks all sync/communication.

**Response:** 204 No Content.

**curl:**

```sh
curl -X POST http://localhost:9090/peers/a1b2c3d4e5f67890/block
```

---

## Meta

### Capabilities

```
GET /capabilities
```

Self-description of engine features. Used by SDKs to negotiate
protocol version.

**Response (200):**

```json
{
  "version": "1.0.0",
  "features": ["documents", "sync", "identity", "peers", "events"],
  "max_document_size": 10485760
}
```

**curl:**

```sh
curl http://localhost:9090/capabilities
```

---

### Health

```
GET /health
```

**Response (200):**

```json
{"status": "ok", "uptime_seconds": 3600}
```

**curl:**

```sh
curl http://localhost:9090/health
```

---

### Shutdown

```
POST /shutdown
```

Graceful shutdown. Flushes pending syncs, closes connections.

**Response:** 204 No Content. Engine process exits.

**curl:**

```sh
curl -X POST http://localhost:9090/shutdown
```

---

## WebSocket: /events

Connect via WS to receive real-time bus events.

```
ws://localhost:9090/events
```

### Message Format

Each frame is a JSON object:

```json
{
  "topic": "document.created",
  "source": "engine",
  "timestamp": "RFC3339",
  "payload": {
    "type": "notes",
    "id": "abc",
    "data": {"title": "Hello"}
  }
}
```

### Event Topics

| Topic              | Fires when                    |
|--------------------|-------------------------------|
| document.created   | new document inserted         |
| document.updated   | existing document modified    |
| document.deleted   | document removed              |
| sync.push.start    | push cycle begins             |
| sync.push.complete | push cycle finishes           |
| sync.pull.start    | pull cycle begins             |
| sync.pull.complete | pull cycle finishes           |
| sync.conflict      | LWW conflict resolved         |
| peer.discovered    | new peer found via mDNS       |
| peer.connected     | peer handshake complete       |
| peer.disconnected  | peer connection lost          |

### Subscribing (filter)

Send a JSON frame after connecting to filter topics:

```json
{"subscribe": ["document.*", "sync.*"]}
```

MQTT-style wildcards: `*` matches one segment, `#` matches
all remaining segments.

### curl (wscat)

```sh
wscat -c ws://localhost:9090/events
> {"subscribe":["document.*"]}
```
