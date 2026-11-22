// Package agent contains a small, dependency-light ReAct loop, tool
// registry, run executor and SSE event model. It knows nothing about the
// travel domain: tools are injected by the service layer.
package agent

import (
	"context"
	"encoding/json"
	"sync"
	"time"
)

type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

type Message struct {
	Role       Role       `json:"role"`
	Content    string     `json:"content,omitempty"`
	ToolCalls  []ToolCall `json:"toolCalls,omitempty"`
	ToolCallID string     `json:"toolCallId,omitempty"`
	Name       string     `json:"name,omitempty"`
}

type ToolCall struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type ToolSpec struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

type Usage struct {
	PromptTokens     int `json:"promptTokens"`
	CompletionTokens int `json:"completionTokens"`
	TotalTokens      int `json:"totalTokens"`
}

type ChatRequest struct {
	Model             string
	Messages          []Message
	Tools             []ToolSpec
	ParallelToolCalls bool
	Temperature       float64
}

type ChatResponse struct {
	Content      string
	ToolCalls    []ToolCall
	FinishReason string
	Usage        Usage
}

// Delta is an incremental chunk delivered while the provider streams.
type Delta struct {
	Content   string
	Reasoning string
	ToolCall  *ToolCallDelta
}

type ToolCallDelta struct {
	Index     int
	ID        string
	Name      string
	Arguments string
}

// Provider is implemented by the OpenAI-compatible client (ai package) and
// by test fakes.
type Provider interface {
	StreamChat(ctx context.Context, req ChatRequest, onDelta func(Delta) error) (ChatResponse, error)
}

// ExecutionContext is the trusted identity handed to tools. UserID must come
// from the authenticated request, never from model arguments.
type ExecutionContext struct {
	RunID          int64
	ConversationID int64
	UserID         int64
	Language       string
	PlanVersion    int // conversation plan_version when the run started
	State          *RunState
}

// RunState is per-run scratch space shared by tools. It is deliberately
// untyped so the framework stays domain-agnostic.
type RunState struct {
	mu     sync.Mutex
	values map[string]any
	counts map[string]int
	seen   map[string]struct{}
}

func NewRunState() *RunState {
	return &RunState{
		values: make(map[string]any),
		counts: make(map[string]int),
		seen:   make(map[string]struct{}),
	}
}

func (s *RunState) Set(key string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values[key] = value
}

func (s *RunState) Get(key string) (any, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.values[key]
	return v, ok
}

func (s *RunState) Incr(key string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counts[key]++
	return s.counts[key]
}

func (s *RunState) Count(key string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.counts[key]
}

// MarkSeen returns true when the key was already present.
func (s *RunState) MarkSeen(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.seen[key]; ok {
		return true
	}
	s.seen[key] = struct{}{}
	return false
}

// PlanDraft is a server-owned plan snapshot produced by submit_plan.
type PlanDraft struct {
	Version int             `json:"version"`
	Title   string          `json:"title"`
	Summary string          `json:"summary,omitempty"`
	Plan    json.RawMessage `json:"plan"`
}

type ToolResult struct {
	// Content is what the model sees. The runner projects it before sending.
	Content string
	// Plan is non-nil when the tool produced a draft; the runner emits it
	// as a separate SSE event so the UI can render a card.
	Plan *PlanDraft
}

type Tool interface {
	Name() string
	Description() string
	Parameters() json.RawMessage
	Execute(ctx context.Context, ec *ExecutionContext, args json.RawMessage) (ToolResult, error)
}

// StepRecord is the minimal trace unit persisted as ai_steps.
type StepRecord struct {
	StepNo    int
	Type      string
	Name      string
	Arguments string
	Result    string
	Latency   time.Duration
	Error     string
}
