package aim

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fixtureServer(t *testing.T) *httptest.Server {
	t.Helper()
	data, err := os.ReadFile("testdata/api-fixture.json")
	require.NoError(t, err)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func fixtureRegistry(t *testing.T, srv *httptest.Server) *Registry {
	t.Helper()
	cacheDir := t.TempDir()
	r, err := NewRegistry(
		WithSource(NewModelsDevSource(WithURL(srv.URL))),
		WithRegistryCacheDir(cacheDir),
	)
	require.NoError(t, err)
	return r
}

func TestE2E_ImageInput(t *testing.T) {
	r := fixtureRegistry(t, fixtureServer(t))
	models := r.Models(Filter{Input: "image"})
	require.NotEmpty(t, models)
	for _, m := range models {
		assert.Contains(t, m.Input, "image")
	}
	// sonnet, haiku, claude-3-7-sonnet, gpt-4o all accept image
	assert.Len(t, models, 4)
}

func TestE2E_ImageOutput(t *testing.T) {
	r := fixtureRegistry(t, fixtureServer(t))
	models := r.Models(Filter{Output: "image"})
	require.Len(t, models, 1)
	assert.Equal(t, "dall-e-3", models[0].ID)
	assert.Equal(t, "openai", models[0].Provider)
}

func TestE2E_ReasoningAcrossProviders(t *testing.T) {
	r := fixtureRegistry(t, fixtureServer(t))
	tr := true
	models := r.Models(Filter{Reasoning: &tr})
	require.Len(t, models, 2)
	providers := map[string]bool{}
	for _, m := range models {
		providers[m.Provider] = true
	}
	assert.True(t, providers["anthropic"], "anthropic reasoning model")
	assert.True(t, providers["openai"], "openai reasoning model")
}

func TestE2E_GetModel(t *testing.T) {
	r := fixtureRegistry(t, fixtureServer(t))
	m, ok := r.Get("anthropic", "claude-3-5-sonnet")
	require.True(t, ok)
	assert.Equal(t, "Claude 3.5 Sonnet", m.Name)
	assert.Equal(t, "anthropic", m.Provider)
	assert.Equal(t, "claude", m.Family)
	assert.Equal(t, 200000, m.Context)
	assert.Equal(t, 8192, m.MaxOutput)
	assert.Equal(t, 3.0, m.CostInput)
	assert.Equal(t, 15.0, m.CostOutput)
	assert.True(t, m.ToolCall)
	assert.False(t, m.Reasoning)
	assert.Contains(t, m.Input, "image")
}

func TestE2E_ProviderFilter(t *testing.T) {
	r := fixtureRegistry(t, fixtureServer(t))
	models := r.Models(Filter{Provider: "openai"})
	require.Len(t, models, 3)
	for _, m := range models {
		assert.Equal(t, "openai", m.Provider)
	}
}

func TestE2E_QueryMatchesFilter(t *testing.T) {
	r := fixtureRegistry(t, fixtureServer(t))
	fromFilter := r.Models(Filter{Provider: "openai"})
	fromQuery, err := r.Query("provider:openai")
	require.NoError(t, err)
	assert.Equal(t, fromFilter, fromQuery)
}

func TestE2E_CustomSource(t *testing.T) {
	srv := fixtureServer(t)
	custom := &staticSource{providers: map[string]Provider{
		"internal": {
			ID: "internal", Name: "Internal",
			Models: map[string]Model{
				"custom-1": {
					ID: "custom-1", Name: "Custom One", Provider: "internal",
					Modalities: &Modalities{Input: []string{"text"}, Output: []string{"text"}},
				},
			},
		},
	}}
	cacheDir := t.TempDir()
	r, err := NewRegistry(
		WithSource(NewModelsDevSource(WithURL(srv.URL))),
		WithSource(custom),
		WithRegistryCacheDir(cacheDir),
	)
	require.NoError(t, err)
	m, ok := r.Get("internal", "custom-1")
	require.True(t, ok)
	assert.Equal(t, "Custom One", m.Name)
	// fixture providers still present
	_, ok = r.Get("anthropic", "claude-3-5-sonnet")
	assert.True(t, ok)
}

func TestE2E_UnknownFieldsIgnored(t *testing.T) {
	// llama-3-8b has "quantization" and "future_field" — extra fields
	r := fixtureRegistry(t, fixtureServer(t))
	m, ok := r.Get("meta", "llama-3-8b")
	require.True(t, ok)
	assert.Equal(t, "Llama 3 8B", m.Name)
	assert.True(t, m.OpenWeights)
}

func TestE2E_EmptyResult(t *testing.T) {
	r := fixtureRegistry(t, fixtureServer(t))
	models := r.Models(Filter{Provider: "nonexistent"})
	assert.Empty(t, models)
	assert.Len(t, models, 0)
}

func TestE2E_ForceRefresh(t *testing.T) {
	var calls atomic.Int32
	data, err := os.ReadFile("testdata/api-fixture.json")
	require.NoError(t, err)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	}))
	t.Cleanup(srv.Close)

	r := fixtureRegistry(t, srv)
	_ = r.Providers() // initial fetch
	before := calls.Load()

	err = r.Refresh(context.Background())
	require.NoError(t, err)
	assert.Greater(t, calls.Load(), before, "refresh should re-fetch")
}

func TestE2E_ProviderFieldFromParent(t *testing.T) {
	r := fixtureRegistry(t, fixtureServer(t))
	for _, p := range r.Providers() {
		for _, m := range p.Models {
			assert.Equal(t, p.ID, m.Provider,
				"Model.Provider should match parent key for %s/%s", p.ID, m.ID)
		}
	}
}

func TestE2E_FreeTextQuery(t *testing.T) {
	r := fixtureRegistry(t, fixtureServer(t))
	models, err := r.Query("claude")
	require.NoError(t, err)
	require.NotEmpty(t, models)
	for _, m := range models {
		assert.Contains(t, m.ID, "claude", "free-text should match model ID")
	}
	assert.Len(t, models, 3) // sonnet, haiku, 3-7-sonnet
}

// staticSource is a test-only Source returning fixed data.
type staticSource struct {
	providers map[string]Provider
}

func (s *staticSource) Fetch(_ context.Context) (map[string]Provider, error) {
	// normalize before returning
	for k, p := range s.providers {
		for id, m := range p.Models {
			m.Normalize()
			p.Models[id] = m
		}
		s.providers[k] = p
	}
	return s.providers, nil
}
