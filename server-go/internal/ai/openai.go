package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/travelgd/server-go/internal/agent"
	"github.com/travelgd/server-go/internal/common"
)

// OpenAIClient speaks the OpenAI-compatible chat-completions protocol used by
// DashScope. It supports tools, streaming deltas, reasoning content and usage.
// The legacy QwenClient stays untouched for /guide/chat and /routes/generate.
type OpenAIClient struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
	stream  bool
}

func NewOpenAIClient(apiKey, baseURL, model string, stream bool) *OpenAIClient {
	if baseURL == "" {
		baseURL = "https://dashscope.aliyuncs.com/compatible-mode/v1"
	}
	baseURL = strings.TrimRight(baseURL, "/")
	if model == "" {
		model = "qwen-plus"
	}
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 90 * time.Second,
		MaxIdleConns:          20,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
	}
	return &OpenAIClient{
		apiKey:  apiKey,
		baseURL: baseURL,
		model:   model,
		client:  &http.Client{Transport: transport},
		stream:  stream,
	}
}

func (c *OpenAIClient) Enabled() bool { return c != nil && c.apiKey != "" }
func (c *OpenAIClient) Model() string { return c.model }

func (c *OpenAIClient) StreamChat(ctx context.Context, req agent.ChatRequest, onDelta func(agent.Delta) error) (agent.ChatResponse, error) {
	if !c.Enabled() {
		return agent.ChatResponse{}, common.BadGateway("AI provider not configured (DASHSCOPE_API_KEY missing)")
	}
	model := req.Model
	if model == "" {
		model = c.model
	}
	wire := oaRequest{
		Model:             model,
		Messages:          toOAMessages(req.Messages),
		Tools:             toOATools(req.Tools),
		ParallelToolCalls: req.ParallelToolCalls,
		Stream:            c.stream,
	}
	if len(wire.Tools) > 0 {
		wire.ToolChoice = "auto"
	}
	if c.stream {
		wire.StreamOptions = &oaStreamOptions{IncludeUsage: true}
	}
	if req.Temperature > 0 {
		t := req.Temperature
		wire.Temperature = &t
	}

	if c.stream {
		return c.streamOnce(ctx, wire, onDelta)
	}
	return c.completeOnce(ctx, wire, onDelta)
}

func (c *OpenAIClient) streamOnce(ctx context.Context, wire oaRequest, onDelta func(agent.Delta) error) (agent.ChatResponse, error) {
	resp, err := c.do(ctx, wire)
	if err != nil {
		return agent.ChatResponse{}, err
	}
	defer resp.Body.Close()

	var (
		content      strings.Builder
		reasoning    strings.Builder
		accs         = map[int]*toolAccumulator{}
		finishReason string
		usage        agent.Usage
		sawChunk     bool
	)
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return agent.ChatResponse{}, err
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payload == "[DONE]" {
			break
		}
		var chunk oaChunk
		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			continue
		}
		sawChunk = true
		if chunk.Usage != nil {
			usage = chunk.Usage.toAgent()
		}
		for _, choice := range chunk.Choices {
			if choice.FinishReason != "" {
				finishReason = choice.FinishReason
			}
			if choice.Delta.ReasoningContent != "" {
				reasoning.WriteString(choice.Delta.ReasoningContent)
				if onDelta != nil {
					if err := onDelta(agent.Delta{Reasoning: choice.Delta.ReasoningContent}); err != nil {
						return agent.ChatResponse{}, err
					}
				}
			}
			if choice.Delta.Content != "" {
				content.WriteString(choice.Delta.Content)
				if onDelta != nil {
					if err := onDelta(agent.Delta{Content: choice.Delta.Content}); err != nil {
						return agent.ChatResponse{}, err
					}
				}
			}
			for _, tc := range choice.Delta.ToolCalls {
				acc := accs[tc.Index]
				if acc == nil {
					acc = &toolAccumulator{}
					accs[tc.Index] = acc
				}
				if tc.ID != "" {
					acc.ID = tc.ID
				}
				if tc.Function.Name != "" && acc.Name == "" {
					acc.Name = tc.Function.Name
				}
				if tc.Function.Arguments != "" {
					acc.Arguments += tc.Function.Arguments
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		// A canceled run or a deadline must stay recognizable so the run
		// service can mark it canceled/timed out instead of llm_error.
		if ctxErr := ctx.Err(); ctxErr != nil {
			return agent.ChatResponse{}, ctxErr
		}
		return agent.ChatResponse{}, common.BadGateway("AI stream interrupted: " + err.Error())
	}
	if !sawChunk {
		return agent.ChatResponse{}, common.BadGateway("AI provider returned an empty stream")
	}

	toolCalls := collectToolCalls(accs)
	if usage.TotalTokens == 0 {
		usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
	}
	return agent.ChatResponse{
		Content:      content.String(),
		ToolCalls:    toolCalls,
		FinishReason: finishReason,
		Usage:        usage,
	}, nil
}

func (c *OpenAIClient) completeOnce(ctx context.Context, wire oaRequest, onDelta func(agent.Delta) error) (agent.ChatResponse, error) {
	wire.Stream = false
	resp, err := c.do(ctx, wire)
	if err != nil {
		return agent.ChatResponse{}, err
	}
	defer resp.Body.Close()

	var out oaResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 8*1024*1024)).Decode(&out); err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return agent.ChatResponse{}, ctxErr
		}
		return agent.ChatResponse{}, common.BadGateway("Unexpected AI response shape")
	}
	if len(out.Choices) == 0 {
		return agent.ChatResponse{}, common.BadGateway("Unexpected AI response shape")
	}
	choice := out.Choices[0]
	if choice.Message.Content != "" && onDelta != nil {
		if err := onDelta(agent.Delta{Content: choice.Message.Content}); err != nil {
			return agent.ChatResponse{}, err
		}
	}
	toolCalls := make([]agent.ToolCall, 0, len(choice.Message.ToolCalls))
	for _, tc := range choice.Message.ToolCalls {
		toolCalls = append(toolCalls, agent.ToolCall{
			ID:        tc.ID,
			Name:      tc.Function.Name,
			Arguments: tc.Function.Arguments,
		})
	}
	return agent.ChatResponse{
		Content:      choice.Message.Content,
		ToolCalls:    toolCalls,
		FinishReason: choice.FinishReason,
		Usage:        out.Usage.toAgent(),
	}, nil
}

func (c *OpenAIClient) do(ctx context.Context, wire oaRequest) (*http.Response, error) {
	body, err := json.Marshal(wire)
	if err != nil {
		return nil, common.BadGateway("AI request build failed")
	}
	endpoint := c.baseURL + "/chat/completions"
	if strings.HasSuffix(c.baseURL, "/chat/completions") {
		endpoint = c.baseURL
	}

	for attempt := 0; attempt < 2; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
		if err != nil {
			return nil, common.BadGateway("AI request build failed")
		}
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("Content-Type", "application/json")
		if c.stream {
			req.Header.Set("Accept", "text/event-stream")
		}

		resp, err := c.client.Do(req)
		if err != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			return nil, common.BadGateway("AI provider unreachable: " + err.Error())
		}
		if resp.StatusCode == http.StatusOK {
			return resp, nil
		}

		detail, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		resp.Body.Close()
		log.Printf("ai provider error: HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(detail)))

		if resp.StatusCode == http.StatusTooManyRequests {
			if attempt == 0 {
				select {
				case <-time.After(500 * time.Millisecond):
					continue
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			}
			return nil, common.TooManyRequests("AI 配额或频率受限，请稍后重试")
		}
		if resp.StatusCode >= 500 && attempt == 0 {
			select {
			case <-time.After(500 * time.Millisecond):
				continue
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
		return nil, common.BadGateway(fmt.Sprintf("AI provider error: HTTP %d", resp.StatusCode))
	}
	return nil, common.BadGateway("AI provider error")
}

// --- wire types ---

type oaRequest struct {
	Model             string           `json:"model"`
	Messages          []oaMessage      `json:"messages"`
	Tools             []oaTool         `json:"tools,omitempty"`
	ToolChoice        string           `json:"tool_choice,omitempty"`
	ParallelToolCalls bool             `json:"parallel_tool_calls"`
	Stream            bool             `json:"stream"`
	StreamOptions     *oaStreamOptions `json:"stream_options,omitempty"`
	Temperature       *float64         `json:"temperature,omitempty"`
}

type oaStreamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

type oaMessage struct {
	Role       string       `json:"role"`
	Content    string       `json:"content,omitempty"`
	ToolCalls  []oaToolCall `json:"tool_calls,omitempty"`
	ToolCallID string       `json:"tool_call_id,omitempty"`
}

type oaTool struct {
	Type     string         `json:"type"`
	Function oaToolFunction `json:"function"`
}

type oaToolFunction struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

type oaCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type oaToolCall struct {
	ID       string         `json:"id"`
	Type     string         `json:"type,omitempty"`
	Function oaCallFunction `json:"function"`
}

type oaUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

func (u oaUsage) toAgent() agent.Usage {
	total := u.TotalTokens
	if total == 0 {
		total = u.PromptTokens + u.CompletionTokens
	}
	return agent.Usage{PromptTokens: u.PromptTokens, CompletionTokens: u.CompletionTokens, TotalTokens: total}
}

type oaChunk struct {
	Choices []struct {
		Delta struct {
			Content          string `json:"content"`
			ReasoningContent string `json:"reasoning_content"`
			ToolCalls        []struct {
				Index    int    `json:"index"`
				ID       string `json:"id"`
				Function struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				} `json:"function"`
			} `json:"tool_calls"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage *oaUsage `json:"usage"`
}

type oaResponse struct {
	Choices []struct {
		Message struct {
			Content          string       `json:"content"`
			ReasoningContent string       `json:"reasoning_content"`
			ToolCalls        []oaToolCall `json:"tool_calls"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage oaUsage `json:"usage"`
}

type toolAccumulator struct {
	ID        string
	Name      string
	Arguments string
}

func collectToolCalls(accs map[int]*toolAccumulator) []agent.ToolCall {
	if len(accs) == 0 {
		return nil
	}
	max := -1
	for idx := range accs {
		if idx > max {
			max = idx
		}
	}
	out := make([]agent.ToolCall, 0, len(accs))
	for i := 0; i <= max; i++ {
		acc, ok := accs[i]
		if !ok || acc.Name == "" {
			continue
		}
		out = append(out, agent.ToolCall{ID: acc.ID, Name: acc.Name, Arguments: acc.Arguments})
	}
	return out
}

func toOAMessages(messages []agent.Message) []oaMessage {
	out := make([]oaMessage, 0, len(messages))
	for _, m := range messages {
		msg := oaMessage{
			Role:       string(m.Role),
			Content:    m.Content,
			ToolCallID: m.ToolCallID,
		}
		for _, tc := range m.ToolCalls {
			msg.ToolCalls = append(msg.ToolCalls, oaToolCall{
				ID:   tc.ID,
				Type: "function",
				Function: oaCallFunction{
					Name:      tc.Name,
					Arguments: tc.Arguments,
				},
			})
		}
		out = append(out, msg)
	}
	return out
}

func toOATools(specs []agent.ToolSpec) []oaTool {
	if len(specs) == 0 {
		return nil
	}
	out := make([]oaTool, 0, len(specs))
	for _, spec := range specs {
		out = append(out, oaTool{
			Type: "function",
			Function: oaToolFunction{
				Name:        spec.Name,
				Description: spec.Description,
				Parameters:  spec.Parameters,
			},
		})
	}
	return out
}
