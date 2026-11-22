package agent

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
)

type fakeProvider struct {
	mu        sync.Mutex
	responses []ChatResponse
	requests  []ChatRequest
	block     bool
}

func (f *fakeProvider) StreamChat(ctx context.Context, req ChatRequest, onDelta func(Delta) error) (ChatResponse, error) {
	f.mu.Lock()
	f.requests = append(f.requests, req)
	if f.block {
		f.mu.Unlock()
		<-ctx.Done()
		return ChatResponse{}, ctx.Err()
	}
	if len(f.responses) == 0 {
		f.mu.Unlock()
		return ChatResponse{}, errors.New("no scripted response")
	}
	resp := f.responses[0]
	f.responses = f.responses[1:]
	f.mu.Unlock()

	if resp.Content != "" && onDelta != nil {
		if err := onDelta(Delta{Content: resp.Content}); err != nil {
			return ChatResponse{}, err
		}
	}
	return resp, nil
}

func (f *fakeProvider) lastRequest() ChatRequest {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.requests) == 0 {
		return ChatRequest{}
	}
	return f.requests[len(f.requests)-1]
}

func textTool(name, reply string) Tool {
	return NewTool(name, "test tool", json.RawMessage(`{"type":"object","properties":{}}`), func(ctx context.Context, ec *ExecutionContext, args json.RawMessage) (ToolResult, error) {
		return ToolResult{Content: reply}, nil
	})
}

func collectEvents() (EmitFunc, *[]EventRecord) {
	var mu sync.Mutex
	records := make([]EventRecord, 0)
	emit := func(t EventType, data any) {
		mu.Lock()
		defer mu.Unlock()
		records = append(records, EventRecord{Type: t, Data: data})
	}
	return emit, &records
}

type EventRecord struct {
	Type EventType
	Data any
}

func TestRunExecutesToolThenAnswers(t *testing.T) {
	provider := &fakeProvider{responses: []ChatResponse{
		{ToolCalls: []ToolCall{{ID: "call-1", Name: "echo", Arguments: `{"q":"x"}`}}},
		{Content: "done"},
	}}
	registry, err := NewRegistry(textTool("echo", `{"ok":true}`))
	if err != nil {
		t.Fatal(err)
	}
	emit, records := collectEvents()
	runner := NewRunner(provider, registry, DefaultLimits())

	result, err := runner.Run(context.Background(), RunRequest{
		UserID:       7,
		SystemPrompt: "system",
		UserMessage:  "plan",
		Emit:         emit,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Output != "done" || result.StepCount != 2 {
		t.Fatalf("unexpected result: %+v", result)
	}
	last := provider.lastRequest()
	foundTool := false
	for _, msg := range last.Messages {
		if msg.Role == RoleTool && msg.ToolCallID == "call-1" && msg.Content == `{"ok":true}` {
			foundTool = true
		}
	}
	if !foundTool {
		t.Fatalf("tool result was not appended: %+v", last.Messages)
	}
	seen := map[EventType]bool{}
	for _, rec := range *records {
		seen[rec.Type] = true
	}
	for _, want := range []EventType{EventToolCall, EventToolResult, EventDelta} {
		if !seen[want] {
			t.Fatalf("missing event %s: %+v", want, *records)
		}
	}
}

func TestRunExecutesMultipleToolCallsInOneStep(t *testing.T) {
	provider := &fakeProvider{responses: []ChatResponse{
		{ToolCalls: []ToolCall{
			{ID: "a", Name: "echo", Arguments: `{"v":1}`},
			{ID: "b", Name: "echo2", Arguments: `{"v":2}`},
		}},
		{Content: "ok"},
	}}
	registry, err := NewRegistry(textTool("echo", "one"), textTool("echo2", "two"))
	if err != nil {
		t.Fatal(err)
	}
	runner := NewRunner(provider, registry, DefaultLimits())
	if _, err := runner.Run(context.Background(), RunRequest{SystemPrompt: "s", UserMessage: "u"}); err != nil {
		t.Fatal(err)
	}
	last := provider.lastRequest()
	count := 0
	for _, msg := range last.Messages {
		if msg.Role == RoleTool {
			count++
		}
	}
	if count != 2 {
		t.Fatalf("expected two tool messages, got %d: %+v", count, last.Messages)
	}
}

func TestRunStopsAtMaxSteps(t *testing.T) {
	provider := &fakeProvider{responses: []ChatResponse{
		{ToolCalls: []ToolCall{{ID: "1", Name: "echo", Arguments: "{}"}}},
		{ToolCalls: []ToolCall{{ID: "2", Name: "echo", Arguments: "{}"}}},
	}}
	registry, _ := NewRegistry(textTool("echo", "x"))
	limits := DefaultLimits()
	limits.MaxSteps = 2
	runner := NewRunner(provider, registry, limits)
	if _, err := runner.Run(context.Background(), RunRequest{SystemPrompt: "s", UserMessage: "u"}); !errors.Is(err, ErrMaxSteps) {
		t.Fatalf("expected ErrMaxSteps, got %v", err)
	}
}

func TestRunFeedsToolErrorsBackToTheModel(t *testing.T) {
	provider := &fakeProvider{responses: []ChatResponse{
		{ToolCalls: []ToolCall{{ID: "1", Name: "boom", Arguments: "{}"}}},
		{Content: "recovered"},
	}}
	registry, _ := NewRegistry(NewTool("boom", "fails", nil, func(ctx context.Context, ec *ExecutionContext, args json.RawMessage) (ToolResult, error) {
		return ToolResult{}, errors.New("database unavailable")
	}))
	runner := NewRunner(provider, registry, DefaultLimits())
	result, err := runner.Run(context.Background(), RunRequest{SystemPrompt: "s", UserMessage: "u"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Output != "recovered" {
		t.Fatalf("unexpected output: %s", result.Output)
	}
	last := provider.lastRequest()
	found := false
	for _, msg := range last.Messages {
		if msg.Role == RoleTool && strings.Contains(msg.Content, "database unavailable") {
			found = true
		}
	}
	if !found {
		t.Fatalf("tool error was not fed back: %+v", last.Messages)
	}
}

func TestRunRejectsDuplicateCalls(t *testing.T) {
	provider := &fakeProvider{responses: []ChatResponse{
		{ToolCalls: []ToolCall{{ID: "1", Name: "echo", Arguments: `{"v":1}`}}},
		{ToolCalls: []ToolCall{{ID: "2", Name: "echo", Arguments: `{"v":1}`}}},
		{ToolCalls: []ToolCall{{ID: "3", Name: "echo", Arguments: `{"v":1}`}}},
	}}
	registry, _ := NewRegistry(textTool("echo", "ok"))
	limits := DefaultLimits()
	limits.MaxSteps = 3
	runner := NewRunner(provider, registry, limits)
	_, _ = runner.Run(context.Background(), RunRequest{SystemPrompt: "s", UserMessage: "u"})

	provider.mu.Lock()
	defer provider.mu.Unlock()
	if len(provider.requests) != 3 {
		t.Fatalf("expected 3 requests, got %d", len(provider.requests))
	}
	found := false
	for _, msg := range provider.requests[2].Messages {
		if msg.Role == RoleTool && strings.Contains(msg.Content, "duplicate_call") {
			found = true
		}
	}
	if !found {
		t.Fatalf("duplicate call was not rejected: %+v", provider.requests[2].Messages)
	}
}

func TestRunEmitsPlanDraft(t *testing.T) {
	provider := &fakeProvider{responses: []ChatResponse{
		{ToolCalls: []ToolCall{{ID: "1", Name: "submit_plan", Arguments: `{"solutionId":"sol-1"}`}}},
		{Content: "saved"},
	}}
	planTool := NewTool("submit_plan", "save", nil, func(ctx context.Context, ec *ExecutionContext, args json.RawMessage) (ToolResult, error) {
		return ToolResult{
			Content: `{"status":"draft"}`,
			Plan: &PlanDraft{
				Version: 1,
				Title:   "Plan",
				Plan:    json.RawMessage(`{"stops":[]}`),
			},
		}, nil
	})
	registry, _ := NewRegistry(planTool)
	emit, records := collectEvents()
	runner := NewRunner(provider, registry, DefaultLimits())
	result, err := runner.Run(context.Background(), RunRequest{SystemPrompt: "s", UserMessage: "u", Emit: emit})
	if err != nil {
		t.Fatal(err)
	}
	if result.Draft == nil || result.Draft.Version != 1 {
		t.Fatalf("expected a draft, got %+v", result.Draft)
	}
	found := false
	for _, rec := range *records {
		if rec.Type == EventPlan {
			found = true
		}
	}
	if !found {
		t.Fatal("plan event was not emitted")
	}
}

func TestRunHonorsCancellation(t *testing.T) {
	provider := &fakeProvider{block: true}
	registry, _ := NewRegistry(textTool("echo", "ok"))
	runner := NewRunner(provider, registry, DefaultLimits())
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, err := runner.Run(ctx, RunRequest{SystemPrompt: "s", UserMessage: "u"})
		done <- err
	}()
	time.Sleep(20 * time.Millisecond)
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context.Canceled, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("run did not stop after cancellation")
	}
}
