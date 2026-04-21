package parity_test

import (
	"testing"

	"hop.top/kit/contracts/parity"
)

func TestParityStatusSymbols(t *testing.T) {
	kinds := []string{"info", "success", "error", "warn"}
	for _, k := range kinds {
		sym, ok := parity.Values.Status.Symbols[k]
		if !ok || sym == "" {
			t.Errorf("status.symbols[%q]: missing or empty", k)
		}
	}
}

func TestParityStatusSymbolValues(t *testing.T) {
	// Pin the exact values so any accidental change fails the test.
	want := map[string]string{
		"info":    "ℹ",
		"success": "✓",
		"error":   "●",
		"warn":    "▲",
	}
	for k, v := range want {
		if got := parity.Values.Status.Symbols[k]; got != v {
			t.Errorf("status.symbols[%q] = %q, want %q", k, got, v)
		}
	}
}

func TestParitySpinnerFrames(t *testing.T) {
	frames := parity.Values.Spinner.Frames
	if len(frames) == 0 {
		t.Fatal("spinner.frames: empty")
	}
	if parity.Values.Spinner.IntervalMs <= 0 {
		t.Errorf("spinner.interval_ms = %d, want > 0", parity.Values.Spinner.IntervalMs)
	}
}

func TestParityAnimRunes(t *testing.T) {
	if parity.Values.Anim.Runes == "" {
		t.Fatal("anim.runes: empty")
	}
	if parity.Values.Anim.IntervalMs <= 0 {
		t.Errorf("anim.interval_ms = %d, want > 0", parity.Values.Anim.IntervalMs)
	}
	if parity.Values.Anim.DefaultWidth <= 0 {
		t.Errorf("anim.default_width = %d, want > 0", parity.Values.Anim.DefaultWidth)
	}
}
