package aim

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	defaultURL         = "https://models.dev/api.json"
	defaultTimeout     = 30 * time.Second
	defaultMaxRespSize = 50 << 20 // 50 MB
)

// SourceOption configures a ModelsDevSource.
type SourceOption func(*ModelsDevSource)

// WithURL overrides the default API endpoint.
func WithURL(url string) SourceOption {
	return func(s *ModelsDevSource) { s.url = url }
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(c *http.Client) SourceOption {
	return func(s *ModelsDevSource) { s.client = c }
}

// WithTimeout sets the HTTP client timeout.
func WithTimeout(d time.Duration) SourceOption {
	return func(s *ModelsDevSource) { s.client.Timeout = d }
}

// WithMaxResponseSize limits the response body size in bytes.
func WithMaxResponseSize(n int64) SourceOption {
	return func(s *ModelsDevSource) { s.maxSize = n }
}

// ModelsDevSource fetches the provider catalog from models.dev.
type ModelsDevSource struct {
	url     string
	client  *http.Client
	maxSize int64
}

// NewModelsDevSource returns a Source backed by models.dev/api.json.
func NewModelsDevSource(opts ...SourceOption) *ModelsDevSource {
	s := &ModelsDevSource{
		url:     defaultURL,
		client:  &http.Client{Timeout: defaultTimeout},
		maxSize: defaultMaxRespSize,
	}
	for _, o := range opts {
		o(s)
	}
	return s
}

// Fetch retrieves and validates the provider catalog.
func (s *ModelsDevSource) Fetch(ctx context.Context) (map[string]Provider, error) {
	providers, _, _, err := s.FetchWithETag(ctx, "")
	return providers, err
}

// ETag returns the ETag from a conditional GET (used by Cache).
func (s *ModelsDevSource) FetchWithETag(ctx context.Context, etag string) (map[string]Provider, string, bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.url, nil)
	if err != nil {
		return nil, "", false, fmt.Errorf("aim: build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if etag != "" {
		req.Header.Set("If-None-Match", etag)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, "", false, fmt.Errorf("aim: fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotModified {
		return nil, etag, true, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, "", false, fmt.Errorf("aim: unexpected status %d from %s", resp.StatusCode, s.url)
	}

	r := io.LimitReader(resp.Body, s.maxSize+1)
	body, err := io.ReadAll(r)
	if err != nil {
		return nil, "", false, fmt.Errorf("aim: read body: %w", err)
	}
	if int64(len(body)) > s.maxSize {
		return nil, "", false, fmt.Errorf("aim: response exceeds max size (%d bytes)", s.maxSize)
	}

	var providers map[string]Provider
	if err := json.Unmarshal(body, &providers); err != nil {
		return nil, "", false, fmt.Errorf("aim: decode json: %w", err)
	}

	for key, p := range providers {
		if key != p.ID {
			return nil, "", false, fmt.Errorf("aim: map key %q != provider.id %q", key, p.ID)
		}
		for id, m := range p.Models {
			m.Normalize()
			if m.Provider == "" {
				m.Provider = key
			}
			p.Models[id] = m
		}
	}

	return providers, resp.Header.Get("ETag"), false, nil
}
