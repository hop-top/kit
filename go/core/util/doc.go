// Package util provides common utility helpers for kit projects.
//
// Helpers:
//   - env: typed environment variable reader with defaults
//   - fingerprint: consistent short SHA-256 hashes for IDs
//   - humanize: human-friendly durations and byte sizes
//   - jsonl: newline-delimited JSON read/write/stream
//   - must: panic-on-error wrappers for init-time setup
//   - ptr: generic pointer helpers
//   - retry: configurable retry with backoff
//   - since: duration parsing from human strings
//   - slug: URL-safe slug generation
package util
