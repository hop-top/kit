package aim

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseQuery_Vectors(t *testing.T) {
	data, err := os.ReadFile("testdata/query-vectors.json")
	require.NoError(t, err)

	var vectors []struct {
		Input       string           `json:"input"`
		Expected    *json.RawMessage `json:"expected"`
		Error       string           `json:"error"`
		Description string           `json:"description"`
	}
	require.NoError(t, json.Unmarshal(data, &vectors))
	require.NotEmpty(t, vectors)

	for _, v := range vectors {
		t.Run(v.Description, func(t *testing.T) {
			got, err := ParseQuery(v.Input)

			if v.Error != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), v.Error)
				return
			}

			require.NoError(t, err)

			// Decode expected into a map to handle tristate bools.
			var exp map[string]interface{}
			require.NoError(t, json.Unmarshal(*v.Expected, &exp))

			strField := func(key string) string {
				if v, ok := exp[key]; ok {
					return v.(string)
				}
				return ""
			}
			assert.Equal(t, strField("Provider"), got.Provider, "Provider")
			assert.Equal(t, strField("Family"), got.Family, "Family")
			assert.Equal(t, strField("Input"), got.Input, "Input")
			assert.Equal(t, strField("Output"), got.Output, "Output")
			assert.Equal(t, strField("Query"), got.Query, "Query")

			boolField := func(key string) *bool {
				v, ok := exp[key]
				if !ok {
					return nil
				}
				b := v.(bool)
				return &b
			}
			assert.Equal(t, boolField("ToolCall"), got.ToolCall, "ToolCall")
			assert.Equal(t, boolField("Reasoning"), got.Reasoning, "Reasoning")
			assert.Equal(t, boolField("OpenWeights"), got.OpenWeights, "OpenWeights")
		})
	}
}
