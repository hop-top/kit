package aim

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fixtureProviders() map[string]Provider {
	return map[string]Provider{
		"anthropic": {ID: "anthropic", Name: "Anthropic", Models: map[string]Model{
			"claude-4": {
				ID: "claude-4", Name: "Claude 4", Provider: "anthropic",
				Family: "claude", Input: []string{"text", "image"},
				Output: []string{"text"}, ToolCall: true, Reasoning: true,
			},
			"claude-haiku": {
				ID: "claude-haiku", Name: "Claude Haiku", Provider: "anthropic",
				Family: "claude", Input: []string{"text"},
				Output: []string{"text"}, ToolCall: true, Reasoning: false,
			},
		}},
		"meta": {ID: "meta", Name: "Meta", Models: map[string]Model{
			"llama-4": {
				ID: "llama-4", Name: "Llama 4", Provider: "meta",
				Family: "llama", Input: []string{"text"},
				Output: []string{"text", "code"}, OpenWeights: true,
				ToolCall: false, Reasoning: false,
			},
		}},
		"openai": {ID: "openai", Name: "OpenAI", Models: map[string]Model{
			"gpt-5": {
				ID: "gpt-5", Name: "GPT-5", Provider: "openai",
				Family: "gpt", Input: []string{"text", "image", "video"},
				Output: []string{"text"}, ToolCall: true, Reasoning: true,
			},
		}},
	}
}

func newFixtureRegistry(t *testing.T) *Registry {
	t.Helper()
	src := &mockSource{providers: fixtureProviders()}
	r, err := NewRegistry(WithSource(src), WithRegistryCacheDir(t.TempDir()))
	require.NoError(t, err)
	return r
}

func TestNewRegistry_Default(t *testing.T) {
	// Just verify construction succeeds with no sources (uses ModelsDevSource).
	r, err := NewRegistry(WithRegistryCacheDir(t.TempDir()))
	require.NoError(t, err)
	require.NotNil(t, r)
}

func TestProviders_SortedByID(t *testing.T) {
	r := newFixtureRegistry(t)
	ps := r.Providers()
	require.Len(t, ps, 3)
	assert.Equal(t, "anthropic", ps[0].ID)
	assert.Equal(t, "meta", ps[1].ID)
	assert.Equal(t, "openai", ps[2].ID)
}

func TestGet_HappyPath(t *testing.T) {
	r := newFixtureRegistry(t)
	m, ok := r.Get("anthropic", "claude-4")
	require.True(t, ok)
	assert.Equal(t, "Claude 4", m.Name)
}

func TestGet_MissingProvider(t *testing.T) {
	r := newFixtureRegistry(t)
	_, ok := r.Get("google", "gemini")
	assert.False(t, ok)
}

func TestGet_MissingModel(t *testing.T) {
	r := newFixtureRegistry(t)
	_, ok := r.Get("anthropic", "nonexistent")
	assert.False(t, ok)
}

func TestModels_FilterProvider(t *testing.T) {
	r := newFixtureRegistry(t)
	ms := r.Models(Filter{Provider: "meta"})
	require.Len(t, ms, 1)
	assert.Equal(t, "llama-4", ms[0].ID)
}

func TestModels_FilterFamily(t *testing.T) {
	r := newFixtureRegistry(t)
	ms := r.Models(Filter{Family: "claude"})
	require.Len(t, ms, 2)
}

func TestModels_FilterModality(t *testing.T) {
	r := newFixtureRegistry(t)
	// in:image — models accepting image input
	ms := r.Models(Filter{Input: "image"})
	ids := make([]string, len(ms))
	for i, m := range ms {
		ids[i] = m.ID
	}
	assert.Contains(t, ids, "claude-4")
	assert.Contains(t, ids, "gpt-5")
	assert.NotContains(t, ids, "llama-4")
}

func TestModels_FilterModalitySubset(t *testing.T) {
	r := newFixtureRegistry(t)
	// out:text,code — only models outputting both
	ms := r.Models(Filter{Output: "text,code"})
	require.Len(t, ms, 1)
	assert.Equal(t, "llama-4", ms[0].ID)
}

func TestModels_FilterBoolTrue(t *testing.T) {
	r := newFixtureRegistry(t)
	tr := true
	ms := r.Models(Filter{Reasoning: &tr})
	for _, m := range ms {
		assert.True(t, m.Reasoning, m.ID)
	}
	assert.Len(t, ms, 2) // claude-4, gpt-5
}

func TestModels_FilterBoolFalse(t *testing.T) {
	r := newFixtureRegistry(t)
	fa := false
	ms := r.Models(Filter{OpenWeights: &fa})
	for _, m := range ms {
		assert.False(t, m.OpenWeights, m.ID)
	}
}

func TestModels_FilterFreeText(t *testing.T) {
	r := newFixtureRegistry(t)
	ms := r.Models(Filter{Query: "haiku"})
	require.Len(t, ms, 1)
	assert.Equal(t, "claude-haiku", ms[0].ID)
}

func TestModels_FilterFreeTextCaseInsensitive(t *testing.T) {
	r := newFixtureRegistry(t)
	ms := r.Models(Filter{Query: "GPT"})
	require.Len(t, ms, 1)
	assert.Equal(t, "gpt-5", ms[0].ID)
}

func TestQuery_Delegates(t *testing.T) {
	r := newFixtureRegistry(t)
	ms, err := r.Query("provider:anthropic")
	require.NoError(t, err)
	require.Len(t, ms, 2)
}

func TestQuery_ParseError(t *testing.T) {
	r := newFixtureRegistry(t)
	_, err := r.Query("foo:bar")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown tag")
}

func TestModalitySubset_SupersetNoMatch(t *testing.T) {
	// Filter requires image+video, model only has text
	assert.False(t, modalitySubset("image,video", []string{"text"}))
}

func TestModalitySubset_ExactMatch(t *testing.T) {
	assert.True(t, modalitySubset("text", []string{"text"}))
}

func TestModalitySubset_SubsetMatch(t *testing.T) {
	assert.True(t, modalitySubset("text", []string{"text", "image"}))
}

func TestMultiSource_LastWins(t *testing.T) {
	s1 := &mockSource{providers: map[string]Provider{
		"p": {ID: "p", Name: "First", Models: map[string]Model{}},
	}}
	s2 := &mockSource{providers: map[string]Provider{
		"p": {ID: "p", Name: "Second", Models: map[string]Model{}},
	}}
	r, err := NewRegistry(
		WithSource(s1), WithSource(s2),
		WithRegistryCacheDir(t.TempDir()),
	)
	require.NoError(t, err)
	ps := r.Providers()
	require.Len(t, ps, 1)
	assert.Equal(t, "Second", ps[0].Name)
}
