package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"hop.top/kit/go/core/util"
	"hop.top/kit/go/storage/sqldb"
)

// Document is a type-tagged JSON blob stored in SQLite.
type Document struct {
	Type      string          `json:"type"`
	ID        string          `json:"id"`
	Data      json.RawMessage `json:"data"`
	CreatedAt string          `json:"created_at"`
	UpdatedAt string          `json:"updated_at"`
}

// GetID satisfies domain.Entity.
func (d Document) GetID() string { return d.ID }

// Query holds list/search parameters.
type Query struct {
	Limit  int
	Offset int
	Sort   string
	Search string
}

// DocumentStore manages typed JSON documents in a single SQLite table.
type DocumentStore struct {
	db *sql.DB
}

const createTableSQL = `CREATE TABLE IF NOT EXISTS documents (
	type       TEXT NOT NULL,
	id         TEXT NOT NULL,
	data       TEXT NOT NULL,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL,
	PRIMARY KEY (type, id)
)`

// NewDocumentStore opens (or creates) an SQLite DB at dbPath and
// ensures the documents table exists.
func NewDocumentStore(dbPath string) (*DocumentStore, error) {
	db, err := sqldb.Open(sqldb.Options{Path: dbPath})
	if err != nil {
		return nil, fmt.Errorf("store: open db: %w", err)
	}
	if _, err := db.Exec(createTableSQL); err != nil {
		db.Close()
		return nil, fmt.Errorf("store: create table: %w", err)
	}
	return &DocumentStore{db: db}, nil
}

// Create inserts a new document. If the JSON data contains an "id"
// field it is used; otherwise one is generated via util.Short.
func (s *DocumentStore) Create(ctx context.Context, docType string, data json.RawMessage) (Document, error) {
	id := extractID(data)
	if id == "" {
		id = util.Short([]byte(fmt.Sprintf("%s-%d", docType, time.Now().UnixNano())), 12)
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)
	doc := Document{
		Type:      docType,
		ID:        id,
		Data:      data,
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO documents (type, id, data, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		doc.Type, doc.ID, string(doc.Data), doc.CreatedAt, doc.UpdatedAt,
	)
	if err != nil {
		return Document{}, fmt.Errorf("store: create: %w", err)
	}
	return doc, nil
}

// Get retrieves a single document by type and ID.
func (s *DocumentStore) Get(ctx context.Context, docType, id string) (Document, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT type, id, data, created_at, updated_at FROM documents WHERE type = ? AND id = ?`,
		docType, id,
	)
	return scanDocument(row)
}

// List returns documents of the given type matching the query.
func (s *DocumentStore) List(ctx context.Context, docType string, q Query) ([]Document, error) {
	if q.Limit == 0 {
		q.Limit = 100
	}

	query := `SELECT type, id, data, created_at, updated_at FROM documents WHERE type = ?`
	args := []any{docType}

	if q.Search != "" {
		query += ` AND data LIKE ? ESCAPE '\'`
		args = append(args, "%"+escapeLIKE(q.Search)+"%")
	}

	switch q.Sort {
	case "created_at", "updated_at", "id":
		query += ` ORDER BY ` + q.Sort
	default:
		query += ` ORDER BY created_at`
	}

	query += ` LIMIT ?`
	args = append(args, q.Limit)
	if q.Offset > 0 {
		query += ` OFFSET ?`
		args = append(args, q.Offset)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("store: list: %w", err)
	}
	defer rows.Close()

	var docs []Document
	for rows.Next() {
		var d Document
		var data string
		if err := rows.Scan(&d.Type, &d.ID, &data, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, fmt.Errorf("store: scan: %w", err)
		}
		d.Data = json.RawMessage(data)
		docs = append(docs, d)
	}
	return docs, rows.Err()
}

// Update replaces the data for an existing document.
func (s *DocumentStore) Update(ctx context.Context, docType, id string, data json.RawMessage) (Document, error) {
	now := time.Now().UTC().Format(time.RFC3339Nano)

	res, err := s.db.ExecContext(ctx,
		`UPDATE documents SET data = ?, updated_at = ? WHERE type = ? AND id = ?`,
		string(data), now, docType, id,
	)
	if err != nil {
		return Document{}, fmt.Errorf("store: update: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return Document{}, fmt.Errorf("store: update: not found")
	}
	return s.Get(ctx, docType, id)
}

// Delete removes a document by type and ID.
func (s *DocumentStore) Delete(ctx context.Context, docType, id string) error {
	res, err := s.db.ExecContext(ctx,
		`DELETE FROM documents WHERE type = ? AND id = ?`,
		docType, id,
	)
	if err != nil {
		return fmt.Errorf("store: delete: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("store: delete: not found")
	}
	return nil
}

// Close closes the underlying database connection.
func (s *DocumentStore) Close() error {
	return s.db.Close()
}

func escapeLIKE(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, "%", `\%`)
	s = strings.ReplaceAll(s, "_", `\_`)
	return s
}

func extractID(data json.RawMessage) string {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return ""
	}
	raw, ok := m["id"]
	if !ok {
		return ""
	}
	var id string
	if err := json.Unmarshal(raw, &id); err != nil {
		return ""
	}
	return id
}

func scanDocument(row *sql.Row) (Document, error) {
	var d Document
	var data string
	if err := row.Scan(&d.Type, &d.ID, &data, &d.CreatedAt, &d.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return Document{}, fmt.Errorf("store: not found")
		}
		return Document{}, fmt.Errorf("store: scan: %w", err)
	}
	d.Data = json.RawMessage(data)
	return d, nil
}
