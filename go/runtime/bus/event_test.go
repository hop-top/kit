package bus

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTopic_Match_Exact(t *testing.T) {
	topic := Topic("llm.request")
	if !topic.Match("llm.request") {
		t.Error("exact match should succeed")
	}
	if topic.Match("llm.response") {
		t.Error("different topic should not match")
	}
}

func TestTopic_Match_SingleWildcard(t *testing.T) {
	topic := Topic("llm.request")
	if !topic.Match("llm.*") {
		t.Error("llm.* should match llm.request")
	}
	if !topic.Match("*.request") {
		t.Error("*.request should match llm.request")
	}

	deep := Topic("llm.request.start")
	if deep.Match("llm.*") {
		t.Error("llm.* should NOT match llm.request.start (too deep)")
	}
}

func TestTopic_Match_MultiWildcard(t *testing.T) {
	cases := []struct {
		topic   Topic
		pattern string
		want    bool
	}{
		{"llm.request", "llm.#", true},
		{"llm.request.start", "llm.#", true},
		{"llm", "llm.#", true},
		{"tool.exec", "llm.#", false},
		{"llm.request.start", "#", true},
		{"anything", "#", true},
	}
	for _, tc := range cases {
		got := tc.topic.Match(tc.pattern)
		if got != tc.want {
			t.Errorf("Topic(%q).Match(%q) = %v, want %v",
				tc.topic, tc.pattern, got, tc.want)
		}
	}
}

func TestEvent_JSONRoundTrip(t *testing.T) {
	type payload struct {
		Model string `json:"model"`
	}
	e := Event{
		Topic:     "llm.request",
		Source:    "test",
		Timestamp: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		Payload:   payload{Model: "claude-4"},
	}

	data, err := json.Marshal(e)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded struct {
		Topic     string          `json:"Topic"`
		Source    string          `json:"Source"`
		Timestamp time.Time       `json:"Timestamp"`
		Payload   json.RawMessage `json:"Payload"`
	}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.Topic != "llm.request" {
		t.Errorf("topic = %q, want llm.request", decoded.Topic)
	}
	if decoded.Source != "test" {
		t.Errorf("source = %q, want test", decoded.Source)
	}

	var p payload
	if err := json.Unmarshal(decoded.Payload, &p); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if p.Model != "claude-4" {
		t.Errorf("model = %q, want claude-4", p.Model)
	}
}

func TestNewEvent_SetsTimestamp(t *testing.T) {
	before := time.Now()
	e := NewEvent("test.topic", "src", nil)
	after := time.Now()

	if e.Timestamp.Before(before) || e.Timestamp.After(after) {
		t.Error("timestamp should be between before and after")
	}
}
