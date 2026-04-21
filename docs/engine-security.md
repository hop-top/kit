# Engine Security Guide

Each kit serve instance has its OWN Ed25519 keypair. A Go app's
identity and a TS engine's identity are equivalent — same key
format, same JWT signing, same trust model.

---

## 1. Identity Lifecycle

On first run, the engine auto-generates an Ed25519 keypair and
persists it in the data directory:

```
<data-dir>/
  identity/
    public.pem      # PKIX-encoded Ed25519 public key
    private.pem     # PKCS8-encoded Ed25519 private key
```

If keys already exist, the engine loads them. No manual key
generation required.

Fingerprint = first 8 bytes of SHA-256(public key), hex-encoded
(16 chars). Used as the engine's peer ID.

---

## 2. GET /identity

Retrieve the engine's public identity:

```sh
curl http://localhost:9090/identity
```

**Response:**

```json
{
  "public_key": "-----BEGIN PUBLIC KEY-----\nMCow...\n-----END PUBLIC KEY-----",
  "fingerprint": "a1b2c3d4e5f67890"
}
```

The fingerprint uniquely identifies this engine in the peer mesh.
Other peers store this fingerprint + public key in their registry.

---

## 3. JWT Verification

### Verify

Validate a JWT against known peer public keys:

```sh
curl -X POST http://localhost:9090/identity/verify \
  -H 'Content-Type: application/json' \
  -d '{"token":"eyJhbGciOiJFZERTQSIs..."}'
```

**Response (valid):**

```json
{
  "valid": true,
  "payload": {"sub": "peer-a", "scope": "sync"},
  "signer": "a1b2c3d4e5f67890"
}
```

**Response (invalid):**

```json
{"valid": false, "error": "unknown signer"}
```

Verification checks:
1. Signature valid against any known peer's public key
2. Token not expired
3. If `expected_fingerprint` provided: signer must match

---

## 4. Peer Authentication

Peers authenticate via JWTs signed with their Ed25519 keypairs.
Sync endpoints validate incoming requests:

1. Peer presents JWT in `Authorization: Bearer <token>` header
2. Engine verifies signature against peer registry
3. Only `trusted` peers accepted; `blocked` peers rejected (401)

No shared secrets. Each peer signs with their own private key;
receivers verify with the sender's stored public key.

---

## 5. Trust Flow

Complete trust establishment from discovery to sync:

```
  ┌──────────┐                         ┌──────────┐
  │ Engine A │                         │ Engine B │
  └────┬─────┘                         └────┬─────┘
       │                                     │
       │  1. mDNS announce + browse          │
       │<────────── discover ────────────────>│
       │                                     │
       │  2. GET /identity                   │
       │────────────────────────────────────>│
       │<──── {public_key, fingerprint} ─────│
       │                                     │
       │  3. AcceptTOFU → PendingTOFU        │
       │  (stored locally, not yet trusted)  │
       │                                     │
       │  4. User approves:                  │
       │     POST /peers/:id/trust           │
       │  → TrustLevel = Trusted             │
       │                                     │
       │  5. Sync begins (both directions)   │
       │<═══════════ sync diffs ════════════>│
       │                                     │
```

### Trust Levels

| Level        | Meaning                           |
|--------------|-----------------------------------|
| unknown      | never seen                        |
| pending_tofu | discovered, awaiting user approval|
| trusted      | explicitly approved, sync allowed |
| blocked      | rejected, all communication denied|

### Promoting Trust

```sh
# List discovered peers
curl http://localhost:9090/peers

# Trust a pending peer
curl -X POST http://localhost:9090/peers/a1b2c3d4e5f67890/trust

# Block a peer
curl -X POST http://localhost:9090/peers/a1b2c3d4e5f67890/block
```

---

## 6. Encryption at Rest

Enable with `--encrypt` flag:

```sh
kit serve --port 9090 --data ./mydata --encrypt
```

### How It Works

1. Derive 32-byte symmetric key from Ed25519 private key
   via HKDF-SHA256 (domain: `kit-identity-encryption-v1`)
2. Encrypt each document with NaCl secretbox (XSalsa20-Poly1305)
3. Storage format: `nonce (24 bytes) || ciphertext`
4. Decrypt on read using same derived key

Key derivation is deterministic — same keypair always produces
same encryption key. No separate key management needed.

### Properties

- Authenticated encryption (Poly1305 MAC)
- Random nonce per write (no nonce reuse)
- Tied to identity — moving data without the private key = useless
- Transparent to API consumers (encrypt/decrypt in engine layer)

---

## 7. Cross-Language Trust

A Go app and a TS engine trust each other identically:

### Setup

```sh
# Go app at :8080, TS engine at :9090

# From TS: fetch Go app's identity
curl http://localhost:8080/identity
# Returns: {"public_key":"...", "fingerprint":"go-fp-123"}

# After mDNS discovery or manual peer add, trust it:
curl -X POST http://localhost:9090/peers/go-fp-123/trust

# From Go: trust TS engine
curl -X POST http://localhost:8080/peers/ts-fp-456/trust
```

Both peers now sync freely. The wire protocol is identical
regardless of implementation language.

### Key Format Compatibility

| Aspect     | Go (kit/identity)  | Engine (kit serve)     |
|------------|--------------------|-----------------------|
| Algorithm  | Ed25519            | Ed25519               |
| Public PEM | PKIX               | PKIX                  |
| Private PEM| PKCS8              | PKCS8                 |
| JWT algo   | EdDSA              | EdDSA                 |
| Fingerprint| SHA-256 first 8B   | SHA-256 first 8B      |

No conversion needed. Keys are byte-for-byte compatible.

---

## 8. Threat Model

### Protected

| Threat                | Mitigation                        |
|-----------------------|-----------------------------------|
| Data at rest exposure | NaCl secretbox encryption         |
| Peer impersonation    | Ed25519 signature verification    |
| Replay attacks        | JWT expiry + HLC timestamps       |
| Unauthorized sync     | trust registry (must be Trusted)  |
| Key compromise detect | pubkey mismatch = hard error      |

### NOT Protected (by design)

| Scenario             | Assumption                         |
|----------------------|------------------------------------|
| Localhost transport  | trusted host; no TLS on loopback   |
| LAN eavesdropping    | mDNS is plaintext; use VPN if needed |
| Physical access      | if attacker has disk + private key, game over |
| DoS on engine port   | host-level firewall responsibility |

### Recommendations

- Run engine on loopback only (default `127.0.0.1`)
- For remote peers over WAN: terminate TLS at reverse proxy
- Rotate keys by deleting `identity/` dir and restarting
  (requires re-establishing trust with all peers)
- Monitor `peer.discovered` events for unexpected peers
- Block unknown peers promptly in production environments
