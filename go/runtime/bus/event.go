package bus

import (
	"strings"
	"time"
)

// Topic is a dot-separated event path (e.g. "llm.request.start").
type Topic string

// Match reports whether pattern matches topic t.
//
// Rules:
//   - Exact match: "llm.request" matches "llm.request"
//   - `*` matches one segment: "llm.*" matches "llm.request" but not "llm.request.start"
//   - `#` matches zero or more trailing segments: "llm.#" matches "llm", "llm.request", "llm.request.start"
func (t Topic) Match(pattern string) bool {
	tParts := strings.Split(string(t), ".")
	pParts := strings.Split(pattern, ".")
	return matchParts(tParts, pParts)
}

func matchParts(topic, pattern []string) bool {
	ti, pi := 0, 0
	for pi < len(pattern) {
		if pattern[pi] == "#" {
			// Per MQTT convention, # must be the last segment.
			if pi != len(pattern)-1 {
				return false
			}
			return true
		}
		if ti >= len(topic) {
			return false
		}
		if pattern[pi] != "*" && pattern[pi] != topic[ti] {
			return false
		}
		ti++
		pi++
	}
	return ti == len(topic)
}

// Event is the standard envelope for all bus messages.
type Event struct {
	// Topic identifies the event type (e.g. "llm.request").
	Topic Topic
	// Source identifies the emitter (e.g. "llm.client", "tool.exec").
	Source string
	// Timestamp is when the event was created.
	Timestamp time.Time
	// Payload carries event-specific data.
	Payload any
}

// NewEvent creates an Event with the current timestamp.
func NewEvent(topic Topic, source string, payload any) Event {
	return Event{
		Topic:     topic,
		Source:    source,
		Timestamp: time.Now(),
		Payload:   payload,
	}
}
