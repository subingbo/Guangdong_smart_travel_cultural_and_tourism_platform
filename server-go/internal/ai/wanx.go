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

// WanxClient drives the DashScope text-to-image async API (Wanx models).
// CreateTask returns the provider task id; PollTask maps the provider status
// into the small PosterStatus vocabulary the client polls.
type WanxClient struct {
	apiKey     string
	createURL  string
	taskURL    string
	model      string
	httpClient *http.Client
}

const (
	PosterPending   = "pending"
	PosterRunning   = "running"
	PosterSucceeded = "succeeded"
	PosterFailed    = "failed"
)

func NewWanxClient(apiKey string) *WanxClient {
	return &WanxClient{
		apiKey:     apiKey,
		createURL:  "https://dashscope.aliyuncs.com/api/v1/services/aigc/text2image/image-synthesis",
		taskURL:    "https://dashscope.aliyuncs.com/api/v1/tasks/",
		model:      "wanx2.1-t2i-turbo",
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (w *WanxClient) Enabled() bool { return w != nil && w.apiKey != "" }

type wanxCreateResponse struct {
	Output struct {
		TaskID     string `json:"task_id"`
		TaskStatus string `json:"task_status"`
	} `json:"output"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type wanxTaskResponse struct {
	Output struct {
		TaskID     string `json:"task_id"`
		TaskStatus string `json:"task_status"`
		Results    []struct {
			URL string `json:"url"`
		} `json:"results"`
		Message string `json:"message"`
		Code    string `json:"code"`
	} `json:"output"`
}

// CreateTask submits one 1024x1024 image generation and returns the provider
// task id used by PollTask.
func (w *WanxClient) CreateTask(ctx context.Context, prompt, negativePrompt string) (string, error) {
	if !w.Enabled() {
		return "", common.BadGateway("AI provider not configured (DASHSCOPE_API_KEY missing)")
	}
	input := map[string]interface{}{"prompt": prompt}
	if negativePrompt != "" {
		input["negative_prompt"] = negativePrompt
	}
	body, _ := json.Marshal(map[string]interface{}{
		"model":      w.model,
		"input":      input,
		"parameters": map[string]interface{}{"size": "1024*1024", "n": 1},
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.createURL, bytes.NewReader(body))
	if err != nil {
		return "", common.BadGateway("AI request build failed: " + err.Error())
	}
	req.Header.Set("Authorization", "Bearer "+w.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-DashScope-Async", "enable")

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return "", common.BadGateway("AI provider unreachable: " + err.Error())
	}
	defer resp.Body.Close()
	detail, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if resp.StatusCode != http.StatusOK {
		return "", common.BadGateway(fmt.Sprintf("AI provider error: HTTP %d: %s",
			resp.StatusCode, strings.TrimSpace(string(detail))))
	}
	var out wanxCreateResponse
	if err := json.Unmarshal(detail, &out); err != nil || out.Output.TaskID == "" {
		return "", common.BadGateway("Unexpected AI response shape")
	}
	return out.Output.TaskID, nil
}

// PollTask returns the normalized status plus the image URL once SUCCEEDED.
func (w *WanxClient) PollTask(ctx context.Context, taskID string) (string, string, error) {
	if !w.Enabled() {
		return PosterFailed, "", common.BadGateway("AI provider not configured (DASHSCOPE_API_KEY missing)")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, w.taskURL+taskID, nil)
	if err != nil {
		return PosterFailed, "", common.BadGateway("AI request build failed: " + err.Error())
	}
	req.Header.Set("Authorization", "Bearer "+w.apiKey)

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return PosterFailed, "", common.BadGateway("AI provider unreachable: " + err.Error())
	}
	defer resp.Body.Close()
	detail, _ := io.ReadAll(io.LimitReader(resp.Body, 256*1024))
	if resp.StatusCode != http.StatusOK {
		return PosterFailed, "", common.BadGateway(fmt.Sprintf("AI provider error: HTTP %d", resp.StatusCode))
	}
	var out wanxTaskResponse
	if err := json.Unmarshal(detail, &out); err != nil {
		return PosterFailed, "", common.BadGateway("Unexpected AI response shape")
	}
	url := ""
	if len(out.Output.Results) > 0 {
		url = out.Output.Results[0].URL
	}
	switch out.Output.TaskStatus {
	case "SUCCEEDED":
		if url == "" {
			return PosterFailed, "", common.BadGateway("AI task finished without an image")
		}
		return PosterSucceeded, url, nil
	case "FAILED", "CANCELED", "UNKNOWN":
		reason := out.Output.Message
		if reason == "" {
			reason = out.Output.Code
		}
		if reason == "" {
			reason = out.Output.TaskStatus
		}
		return PosterFailed, "", common.BadGateway("AI poster task failed: " + reason)
	case "RUNNING":
		return PosterRunning, "", nil
	default: // PENDING and anything else
		return PosterPending, "", nil
	}
}
