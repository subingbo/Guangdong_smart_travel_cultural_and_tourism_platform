package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/travelgd/server-go/internal/common"
)

// VisionClient calls the DashScope OpenAI-compatible endpoint with a
// multimodal (text + image) user message. Used for menu OCR and translation,
// where the image is handed over as a base64 data URL.
type VisionClient struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

func NewVisionClient(apiKey, baseURL, model string) *VisionClient {
	if baseURL == "" {
		baseURL = "https://dashscope.aliyuncs.com/compatible-mode/v1"
	}
	if model == "" {
		// qwen-vl-plus is not enabled on every DashScope account; the 3.x
		// generation is the multimodal default that works with compatible-mode.
		model = "qwen3-vl-plus"
	}
	return &VisionClient{
		apiKey:  apiKey,
		baseURL: strings.TrimRight(baseURL, "/"),
		model:   model,
		client:  &http.Client{Timeout: 90 * time.Second},
	}
}

func (v *VisionClient) Enabled() bool { return v != nil && v.apiKey != "" }

type vlContentPart struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	ImageURL *struct {
		URL string `json:"url"`
	} `json:"image_url,omitempty"`
}

type vlMessage struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

type vlRequest struct {
	Model          string      `json:"model"`
	Messages       []vlMessage `json:"messages"`
	Temperature    float64     `json:"temperature"`
	MaxTokens      int         `json:"max_tokens,omitempty"`
	ResponseFormat *struct {
		Type string `json:"type"`
	} `json:"response_format,omitempty"`
}

type vlResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// NormalizeImagePayload accepts a base64 data URL or a bare base64 string and
// returns a data URL the vision endpoint accepts.
func NormalizeImagePayload(image string) (string, error) {
	image = strings.TrimSpace(image)
	if image == "" {
		return "", common.BadRequest("image is required")
	}
	if strings.HasPrefix(image, "data:image/") {
		return image, nil
	}
	if strings.HasPrefix(image, "http://") || strings.HasPrefix(image, "https://") {
		return image, nil
	}
	return "data:image/jpeg;base64," + image, nil
}

// VisionChat sends one system prompt plus a multimodal user turn and returns
// the assistant text. `question` may be empty for pure describe tasks.
func (v *VisionClient) VisionChat(ctx context.Context, system, question, imageDataURL string) (string, error) {
	if !v.Enabled() {
		return "", common.BadGateway("AI provider not configured (DASHSCOPE_API_KEY missing)")
	}
	parts := []vlContentPart{{Type: "text", Text: question}}
	if imageDataURL != "" {
		part := vlContentPart{Type: "image_url"}
		part.ImageURL = &struct {
			URL string `json:"url"`
		}{URL: imageDataURL}
		parts = append(parts, part)
	}
	userContent, err := json.Marshal(parts)
	if err != nil {
		return "", common.BadGateway("AI request build failed: " + err.Error())
	}
	messages := []vlMessage{}
	if system != "" {
		sysContent, _ := json.Marshal(system)
		messages = append(messages, vlMessage{Role: "system", Content: sysContent})
	}
	messages = append(messages, vlMessage{Role: "user", Content: userContent})

	wire := vlRequest{Model: v.model, Messages: messages, Temperature: 0.3}
	body, _ := json.Marshal(wire)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, v.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", common.BadGateway("AI request build failed: " + err.Error())
	}
	req.Header.Set("Authorization", "Bearer "+v.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := v.client.Do(req)
	if err != nil {
		return "", common.BadGateway("AI provider unreachable: " + err.Error())
	}
	defer resp.Body.Close()
	detail, _ := io.ReadAll(io.LimitReader(resp.Body, 8*1024*1024))
	if resp.StatusCode != http.StatusOK {
		return "", common.BadGateway(fmt.Sprintf("AI provider error: HTTP %d: %s",
			resp.StatusCode, strings.TrimSpace(string(detail))))
	}
	var out vlResponse
	if err := json.Unmarshal(detail, &out); err != nil || len(out.Choices) == 0 {
		return "", common.BadGateway("Unexpected AI response shape")
	}
	return out.Choices[0].Message.Content, nil
}
