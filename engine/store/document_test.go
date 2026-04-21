package store

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestStore(t *testing.T) *DocumentStore {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, err := NewDocumentStore(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { s.Close() })
	return s
}

func TestCreate_AutoGenID(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	doc, err := s.Create(ctx, "note", json.RawMessage(`{"title":"hello"}`))
	require.NoError(t, err)
	assert.NotEmpty(t, doc.ID)
	assert.Equal(t, "note", doc.Type)
	assert.JSONEq(t, `{"title":"hello"}`, string(doc.Data))
}

func TestCreate_CallerProvidedID(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	doc, err := s.Create(ctx, "note", json.RawMessage(`{"id":"my-id","title":"world"}`))
	require.NoError(t, err)
	assert.Equal(t, "my-id", doc.ID)
}

func TestGet(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	created, err := s.Create(ctx, "task", json.RawMessage(`{"id":"t1","done":false}`))
	require.NoError(t, err)

	got, err := s.Get(ctx, "task", "t1")
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.JSONEq(t, `{"id":"t1","done":false}`, string(got.Data))
}

func TestGet_NotFound(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	_, err := s.Get(ctx, "task", "nope")
	assert.ErrorContains(t, err, "not found")
}

func TestUpdate(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	_, err := s.Create(ctx, "note", json.RawMessage(`{"id":"n1","v":1}`))
	require.NoError(t, err)

	updated, err := s.Update(ctx, "note", "n1", json.RawMessage(`{"id":"n1","v":2}`))
	require.NoError(t, err)
	assert.JSONEq(t, `{"id":"n1","v":2}`, string(updated.Data))
	assert.NotEqual(t, updated.CreatedAt, updated.UpdatedAt)
}

func TestUpdate_NotFound(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	_, err := s.Update(ctx, "note", "nope", json.RawMessage(`{}`))
	assert.ErrorContains(t, err, "not found")
}

func TestDelete(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	_, err := s.Create(ctx, "note", json.RawMessage(`{"id":"d1"}`))
	require.NoError(t, err)

	err = s.Delete(ctx, "note", "d1")
	require.NoError(t, err)

	_, err = s.Get(ctx, "note", "d1")
	assert.ErrorContains(t, err, "not found")
}

func TestDelete_NotFound(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	err := s.Delete(ctx, "note", "ghost")
	assert.ErrorContains(t, err, "not found")
}

func TestList_TypeIsolation(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	_, err := s.Create(ctx, "typeA", json.RawMessage(`{"id":"a1"}`))
	require.NoError(t, err)
	_, err = s.Create(ctx, "typeB", json.RawMessage(`{"id":"b1"}`))
	require.NoError(t, err)

	docs, err := s.List(ctx, "typeA", Query{})
	require.NoError(t, err)
	assert.Len(t, docs, 1)
	assert.Equal(t, "a1", docs[0].ID)
}

func TestList_LimitOffset(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	for i := range 5 {
		_, err := s.Create(ctx, "item", json.RawMessage(`{"id":"`+string(rune('a'+i))+`"}`))
		require.NoError(t, err)
	}

	docs, err := s.List(ctx, "item", Query{Limit: 2, Offset: 1})
	require.NoError(t, err)
	assert.Len(t, docs, 2)
}

func TestList_Search(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	_, err := s.Create(ctx, "note", json.RawMessage(`{"id":"s1","body":"find me here"}`))
	require.NoError(t, err)
	_, err = s.Create(ctx, "note", json.RawMessage(`{"id":"s2","body":"nothing special"}`))
	require.NoError(t, err)

	docs, err := s.List(ctx, "note", Query{Search: "find me"})
	require.NoError(t, err)
	assert.Len(t, docs, 1)
	assert.Equal(t, "s1", docs[0].ID)
}

func TestSearchEscapeWildcards(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	// Create docs: one with literal %, one with literal _, one normal.
	_, err := s.Create(ctx, "note", json.RawMessage(`{"id":"w1","body":"100% done"}`))
	require.NoError(t, err)
	_, err = s.Create(ctx, "note", json.RawMessage(`{"id":"w2","body":"under_score"}`))
	require.NoError(t, err)
	_, err = s.Create(ctx, "note", json.RawMessage(`{"id":"w3","body":"nothing here"}`))
	require.NoError(t, err)

	// Search with "%" should only match the doc containing literal "%"
	docs, err := s.List(ctx, "note", Query{Search: "%"})
	require.NoError(t, err)
	assert.Len(t, docs, 1)
	assert.Equal(t, "w1", docs[0].ID)

	// Search with "_" should only match the doc containing literal "_"
	docs, err = s.List(ctx, "note", Query{Search: "_"})
	require.NoError(t, err)
	assert.Len(t, docs, 1)
	assert.Equal(t, "w2", docs[0].ID)
}

func TestTimestampsSet(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	doc, err := s.Create(ctx, "ts", json.RawMessage(`{"id":"t1"}`))
	require.NoError(t, err)
	assert.NotEmpty(t, doc.CreatedAt)
	assert.NotEmpty(t, doc.UpdatedAt)
	assert.Equal(t, doc.CreatedAt, doc.UpdatedAt)
}
