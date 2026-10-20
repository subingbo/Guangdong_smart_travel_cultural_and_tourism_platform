// Package ai holds the external providers: DashScope (Qwen) and Tencent Maps.
// Both mirror server-java's service/ai package one-to-one.
package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/travelgd/server-go/internal/common"
)

type QwenClient struct {
	apiKey  string
	chatURL string
	client  *http.Client
}

func NewQwenClient(apiKey, chatURL string) *QwenClient {
	return &QwenClient{
		apiKey:  apiKey,
		chatURL: chatURL,
		client:  &http.Client{Timeout: 60 * time.Second},
	}
}

func (q *QwenClient) Enabled() bool { return q.apiKey != "" }

type QwenMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

/** Sends a chat completion request and returns the assistant message text. */
func (q *QwenClient) Chat(ctx context.Context, model string, messages []QwenMessage) (string, error) {
	if !q.Enabled() {
		return "", common.BadGateway("AI provider not configured (DASHSCOPE_API_KEY missing)")
	}
	body, _ := json.Marshal(map[string]interface{}{
		"model":      model,
		"input":      map[string]interface{}{"messages": messages},
		"parameters": map[string]interface{}{"result_format": "message"},
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, q.chatURL, bytes.NewReader(body))
	if err != nil {
		return "", common.BadGateway("AI request build failed: " + err.Error())
	}
	req.Header.Set("Authorization", "Bearer "+q.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := q.client.Do(req)
	if err != nil {
		return "", common.BadGateway("AI provider unreachable: " + err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", common.BadGateway(fmt.Sprintf("AI provider error: HTTP %d", resp.StatusCode))
	}
	var out struct {
		Output struct {
			Choices []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
		} `json:"output"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil || len(out.Output.Choices) == 0 {
		return "", common.BadGateway("Unexpected AI response shape")
	}
	return out.Output.Choices[0].Message.Content, nil
}

/** Strips ```json fences the model sometimes adds. */
func StripFences(text string) string {
	t := strings.ReplaceAll(text, "```json", "")
	return strings.TrimSpace(strings.ReplaceAll(t, "```", ""))
}
