package store

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"hop.top/kit/go/core/util"
	"hop.top/kit/go/runtime/domain/version"
)

// Version records a point-in-time snapshot of a document's data.
type Version struct {
	Type      string          `json:"type"`
	ID        string          `json:"id"`
	VersionID string          `json:"version_id"`
	Seq       int             `json:"seq"`
	Data      json.RawMessage `json:"data"`
	CreatedAt string          `json:"created_at"`
}

// VersionedDocumentStore wraps DocumentStore with in-memory version
// tracking via a DAG on every mutation.
// WARNING: version history is ephemeral — lost on restart. For
// persistent versioning, use domain/version with a SQLite-backed
// DAG (planned for future release).
type VersionedDocumentStore struct {
	store     *DocumentStore
	mu        sync.RWMutex
	dags      map[string]*version.DAG    // key: "type:id"
	versions  map[string][]Version       // key: "type:id" -> ordered versions
	snapshots map[string]json.RawMessage // versionID -> data snapshot
}

// NewVersionedDocumentStore wraps an existing DocumentStore.
func NewVersionedDocumentStore(s *DocumentStore) *VersionedDocumentStore {
	return &VersionedDocumentStore{
		store:     s,
		dags:      make(map[string]*version.DAG),
		versions:  make(map[string][]Version),
		snapshots: make(map[string]json.RawMessage),
	}
}

// Create inserts a document and records version 1.
func (vs *VersionedDocumentStore) Create(ctx context.Context, docType string, data json.RawMessage) (Document, error) {
	doc, err := vs.store.Create(ctx, docType, data)
	if err != nil {
		return Document{}, err
	}
	vs.appendVersion(doc.Type, doc.ID, data)
	return doc, nil
}

// Update replaces document data and appends a new version.
func (vs *VersionedDocumentStore) Update(ctx context.Context, docType, id string, data json.RawMessage) (Document, error) {
	doc, err := vs.store.Update(ctx, docType, id, data)
	if err != nil {
		return Document{}, err
	}
	vs.appendVersion(docType, id, data)
	return doc, nil
}

// Get delegates to the inner store.
func (vs *VersionedDocumentStore) Get(ctx context.Context, docType, id string) (Document, error) {
	return vs.store.Get(ctx, docType, id)
}

// List delegates to the inner store.
func (vs *VersionedDocumentStore) List(ctx context.Context, docType string, q Query) ([]Document, error) {
	return vs.store.List(ctx, docType, q)
}

// Delete removes the document and its version history.
func (vs *VersionedDocumentStore) Delete(ctx context.Context, docType, id string) error {
	if err := vs.store.Delete(ctx, docType, id); err != nil {
		return err
	}
	key := docKey(docType, id)
	vs.mu.Lock()
	delete(vs.dags, key)
	delete(vs.versions, key)
	vs.mu.Unlock()
	return nil
}

// History returns all versions for a document ordered by seq.
func (vs *VersionedDocumentStore) History(ctx context.Context, docType, id string) ([]Version, error) {
	key := docKey(docType, id)
	vs.mu.RLock()
	defer vs.mu.RUnlock()

	result := vs.versions[key]
	if len(result) == 0 {
		return nil, fmt.Errorf("store: no history for %s/%s", docType, id)
	}
	return result, nil
}

// Revert restores a document to the given version sequence number.
// The revert itself creates a new version entry.
func (vs *VersionedDocumentStore) Revert(ctx context.Context, docType, id string, seq int) (Document, error) {
	key := docKey(docType, id)
	vs.mu.RLock()
	var target *Version
	for i := range vs.versions[key] {
		if vs.versions[key][i].Seq == seq {
			target = &vs.versions[key][i]
			break
		}
	}
	vs.mu.RUnlock()

	if target == nil {
		return Document{}, fmt.Errorf("store: version %d not found for %s/%s", seq, docType, id)
	}

	doc, err := vs.store.Update(ctx, docType, id, target.Data)
	if err != nil {
		return Document{}, err
	}
	vs.appendVersion(docType, id, target.Data)
	return doc, nil
}

func (vs *VersionedDocumentStore) appendVersion(docType, id string, data json.RawMessage) {
	key := docKey(docType, id)
	vs.mu.Lock()
	defer vs.mu.Unlock()

	dag, ok := vs.dags[key]
	if !ok {
		dag = version.NewDAG()
		vs.dags[key] = dag
	}

	existing := vs.versions[key]
	seq := len(existing) + 1
	vid := util.Short([]byte(fmt.Sprintf("%s-%d-%s", key, seq, data)), 16)

	var parents []string
	if len(existing) > 0 {
		parents = []string{existing[len(existing)-1].VersionID}
	}

	_ = dag.Append(version.Version{
		ID:        vid,
		ParentIDs: parents,
		Timestamp: time.Now().UnixNano(),
		Hash:      util.Short(data, 16),
	})

	v := Version{
		Type:      docType,
		ID:        id,
		VersionID: vid,
		Seq:       seq,
		Data:      append(json.RawMessage(nil), data...),
		CreatedAt: time.Now().UTC().Format(time.RFC3339Nano),
	}
	vs.versions[key] = append(existing, v)
	vs.snapshots[vid] = append(json.RawMessage(nil), data...)
}

func docKey(docType, id string) string {
	return docType + ":" + id
}
