package aim

import "context"

// Provider represents an AI provider (e.g. "anthropic", "openai").
type Provider struct {
	ID     string           `json:"id"`
	Name   string           `json:"name"`
	Models map[string]Model `json:"models"` // key = Model.ID
}

// Model represents an AI model as described by the models.dev wire format.
type Model struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Provider    string   `json:"provider,omitempty"`
	Family      string   `json:"family,omitempty"`
	Input       []string `json:"-"` // populated from modalities.input
	Output      []string `json:"-"` // populated from modalities.output
	ToolCall    bool     `json:"tool_call"`
	Reasoning   bool     `json:"reasoning"`
	OpenWeights bool     `json:"open_weights"`
	Context     int      `json:"-"` // populated from limit.context
	MaxOutput   int      `json:"-"` // populated from limit.output
	CostInput   float64  `json:"-"` // populated from cost.input ($/M tokens)
	CostOutput  float64  `json:"-"` // populated from cost.output ($/M tokens)

	// Wire-only nested objects; exported for json round-trip.
	Modalities *Modalities `json:"modalities,omitempty"`
	Limit      *Limit      `json:"limit,omitempty"`
	Cost       *Cost       `json:"cost,omitempty"`
}

// Modalities maps to the wire format's modalities object.
type Modalities struct {
	Input  []string `json:"input,omitempty"`
	Output []string `json:"output,omitempty"`
}

// Limit maps to the wire format's limit object.
type Limit struct {
	Context   int `json:"context,omitempty"`
	MaxOutput int `json:"output,omitempty"`
}

// Cost maps to the wire format's cost object (USD per 1M tokens).
type Cost struct {
	Input  float64 `json:"input,omitempty"`
	Output float64 `json:"output,omitempty"`
}

// Normalize flattens nested wire fields (Modalities, Limit) into
// the top-level Model fields. Call after JSON unmarshalling.
func (m *Model) Normalize() {
	if m.Modalities != nil {
		m.Input = m.Modalities.Input
		m.Output = m.Modalities.Output
	}
	if m.Limit != nil {
		m.Context = m.Limit.Context
		m.MaxOutput = m.Limit.MaxOutput
	}
	if m.Cost != nil {
		m.CostInput = m.Cost.Input
		m.CostOutput = m.Cost.Output
	}
}

// Filter is a query filter for registry lookups.
//
// Tristate *bool fields: nil = don't filter, true = must have,
// false = must not have. Equivalent to boolean|undefined (TS) or
// Optional[bool] (Python).
type Filter struct {
	Input       string
	Output      string
	Provider    string
	Family      string
	ToolCall    *bool
	Reasoning   *bool
	OpenWeights *bool
	Query       string // free-text search
}

// Source fetches the full provider catalog. Map key MUST equal Provider.ID.
type Source interface {
	Fetch(ctx context.Context) (map[string]Provider, error)
}

// ETagSource extends Source with conditional-GET support.
type ETagSource interface {
	FetchWithETag(ctx context.Context, etag string) (map[string]Provider, string, bool, error)
}
