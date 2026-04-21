package store

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newVersionedStore(t *testing.T) *VersionedDocumentStore {
	t.Helper()
	return NewVersionedDocumentStore(newTestStore(t))
}

func TestVersioned_CreateAddsVersion(t *testing.T) {
	vs := newVersionedStore(t)
	ctx := context.Background()

	_, err := vs.Create(ctx, "note", json.RawMessage(`{"id":"n1","v":1}`))
	require.NoError(t, err)

	history, err := vs.History(ctx, "note", "n1")
	require.NoError(t, err)
	assert.Len(t, history, 1)
	assert.Equal(t, 1, history[0].Seq)
	assert.JSONEq(t, `{"id":"n1","v":1}`, string(history[0].Data))
}

func TestVersioned_UpdateAddsVersion(t *testing.T) {
	vs := newVersionedStore(t)
	ctx := context.Background()

	vs.Create(ctx, "note", json.RawMessage(`{"id":"n1","v":1}`))
	vs.Update(ctx, "note", "n1", json.RawMessage(`{"id":"n1","v":2}`))

	history, err := vs.History(ctx, "note", "n1")
	require.NoError(t, err)
	assert.Len(t, history, 2)
	assert.Equal(t, 2, history[1].Seq)
	assert.JSONEq(t, `{"id":"n1","v":2}`, string(history[1].Data))
}

func TestVersioned_HistoryOrdered(t *testing.T) {
	vs := newVersionedStore(t)
	ctx := context.Background()

	vs.Create(ctx, "doc", json.RawMessage(`{"id":"d1","step":1}`))
	vs.Update(ctx, "doc", "d1", json.RawMessage(`{"id":"d1","step":2}`))
	vs.Update(ctx, "doc", "d1", json.RawMessage(`{"id":"d1","step":3}`))

	history, err := vs.History(ctx, "doc", "d1")
	require.NoError(t, err)
	require.Len(t, history, 3)

	for i, v := range history {
		assert.Equal(t, i+1, v.Seq)
	}
}

func TestVersioned_RevertRestoresOldData(t *testing.T) {
	vs := newVersionedStore(t)
	ctx := context.Background()

	vs.Create(ctx, "note", json.RawMessage(`{"id":"r1","title":"original"}`))
	vs.Update(ctx, "note", "r1", json.RawMessage(`{"id":"r1","title":"changed"}`))

	doc, err := vs.Revert(ctx, "note", "r1", 1)
	require.NoError(t, err)
	assert.JSONEq(t, `{"id":"r1","title":"original"}`, string(doc.Data))

	got, _ := vs.Get(ctx, "note", "r1")
	assert.JSONEq(t, `{"id":"r1","title":"original"}`, string(got.Data))
}

func TestVersioned_RevertCreatesNewVersion(t *testing.T) {
	vs := newVersionedStore(t)
	ctx := context.Background()

	vs.Create(ctx, "note", json.RawMessage(`{"id":"r2","v":1}`))
	vs.Update(ctx, "note", "r2", json.RawMessage(`{"id":"r2","v":2}`))
	vs.Revert(ctx, "note", "r2", 1)

	history, err := vs.History(ctx, "note", "r2")
	require.NoError(t, err)
	assert.Len(t, history, 3)
	assert.Equal(t, 3, history[2].Seq)
}

func TestVersioned_RevertNotFound(t *testing.T) {
	vs := newVersionedStore(t)
	ctx := context.Background()

	vs.Create(ctx, "note", json.RawMessage(`{"id":"r3","v":1}`))

	_, err := vs.Revert(ctx, "note", "r3", 99)
	assert.ErrorContains(t, err, "not found")
}

func TestVersioned_DeleteClearsHistory(t *testing.T) {
	vs := newVersionedStore(t)
	ctx := context.Background()

	vs.Create(ctx, "note", json.RawMessage(`{"id":"del1","v":1}`))
	vs.Update(ctx, "note", "del1", json.RawMessage(`{"id":"del1","v":2}`))

	err := vs.Delete(ctx, "note", "del1")
	require.NoError(t, err)

	_, err = vs.History(ctx, "note", "del1")
	assert.ErrorContains(t, err, "no history")
}
