package aim

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockSource implements Source for testing.
type mockSource struct {
	providers map[string]Provider
	err       error
	calls     int
}

func (m *mockSource) Fetch(_ context.Context) (map[string]Provider, error) {
	m.calls++
	return m.providers, m.err
}

func testProviders() map[string]Provider {
	return map[string]Provider{
		"test": {ID: "test", Name: "Test", Models: map[string]Model{
			"t-1": {ID: "t-1", Name: "Test One"},
		}},
	}
}

func seedCache(t *testing.T, dir string, fetchedAt time.Time, providers map[string]Provider) {
	t.Helper()
	require.NoError(t, os.MkdirAll(dir, 0o755))
	env := cacheEnvelope{FetchedAt: fetchedAt, Providers: providers}
	data, err := json.Marshal(env)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, dataFile), data, 0o644))
}

func newTestCache(t *testing.T, src Source, ttl time.Duration) *Cache {
	t.Helper()
	dir := filepath.Join(t.TempDir(), cacheSubdir)
	c, err := NewCache(src, WithCacheDir(dir), WithTTL(ttl))
	require.NoError(t, err)
	return c
}

func TestCache_FreshFromDisk(t *testing.T) {
	src := &mockSource{providers: testProviders()}
	c := newTestCache(t, src, time.Hour)
	seedCache(t, c.dir, time.Now(), testProviders())

	got, err := c.Fetch(context.Background())
	require.NoError(t, err)
	assert.Contains(t, got, "test")
	assert.Equal(t, 0, src.calls)
}

func TestCache_ExpiredTTL(t *testing.T) {
	src := &mockSource{providers: testProviders()}
	c := newTestCache(t, src, time.Millisecond)
	seedCache(t, c.dir, time.Now().Add(-time.Hour), testProviders())

	got, err := c.Fetch(context.Background())
	require.NoError(t, err)
	assert.Contains(t, got, "test")
	assert.Equal(t, 1, src.calls)
}

func TestCache_ForceRefresh(t *testing.T) {
	src := &mockSource{providers: testProviders()}
	c := newTestCache(t, src, time.Hour)
	seedCache(t, c.dir, time.Now(), testProviders())

	got, err := c.Refresh(context.Background(), true)
	require.NoError(t, err)
	assert.Contains(t, got, "test")
	assert.Equal(t, 1, src.calls)
}

func TestCache_StaleOnError(t *testing.T) {
	src := &mockSource{err: errors.New("network down")}
	c := newTestCache(t, src, time.Millisecond)
	seedCache(t, c.dir, time.Now().Add(-time.Hour), testProviders())

	got, err := c.Fetch(context.Background())
	require.NoError(t, err)
	assert.Contains(t, got, "test")
}

func TestCache_CorruptRecovery(t *testing.T) {
	src := &mockSource{providers: testProviders()}
	c := newTestCache(t, src, time.Hour)
	require.NoError(t, os.MkdirAll(c.dir, 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(c.dir, dataFile), []byte("{bad json"), 0o644,
	))

	got, err := c.Fetch(context.Background())
	require.NoError(t, err)
	assert.Contains(t, got, "test")
	assert.Equal(t, 1, src.calls)
}

func TestCache_AtomicWrite(t *testing.T) {
	src := &mockSource{providers: testProviders()}
	c := newTestCache(t, src, time.Millisecond)

	_, err := c.Fetch(context.Background())
	require.NoError(t, err)

	// no leftover .tmp file
	_, statErr := os.Stat(filepath.Join(c.dir, dataFile+".tmp"))
	assert.True(t, os.IsNotExist(statErr))

	// final file is valid JSON
	data, readErr := os.ReadFile(filepath.Join(c.dir, dataFile))
	require.NoError(t, readErr)
	var env cacheEnvelope
	require.NoError(t, json.Unmarshal(data, &env))
	assert.Contains(t, env.Providers, "test")
}

func TestCache_RoundTripNormalize(t *testing.T) {
	providers := map[string]Provider{
		"acme": {ID: "acme", Name: "Acme", Models: map[string]Model{
			"m-1": {
				ID: "m-1", Name: "Model One",
				Modalities: &Modalities{
					Input:  []string{"text", "image"},
					Output: []string{"text"},
				},
				Limit: &Limit{Context: 128000, MaxOutput: 4096},
				Cost:  &Cost{Input: 3.0, Output: 15.0},
			},
		}},
	}

	src := &mockSource{err: errors.New("should not fetch")}
	c := newTestCache(t, src, time.Hour)
	seedCache(t, c.dir, time.Now(), providers)

	got, err := c.Fetch(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 0, src.calls)

	m := got["acme"].Models["m-1"]
	assert.Equal(t, []string{"text", "image"}, m.Input)
	assert.Equal(t, []string{"text"}, m.Output)
	assert.Equal(t, 128000, m.Context)
	assert.Equal(t, 4096, m.MaxOutput)
	assert.Equal(t, 3.0, m.CostInput)
	assert.Equal(t, 15.0, m.CostOutput)
	assert.Equal(t, "acme", m.Provider)
}

func TestCache_LoadNormalizesWithoutFetch(t *testing.T) {
	providers := map[string]Provider{
		"prov": {ID: "prov", Name: "Prov", Models: map[string]Model{
			"x-1": {
				ID: "x-1", Name: "X One",
				Modalities: &Modalities{
					Input:  []string{"audio"},
					Output: []string{"text"},
				},
				Limit: &Limit{Context: 64000, MaxOutput: 2048},
				Cost:  &Cost{Input: 1.5, Output: 7.5},
			},
		}},
	}

	src := &mockSource{err: errors.New("should not be called")}
	c := newTestCache(t, src, time.Hour)
	seedCache(t, c.dir, time.Now(), providers)

	got, err := c.Fetch(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 0, src.calls)

	m := got["prov"].Models["x-1"]
	assert.Equal(t, []string{"audio"}, m.Input)
	assert.Equal(t, []string{"text"}, m.Output)
	assert.Equal(t, 64000, m.Context)
	assert.Equal(t, 2048, m.MaxOutput)
	assert.Equal(t, 1.5, m.CostInput)
	assert.Equal(t, 7.5, m.CostOutput)
	assert.Equal(t, "prov", m.Provider)
}

func TestCache_ETagRoundTrip(t *testing.T) {
	c := newTestCache(t, &mockSource{}, time.Hour)
	require.NoError(t, os.MkdirAll(c.dir, 0o755))

	assert.Empty(t, c.loadETag())

	c.storeETag("\"v1\"")
	assert.Equal(t, "\"v1\"", c.loadETag())

	// no leftover tmp
	_, err := os.Stat(filepath.Join(c.dir, etagFile+".tmp"))
	assert.True(t, os.IsNotExist(err))
}
