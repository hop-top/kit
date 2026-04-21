package aim

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const validJSON = `{
	"acme": {
		"id": "acme",
		"name": "Acme AI",
		"models": {
			"acme-1": {
				"id": "acme-1",
				"name": "Acme One",
				"modalities": {"input": ["text"], "output": ["text"]},
				"limit": {"context": 128000, "output": 4096}
			}
		}
	}
}`

func testServer(t *testing.T, status int, body string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestFetch_Success(t *testing.T) {
	srv := testServer(t, http.StatusOK, validJSON)
	src := NewModelsDevSource(WithURL(srv.URL))

	providers, err := src.Fetch(context.Background())
	require.NoError(t, err)
	require.Contains(t, providers, "acme")

	m := providers["acme"].Models["acme-1"]
	assert.Equal(t, []string{"text"}, m.Input)
	assert.Equal(t, 128000, m.Context)
	assert.Equal(t, 4096, m.MaxOutput)
}

func TestFetch_KeyMismatch(t *testing.T) {
	body := `{"wrong-key": {"id": "acme", "name": "Acme", "models": {}}}`
	srv := testServer(t, http.StatusOK, body)
	src := NewModelsDevSource(WithURL(srv.URL))

	_, err := src.Fetch(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), `map key "wrong-key" != provider.id "acme"`)
}

func TestFetch_UnknownFields(t *testing.T) {
	body := `{"acme": {"id": "acme", "name": "Acme", "models": {}, "future_field": true}}`
	srv := testServer(t, http.StatusOK, body)
	src := NewModelsDevSource(WithURL(srv.URL))

	_, err := src.Fetch(context.Background())
	assert.NoError(t, err)
}

func TestFetch_MaxSizeExceeded(t *testing.T) {
	srv := testServer(t, http.StatusOK, validJSON)
	src := NewModelsDevSource(WithURL(srv.URL), WithMaxResponseSize(10))

	_, err := src.Fetch(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds max size")
}

func TestFetch_HTTPError(t *testing.T) {
	srv := testServer(t, http.StatusInternalServerError, "")
	src := NewModelsDevSource(WithURL(srv.URL))

	_, err := src.Fetch(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected status 500")
}

func TestFetchWithETag_NotModified(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("If-None-Match") == `"abc123"` {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(validJSON))
	}))
	t.Cleanup(srv.Close)

	src := NewModelsDevSource(WithURL(srv.URL))
	providers, etag, notModified, err := src.FetchWithETag(context.Background(), `"abc123"`)
	require.NoError(t, err)
	assert.True(t, notModified)
	assert.Nil(t, providers)
	assert.Equal(t, `"abc123"`, etag)
}

func TestFetchWithETag_NewETag(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("ETag", `"new-etag"`)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(validJSON))
	}))
	t.Cleanup(srv.Close)

	src := NewModelsDevSource(WithURL(srv.URL))
	providers, etag, notModified, err := src.FetchWithETag(context.Background(), "")
	require.NoError(t, err)
	assert.False(t, notModified)
	assert.NotNil(t, providers)
	assert.Equal(t, `"new-etag"`, etag)
}
