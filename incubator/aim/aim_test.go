package aim

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestModel_Normalize(t *testing.T) {
	tests := []struct {
		name      string
		model     Model
		wantInput []string
		wantOut   []string
		wantCtx   int
		wantMax   int
	}{
		{
			name: "flattens modalities and limit",
			model: Model{
				Modalities: &Modalities{
					Input:  []string{"text", "image"},
					Output: []string{"text"},
				},
				Limit: &Limit{Context: 200000, MaxOutput: 8192},
			},
			wantInput: []string{"text", "image"},
			wantOut:   []string{"text"},
			wantCtx:   200000,
			wantMax:   8192,
		},
		{
			name:  "nil modalities and limit",
			model: Model{},
		},
		{
			name: "nil modalities only",
			model: Model{
				Limit: &Limit{Context: 128000},
			},
			wantCtx: 128000,
		},
		{
			name: "nil limit only",
			model: Model{
				Modalities: &Modalities{Input: []string{"text"}},
			},
			wantInput: []string{"text"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.model
			m.Normalize()
			assert.Equal(t, tt.wantInput, m.Input)
			assert.Equal(t, tt.wantOut, m.Output)
			assert.Equal(t, tt.wantCtx, m.Context)
			assert.Equal(t, tt.wantMax, m.MaxOutput)
		})
	}
}
