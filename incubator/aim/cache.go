package aim

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	defaultTTL   = 24 * time.Hour
	cacheSubdir  = "hop/aim"
	dataFile     = "providers.json"
	etagFile     = "etag"
	lockFile     = ".lock"
	lockTimeout  = 30 * time.Second
	lockPollFreq = 50 * time.Millisecond
)

// cacheEnvelope is the on-disk JSON structure.
type cacheEnvelope struct {
	FetchedAt time.Time           `json:"fetched_at"`
	Providers map[string]Provider `json:"providers"`
}

// CacheOption configures a Cache.
type CacheOption func(*Cache)

// WithTTL sets the cache TTL.
func WithTTL(d time.Duration) CacheOption {
	return func(c *Cache) { c.ttl = d }
}

// WithCacheDir overrides the cache directory.
func WithCacheDir(dir string) CacheOption {
	return func(c *Cache) { c.dir = dir }
}

// Cache wraps a Source with an XDG-compliant file cache.
type Cache struct {
	src Source
	ttl time.Duration
	dir string // resolved cache dir
}

// NewCache returns a caching wrapper around src.
func NewCache(src Source, opts ...CacheOption) (*Cache, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return nil, fmt.Errorf("aim: resolve cache dir: %w", err)
	}
	c := &Cache{
		src: src,
		ttl: defaultTTL,
		dir: filepath.Join(base, cacheSubdir),
	}
	for _, o := range opts {
		o(c)
	}
	return c, nil
}

// Fetch returns cached data if fresh, otherwise fetches from source.
func (c *Cache) Fetch(ctx context.Context) (map[string]Provider, error) {
	if env, err := c.load(); err == nil && time.Since(env.FetchedAt) < c.ttl {
		return env.Providers, nil
	}

	providers, err := c.refresh(ctx)
	if err != nil {
		// stale-on-error: serve stale cache on network failure
		if env, loadErr := c.load(); loadErr == nil {
			return env.Providers, nil
		}
		return nil, err
	}
	return providers, nil
}

// Refresh forces a cache update. force=true ignores TTL.
func (c *Cache) Refresh(ctx context.Context, force bool) (map[string]Provider, error) {
	if !force {
		if env, err := c.load(); err == nil && time.Since(env.FetchedAt) < c.ttl {
			return env.Providers, nil
		}
	}
	return c.refresh(ctx)
}

func (c *Cache) refresh(ctx context.Context) (map[string]Provider, error) {
	if err := os.MkdirAll(c.dir, 0o755); err != nil {
		return nil, fmt.Errorf("aim: create cache dir: %w", err)
	}

	unlock, err := c.lock(ctx)
	if err != nil {
		return nil, err
	}
	defer unlock()

	// conditional fetch with ETag if source supports it
	var providers map[string]Provider
	if es, ok := c.src.(ETagSource); ok {
		etag := c.loadETag()
		p, newETag, notModified, fetchErr := es.FetchWithETag(ctx, etag)
		if fetchErr != nil {
			return nil, fetchErr
		}
		if notModified {
			// touch timestamp so TTL resets
			if env, loadErr := c.load(); loadErr == nil {
				env.FetchedAt = time.Now()
				_ = c.store(env)
				return env.Providers, nil
			}
		}
		providers = p
		if newETag != "" {
			c.storeETag(newETag)
		}
	} else {
		providers, err = c.src.Fetch(ctx)
		if err != nil {
			return nil, err
		}
	}

	env := &cacheEnvelope{
		FetchedAt: time.Now(),
		Providers: providers,
	}
	if storeErr := c.store(env); storeErr != nil {
		return providers, nil // data fetched; cache write failed
	}
	return providers, nil
}

func (c *Cache) load() (*cacheEnvelope, error) {
	data, err := os.ReadFile(filepath.Join(c.dir, dataFile))
	if err != nil {
		return nil, err
	}
	var env cacheEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		// corrupt cache: remove and fall through
		os.Remove(filepath.Join(c.dir, dataFile))
		return nil, fmt.Errorf("aim: corrupt cache: %w", err)
	}
	// re-normalize to populate json:"-" fields lost in round-trip
	for key, p := range env.Providers {
		for id, m := range p.Models {
			m.Normalize()
			if m.Provider == "" {
				m.Provider = key
			}
			p.Models[id] = m
		}
	}
	return &env, nil
}

// store writes the envelope atomically via tmp+rename.
func (c *Cache) store(env *cacheEnvelope) error {
	data, err := json.Marshal(env)
	if err != nil {
		return err
	}
	tmp := filepath.Join(c.dir, dataFile+".tmp")
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("aim: write cache tmp: %w", err)
	}
	if err := os.Rename(tmp, filepath.Join(c.dir, dataFile)); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("aim: rename cache: %w", err)
	}
	return nil
}

func (c *Cache) loadETag() string {
	data, err := os.ReadFile(filepath.Join(c.dir, etagFile))
	if err != nil {
		return ""
	}
	return string(data)
}

func (c *Cache) storeETag(etag string) {
	tmp := filepath.Join(c.dir, etagFile+".tmp")
	if err := os.WriteFile(tmp, []byte(etag), 0o644); err != nil {
		return
	}
	_ = os.Rename(tmp, filepath.Join(c.dir, etagFile))
}

// lock acquires a filesystem lock; blocks until acquired or ctx canceled.
func (c *Cache) lock(ctx context.Context) (func(), error) {
	p := filepath.Join(c.dir, lockFile)
	deadline := time.Now().Add(lockTimeout)
	for {
		f, err := os.OpenFile(p, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err == nil {
			f.Close()
			return func() { os.Remove(p) }, nil
		}
		if !errors.Is(err, os.ErrExist) {
			return nil, fmt.Errorf("aim: acquire lock: %w", err)
		}
		// break stale locks older than lockTimeout
		if info, statErr := os.Stat(p); statErr == nil {
			if time.Since(info.ModTime()) > lockTimeout {
				os.Remove(p)
				continue
			}
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("aim: lock timeout after %s", lockTimeout)
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(lockPollFreq):
		}
	}
}
