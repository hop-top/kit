package aim

import (
	"context"
	"sort"
	"strings"
	"time"
)

type RegistryOption func(*Registry)

func WithSource(s Source) RegistryOption {
	return func(r *Registry) { r.sources = append(r.sources, s) }
}

func WithRegistryCacheDir(dir string) RegistryOption {
	return func(r *Registry) { r.cacheDir = dir }
}

func WithRegistryTTL(d time.Duration) RegistryOption {
	return func(r *Registry) { r.cacheTTL = d }
}

// Registry is the core query engine over one or more model sources.
type Registry struct {
	sources  []Source
	cache    *Cache
	cacheDir string
	cacheTTL time.Duration

	providers map[string]Provider // lazy-loaded
}

// NewRegistry creates a Registry. Defaults to ModelsDevSource when
// no sources provided.
func NewRegistry(opts ...RegistryOption) (*Registry, error) {
	r := &Registry{}
	for _, o := range opts {
		o(r)
	}
	if len(r.sources) == 0 {
		r.sources = []Source{NewModelsDevSource()}
	}

	var cacheOpts []CacheOption
	if r.cacheDir != "" {
		cacheOpts = append(cacheOpts, WithCacheDir(r.cacheDir))
	}
	if r.cacheTTL > 0 {
		cacheOpts = append(cacheOpts, WithTTL(r.cacheTTL))
	}

	// Cache wraps a merged multi-source.
	c, err := NewCache(&multiSource{sources: r.sources}, cacheOpts...)
	if err != nil {
		return nil, err
	}
	r.cache = c
	return r, nil
}

// ensure lazy-loads providers. Errors result in empty results; use
// Query (which returns error) or Refresh to surface fetch failures.
func (r *Registry) ensure() {
	if r.providers != nil {
		return
	}
	p, err := r.cache.Fetch(context.Background())
	if err != nil {
		r.providers = map[string]Provider{}
		return
	}
	r.providers = p
}

func (r *Registry) Refresh(ctx context.Context) error {
	p, err := r.cache.Refresh(ctx, true)
	if err != nil {
		return err
	}
	r.providers = p
	return nil
}

func (r *Registry) Providers() []Provider {
	r.ensure()
	out := make([]Provider, 0, len(r.providers))
	for _, p := range r.providers {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// Models returns all models matching f (AND logic across fields).
func (r *Registry) Models(f Filter) []Model {
	r.ensure()
	var out []Model
	for _, p := range r.providers {
		for _, m := range p.Models {
			if matchesFilter(m, f) {
				out = append(out, m)
			}
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Provider != out[j].Provider {
			return out[i].Provider < out[j].Provider
		}
		return out[i].ID < out[j].ID
	})
	return out
}

func (r *Registry) Get(provider, model string) (Model, bool) {
	r.ensure()
	p, ok := r.providers[provider]
	if !ok {
		return Model{}, false
	}
	m, ok := p.Models[model]
	return m, ok
}

func (r *Registry) Query(q string) ([]Model, error) {
	f, err := ParseQuery(q)
	if err != nil {
		return nil, err
	}
	return r.Models(f), nil
}

func matchesFilter(m Model, f Filter) bool {
	if f.Provider != "" && !csvContains(f.Provider, m.Provider) {
		return false
	}
	if f.Family != "" && !csvContains(f.Family, m.Family) {
		return false
	}
	if f.Input != "" && !modalitySubset(f.Input, m.Input) {
		return false
	}
	if f.Output != "" && !modalitySubset(f.Output, m.Output) {
		return false
	}
	if f.ToolCall != nil && m.ToolCall != *f.ToolCall {
		return false
	}
	if f.Reasoning != nil && m.Reasoning != *f.Reasoning {
		return false
	}
	if f.OpenWeights != nil && m.OpenWeights != *f.OpenWeights {
		return false
	}
	if f.Query != "" {
		q := strings.ToLower(f.Query)
		if !strings.Contains(strings.ToLower(m.ID), q) &&
			!strings.Contains(strings.ToLower(m.Name), q) {
			return false
		}
	}
	return true
}

// modalitySubset: every comma-separated value in filter must appear
// in the model's modality list.
func modalitySubset(filter string, modalities []string) bool {
	set := make(map[string]struct{}, len(modalities))
	for _, v := range modalities {
		set[strings.ToLower(strings.TrimSpace(v))] = struct{}{}
	}
	for _, want := range strings.Split(filter, ",") {
		want = strings.ToLower(strings.TrimSpace(want))
		if want == "" {
			continue
		}
		if _, ok := set[want]; !ok {
			return false
		}
	}
	return true
}

func csvContains(csv, val string) bool {
	for _, v := range strings.Split(csv, ",") {
		if strings.TrimSpace(v) == val {
			return true
		}
	}
	return false
}

// multiSource merges results from multiple sources. When providers
// share the same ID, later sources override earlier ones (last wins).
type multiSource struct {
	sources []Source
}

func (ms *multiSource) Fetch(ctx context.Context) (map[string]Provider, error) {
	merged := make(map[string]Provider)
	for _, s := range ms.sources {
		providers, err := s.Fetch(ctx)
		if err != nil {
			return nil, err
		}
		for k, v := range providers {
			merged[k] = v
		}
	}
	return merged, nil
}
