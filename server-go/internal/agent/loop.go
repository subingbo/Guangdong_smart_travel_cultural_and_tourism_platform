package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

var ErrMaxSteps = errors.New("agent: max steps reached")

type Runner struct {
	provider  Provider
	registry  *Registry
	limits    Limits
	projector *Projector
}

func NewRunner(provider Provider, registry *Registry, limits Limits) *Runner {
	if limits.MaxSteps <= 0 {
		limits = DefaultLimits()
	}
	return &Runner{
		provider:  provider,
		registry:  registry,
		limits:    limits,
		projector: NewProjector(limits.MaxToolResultBytes),
	}
}

type RunRequest struct {
	RunID          int64
	ConversationID int64
	UserID         int64
	Language       string
	PlanVersion    int
	Model          string
	SystemPrompt   string
	History        []Message
	UserMessage    string
	State          *RunState
	Emit           EmitFunc
	OnStep         func(StepRecord)
}

type RunResult struct {
	Output    string
	Usage     Usage
	StepCount int
	Draft     *PlanDraft
}

// Run executes the ReAct loop: LLM → tool calls → tool results → LLM, until
// the model answers without a tool call or the step budget is exhausted.
func (r *Runner) Run(ctx context.Context, req RunRequest) (*RunResult, error) {
	if r.provider == nil {
		return nil, errors.New("agent: provider is nil")
	}
	if r.registry == nil {
		r.registry = &Registry{byName: map[string]Tool{}}
	}
	state := req.State
	if state == nil {
		state = NewRunState()
	}
	ec := &ExecutionContext{
		RunID:          req.RunID,
		ConversationID: req.ConversationID,
		UserID:         req.UserID,
		Language:       req.Language,
		PlanVersion:    req.PlanVersion,
		State:          state,
	}

	messages := make([]Message, 0, len(req.History)+2)
	messages = append(messages, Message{Role: RoleSystem, Content: req.SystemPrompt})
	messages = append(messages, req.History...)
	messages = append(messages, Message{Role: RoleUser, Content: req.UserMessage})

	result := &RunResult{}
	for step := 1; step <= r.limits.MaxSteps; step++ {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		result.StepCount = step
		stepStart := time.Now()
		resp, err := r.provider.StreamChat(ctx, ChatRequest{
			Model:             req.Model,
			Messages:          messages,
			Tools:             r.registry.Specs(),
			ParallelToolCalls: false,
		}, func(d Delta) error {
			if d.Reasoning != "" {
				req.Emit.Emit(EventThinking, ThinkingPayload{Content: d.Reasoning})
			}
			if d.Content != "" {
				req.Emit.Emit(EventDelta, DeltaPayload{Content: d.Content})
			}
			return nil
		})
		result.Usage.add(resp.Usage)
		r.record(req, StepRecord{
			StepNo:  step,
			Type:    "llm",
			Latency: time.Since(stepStart),
			Result:  truncateRunes(resp.Content, 500),
			Error:   errorText(err),
		})
		if err != nil {
			return result, err
		}

		if len(resp.ToolCalls) == 0 {
			messages = append(messages, Message{Role: RoleAssistant, Content: resp.Content})
			result.Output = resp.Content
			return result, nil
		}

		calls := resp.ToolCalls
		if len(calls) > r.limits.MaxToolCallsPerStep {
			calls = calls[:r.limits.MaxToolCallsPerStep]
		}
		for i := range calls {
			if calls[i].ID == "" {
				calls[i].ID = fmt.Sprintf("call_%d_%d", step, i)
			}
		}
		messages = append(messages, Message{
			Role:      RoleAssistant,
			Content:   resp.Content,
			ToolCalls: calls,
		})

		for _, call := range calls {
			if err := ctx.Err(); err != nil {
				return result, err
			}
			content := r.executeTool(ctx, req, ec, state, result, step, call)
			messages = append(messages, Message{
				Role:       RoleTool,
				ToolCallID: call.ID,
				Name:       call.Name,
				Content:    content,
			})
		}
	}
	return result, ErrMaxSteps
}

func (r *Runner) executeTool(ctx context.Context, req RunRequest, ec *ExecutionContext, state *RunState, result *RunResult, step int, call ToolCall) string {
	req.Emit.Emit(EventToolCall, ToolCallPayload{
		Step:      step,
		Name:      call.Name,
		Arguments: truncateRunes(call.Arguments, 1000),
	})

	fail := func(msg, code string) string {
		req.Emit.Emit(EventToolResult, ToolResultPayload{
			Step:    step,
			Name:    call.Name,
			OK:      false,
			Preview: truncateRunes(msg, 300),
			Error:   code,
		})
		r.record(req, StepRecord{
			StepNo:    step,
			Type:      "tool",
			Name:      call.Name,
			Arguments: truncateRunes(call.Arguments, 2000),
			Result:    truncateRunes(msg, 2000),
			Error:     code,
		})
		return msg
	}

	tool, ok := r.registry.Get(call.Name)
	if !ok {
		return fail(fmt.Sprintf("tool_error: tool %q is not available", call.Name), "unknown_tool")
	}
	if len(call.Arguments) > r.limits.MaxArgumentsBytes {
		return fail("tool_error: arguments are too large; pass only the fields required by the schema", "arguments_too_large")
	}

	canonical := canonicalArgs(call.Arguments)
	if state.MarkSeen("call:" + call.Name + ":" + canonical) {
		return fail("duplicate_call: this exact tool call already ran; reuse the previous result instead of calling it again", "duplicate_call")
	}
	if state.Incr("tool:"+call.Name) > r.limits.Budget(call.Name) {
		return fail(fmt.Sprintf("budget_exhausted: %s has reached its call budget for this run; answer with the data you already have", call.Name), "budget_exhausted")
	}

	start := time.Now()
	toolCtx, cancel := context.WithTimeout(ctx, r.limits.ToolTimeout)
	toolResult, err := safeExecute(toolCtx, tool, ec, []byte(call.Arguments))
	cancel()
	latency := time.Since(start)

	if err != nil {
		return fail("tool_error: "+err.Error(), "tool_error")
	}
	content := toolResult.Content
	if r.projector != nil {
		content = r.projector.Project(call.Name, content)
	}

	if toolResult.Plan != nil {
		result.Draft = toolResult.Plan
		req.Emit.Emit(EventPlan, PlanPayload{
			Version: toolResult.Plan.Version,
			Title:   toolResult.Plan.Title,
			Summary: toolResult.Plan.Summary,
			Plan:    toolResult.Plan.Plan,
		})
	}

	req.Emit.Emit(EventToolResult, ToolResultPayload{
		Step:      step,
		Name:      call.Name,
		OK:        true,
		LatencyMS: int(latency / time.Millisecond),
		Preview:   truncateRunes(content, 300),
	})
	r.record(req, StepRecord{
		StepNo:    step,
		Type:      "tool",
		Name:      call.Name,
		Arguments: truncateRunes(call.Arguments, 2000),
		Result:    truncateRunes(content, 4000),
		Latency:   latency,
	})
	return content
}

func (r *Runner) record(req RunRequest, rec StepRecord) {
	if req.OnStep == nil {
		return
	}
	defer func() { _ = recover() }()
	req.OnStep(rec)
}

func safeExecute(ctx context.Context, tool Tool, ec *ExecutionContext, args json.RawMessage) (result ToolResult, err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("tool panic: %v", p)
		}
	}()
	return tool.Execute(ctx, ec, args)
}

func canonicalArgs(raw string) string {
	if raw == "" {
		return ""
	}
	var value any
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		return raw
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return raw
	}
	return string(encoded)
}

func errorText(err error) string {
	if err == nil {
		return ""
	}
	return truncateRunes(err.Error(), 500)
}

func (u *Usage) add(other Usage) {
	u.PromptTokens += other.PromptTokens
	u.CompletionTokens += other.CompletionTokens
	u.TotalTokens += other.TotalTokens
}
