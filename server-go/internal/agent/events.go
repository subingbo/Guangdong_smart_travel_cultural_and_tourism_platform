package agent

import "encoding/json"

type EventType string

const (
	EventSnapshot   EventType = "snapshot"
	EventDelta      EventType = "delta"
	EventThinking   EventType = "thinking"
	EventToolCall   EventType = "tool_call"
	EventToolResult EventType = "tool_result"
	EventPlan       EventType = "plan"
	EventDone       EventType = "done"
	EventError      EventType = "error"
	EventHeartbeat  EventType = "heartbeat"
)

type Event struct {
	Type  EventType `json:"type"`
	Seq   int64     `json:"seq"`
	RunID int64     `json:"runId,omitempty"`
	Data  any       `json:"data,omitempty"`
}

// EmitFunc is the runner's only transport dependency.
type EmitFunc func(EventType, any)

func (f EmitFunc) Emit(t EventType, data any) {
	if f != nil {
		f(t, data)
	}
}

type DeltaPayload struct {
	Content string `json:"content"`
}

type ThinkingPayload struct {
	Content string `json:"content"`
}

type ToolCallPayload struct {
	Step      int    `json:"step"`
	Name      string `json:"name"`
	Arguments string `json:"arguments,omitempty"`
}

type ToolResultPayload struct {
	Step      int    `json:"step"`
	Name      string `json:"name"`
	OK        bool   `json:"ok"`
	LatencyMS int    `json:"latencyMs"`
	Preview   string `json:"preview,omitempty"`
	Error     string `json:"error,omitempty"`
}

type PlanPayload struct {
	Version int             `json:"version"`
	Title   string          `json:"title"`
	Summary string          `json:"summary,omitempty"`
	Plan    json.RawMessage `json:"plan"`
}

type DonePayload struct {
	Status    string `json:"status"`
	StepCount int    `json:"stepCount"`
	Usage     Usage  `json:"usage"`
	Plan      bool   `json:"plan"`
}

type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
