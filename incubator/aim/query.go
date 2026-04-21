package aim

import (
	"fmt"
	"strings"
)

// ParseQuery parses a human query string into a Filter.
// Tokens split by whitespace; quoted strings → free-text literals.
// Unquoted tokens with ":" → key:value tags.
func ParseQuery(q string) (Filter, error) {
	var f Filter
	var free []string
	for i, n := 0, len(q); i < n; {
		if q[i] == ' ' || q[i] == '\t' {
			i++
			continue
		}
		var val string
		var quoted bool
		if q[i] == '"' {
			j := strings.IndexByte(q[i+1:], '"')
			if j < 0 {
				return Filter{}, fmt.Errorf("unterminated quote")
			}
			val, quoted = q[i+1:i+1+j], true
			i += j + 2
		} else {
			j := strings.IndexAny(q[i:], " \t")
			if j < 0 {
				j = n - i
			}
			val = q[i : i+j]
			i += j
		}
		if quoted {
			free = append(free, val)
			continue
		}
		ci := strings.IndexByte(val, ':')
		if ci < 0 {
			free = append(free, val)
			continue
		}
		key, v := val[:ci], val[ci+1:]
		if key == "" {
			return Filter{}, fmt.Errorf("malformed tag %q: missing key", val)
		}
		if v == "" {
			return Filter{}, fmt.Errorf("empty value for tag %q", key)
		}
		switch key {
		case "provider":
			f.Provider = appendCSV(f.Provider, v)
		case "family":
			f.Family = appendCSV(f.Family, v)
		case "in":
			f.Input = appendCSV(f.Input, v)
		case "out":
			f.Output = appendCSV(f.Output, v)
		case "toolcall", "reasoning", "openweights":
			b, err := parseBool(v, key)
			if err != nil {
				return Filter{}, err
			}
			switch key {
			case "toolcall":
				f.ToolCall = &b
			case "reasoning":
				f.Reasoning = &b
			case "openweights":
				f.OpenWeights = &b
			}
		default:
			return Filter{}, fmt.Errorf("unknown tag %q", key)
		}
	}
	if len(free) > 0 {
		f.Query = strings.Join(free, " ")
	}
	return f, nil
}

func parseBool(v, tag string) (bool, error) {
	if v == "true" {
		return true, nil
	}
	if v == "false" {
		return false, nil
	}
	return false, fmt.Errorf("invalid bool value %q for tag %q: must be true or false", v, tag)
}

func appendCSV(existing, val string) string {
	if existing == "" {
		return val
	}
	return existing + "," + val
}
