package llm

import (
	"time"

	"hop.top/kit/go/runtime/bus"
)

// Bus topic constants for llm lifecycle events.
const (
	TopicRequestStart bus.Topic = "llm.request.start"
	TopicRequestEnd   bus.Topic = "llm.request.end"
	TopicRequestError bus.Topic = "llm.request.error"
	TopicFallback     bus.Topic = "llm.fallback"
	TopicRoute        bus.Topic = "llm.route"
	TopicEvaResult    bus.Topic = "llm.eva.result"
)

// RequestStartPayload is published before each LLM call.
type RequestStartPayload struct {
	Request Request
}

// RequestEndPayload is published after a successful LLM call.
type RequestEndPayload struct {
	Response Response
	Duration time.Duration
}

// RequestErrorPayload is published when an LLM call fails terminally.
type RequestErrorPayload struct {
	Err        error  `json:"-"`
	ErrMessage string `json:"error"`
}

// FallbackPayload is published when falling back to the next provider.
type FallbackPayload struct {
	From       int
	To         int
	Err        error  `json:"-"`
	ErrMessage string `json:"error"`
}

// RoutePayload is published after a routing decision.
type RoutePayload struct {
	Router string
	Score  float64
	Model  string
}

// EvaResultPayload is published after an eva contract evaluation.
type EvaResultPayload struct {
	Contract   string
	Passed     bool
	Violations []string
}
