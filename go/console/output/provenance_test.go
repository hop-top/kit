package output

import (
	"testing"
	"time"
)

func TestWithProvenance_AddsMetaField(t *testing.T) {
	data := map[string]string{"name": "test"}
	p := Provenance{
		Source:    "local",
		Timestamp: time.Date(2026, 4, 17, 12, 0, 0, 0, time.UTC),
		Method:    "static",
	}
	result := WithProvenance(data, p)

	if _, ok := result["data"]; !ok {
		t.Error("missing 'data' key")
	}
	meta, ok := result["_meta"]
	if !ok {
		t.Fatal("missing '_meta' key")
	}
	m, ok := meta.(map[string]any)
	if !ok {
		t.Fatalf("_meta is %T, want map[string]any", meta)
	}
	if m["source"] != "local" {
		t.Errorf("source = %v, want local", m["source"])
	}
	if m["method"] != "static" {
		t.Errorf("method = %v, want static", m["method"])
	}
}

func TestWithProvenance_TimestampFormat(t *testing.T) {
	ts := time.Date(2026, 4, 17, 12, 30, 45, 0, time.UTC)
	p := Provenance{
		Source:    "api",
		Timestamp: ts,
		Method:    "query",
	}
	result := WithProvenance("anything", p)
	meta := result["_meta"].(map[string]any)
	got := meta["timestamp"]
	want := "2026-04-17T12:30:45Z"
	if got != want {
		t.Errorf("timestamp = %q, want %q", got, want)
	}
}
