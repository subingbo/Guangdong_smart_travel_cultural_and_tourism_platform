package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/travelgd/server-go/internal/ai"
	"github.com/travelgd/server-go/internal/common"
)

// AiService backs the merchant-facing AI endpoints used by the mini-program:
// marketing copy, poster generation (async Wanx), menu OCR/translation
// (qwen-vl) and dish culture cards.
type AiService struct {
	qwen   *ai.QwenClient
	vision *ai.VisionClient
	wanx   *ai.WanxClient

	postersMu sync.Mutex
	posters   map[string]*posterEntry
}

// posterEntry pairs the API-visible task with the provider task id so the
// client can poll with only the local id.
type posterEntry struct {
	task       PosterTask
	dashTaskID string
}

func NewAiService(qwen *ai.QwenClient, vision *ai.VisionClient, wanx *ai.WanxClient) *AiService {
	return &AiService{qwen: qwen, vision: vision, wanx: wanx, posters: map[string]*posterEntry{}}
}

// --- marketing copy -------------------------------------------------------

type CopyRequest struct {
	Topic    string `json:"topic"`
	Type     string `json:"type"`   // dish | event | scenery | general
	Tone     string `json:"tone"`   // friendly | literary | promotional
	Language string `json:"language"`
}

type CopyResponse struct {
	Content string `json:"content"`
}

func (s *AiService) Copy(ctx context.Context, req CopyRequest) (*CopyResponse, error) {
	if strings.TrimSpace(req.Topic) == "" {
		return nil, common.BadRequest("topic is required")
	}
	if !s.qwen.Enabled() {
		return nil, common.BadGateway("AI provider not configured (DASHSCOPE_API_KEY missing)")
	}
	lang := normalizeLang(req.Language)
	kind := normalizeChoice(req.Type, []string{"dish", "event", "scenery", "general"}, "general")
	tone := normalizeChoice(req.Tone, []string{"friendly", "literary", "promotional"}, "friendly")

	var sb strings.Builder
	if lang == "zh" {
		sb.WriteString("你是广东文旅的资深文案。请围绕主题写一段商家营销文案，中文输出。")
		switch kind {
		case "dish":
			sb.WriteString("主题是一道招牌菜，突出食材、风味与文化典故。")
		case "event":
			sb.WriteString("主题是一场活动，突出时间地点与参与理由。")
		case "scenery":
			sb.WriteString("主题是一处景点，突出景观与游览体验。")
		default:
			sb.WriteString("主题是一家文旅商户。")
		}
		switch tone {
		case "literary":
			sb.WriteString("文笔偏文艺、有画面感。")
		case "promotional":
			sb.WriteString("语气带促销感，并附一句行动号召。")
		default:
			sb.WriteString("语气亲切自然。")
		}
		sb.WriteString("200 字以内，直接输出正文。")
	} else {
		sb.WriteString("You are a senior copywriter for Guangdong tourism. Write marketing copy in English. ")
		switch kind {
		case "dish":
			sb.WriteString("The topic is a signature dish: highlight ingredients, flavor and culinary heritage. ")
		case "event":
			sb.WriteString("The topic is an event: highlight time, place and why to attend. ")
		case "scenery":
			sb.WriteString("The topic is a scenic spot: highlight the view and the visiting experience. ")
		default:
			sb.WriteString("The topic is a tourism business. ")
		}
		switch tone {
		case "literary":
			sb.WriteString("Use a literary, evocative style. ")
		case "promotional":
			sb.WriteString("Use a promotional tone with one call to action. ")
		default:
			sb.WriteString("Use a warm, friendly tone. ")
		}
		sb.WriteString("Keep it under 120 words and output the copy only.")
	}

	raw, err := s.qwen.Chat(ctx, s.qwen.Model(), []ai.QwenMessage{
		{Role: "system", Content: sb.String()},
		{Role: "user", Content: req.Topic},
	})
	if err != nil {
		return nil, err
	}
	return &CopyResponse{Content: strings.TrimSpace(raw)}, nil
}

// --- poster ---------------------------------------------------------------

type PosterCreateRequest struct {
	Prompt    string `json:"prompt"`
	Style     string `json:"style"`
	Negative  string `json:"negativePrompt"`
	Language  string `json:"language"`
}

type PosterCreateResponse struct {
	TaskID string `json:"taskId"`
}

type PosterTask struct {
	TaskID    string `json:"taskId"`
	Status    string `json:"status"`
	ImageURL  string `json:"imageUrl,omitempty"`
	Message   string `json:"message,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

const posterTaskTTL = 30 * time.Minute

func (s *AiService) CreatePoster(ctx context.Context, req PosterCreateRequest) (*PosterCreateResponse, error) {
	if strings.TrimSpace(req.Prompt) == "" {
		return nil, common.BadRequest("prompt is required")
	}
	prompt := req.Prompt
	if req.Style != "" {
		prompt = prompt + ", " + req.Style
	}
	dashTaskID, err := s.wanx.CreateTask(ctx, prompt, req.Negative)
	if err != nil {
		return nil, err
	}
	taskID := newTaskID()
	s.postersMu.Lock()
	s.gcPostersLocked()
	s.posters[taskID] = &posterEntry{
		task: PosterTask{
			TaskID:    taskID,
			Status:    ai.PosterPending,
			CreatedAt: time.Now(),
		},
		dashTaskID: dashTaskID,
	}
	s.postersMu.Unlock()
	return &PosterCreateResponse{TaskID: taskID}, nil
}

func (s *AiService) PosterTask(ctx context.Context, taskID string) (*PosterTask, error) {
	s.postersMu.Lock()
	entry, ok := s.posters[taskID]
	s.postersMu.Unlock()
	if !ok {
		return nil, common.NotFound("Poster task not found")
	}
	if entry.task.Status == ai.PosterSucceeded || entry.task.Status == ai.PosterFailed {
		result := entry.task
		return &result, nil
	}
	status, url, err := s.wanx.PollTask(ctx, entry.dashTaskID)
	if err != nil {
		s.postersMu.Lock()
		entry.task.Status = ai.PosterFailed
		entry.task.Message = err.Error()
		s.postersMu.Unlock()
		return nil, err
	}
	s.postersMu.Lock()
	entry.task.Status = status
	if url != "" {
		entry.task.ImageURL = url
	}
	result := entry.task
	s.postersMu.Unlock()
	return &result, nil
}

func (s *AiService) gcPostersLocked() {
	now := time.Now()
	for id, e := range s.posters {
		if now.Sub(e.task.CreatedAt) > posterTaskTTL {
			delete(s.posters, id)
		}
	}
}

func newTaskID() string {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return time.Now().Format("20060102150405")
	}
	return "p" + hex.EncodeToString(b)
}

// --- menu OCR / translate -------------------------------------------------

type MenuImageRequest struct {
	Image    string `json:"image"` // base64 or data URL
	Language string `json:"language"`
}

type MenuOCRResponse struct {
	Text string `json:"text"`
}

func (s *AiService) MenuOCR(ctx context.Context, req MenuImageRequest) (*MenuOCRResponse, error) {
	if !s.vision.Enabled() {
		return nil, common.BadGateway("AI provider not configured (DASHSCOPE_API_KEY missing)")
	}
	dataURL, err := ai.NormalizeImagePayload(req.Image)
	if err != nil {
		return nil, err
	}
	lang := normalizeLang(req.Language)
	system := "You are an OCR engine for restaurant menus. Output only the recognized text, " +
		"one dish per line in the format 'name | price' when a price exists. No commentary."
	if lang == "zh" {
		system = "你是菜单 OCR 引擎。只输出识别到的文字，每行一道菜，格式为'菜名 | 价格'（有价格时）。不要任何解释。"
	}
	raw, err := s.vision.VisionChat(ctx, system, "Recognize all menu text in this image.", dataURL)
	if err != nil {
		return nil, err
	}
	return &MenuOCRResponse{Text: strings.TrimSpace(raw)}, nil
}

type MenuTranslateResponse struct {
	Items []TranslatedDish `json:"items"`
}

type TranslatedDish struct {
	Name        string `json:"name"`
	Translation string `json:"translation"`
	Notes       string `json:"notes,omitempty"`
}

func (s *AiService) MenuTranslate(ctx context.Context, req MenuImageRequest) (*MenuTranslateResponse, error) {
	if !s.vision.Enabled() {
		return nil, common.BadGateway("AI provider not configured (DASHSCOPE_API_KEY missing)")
	}
	dataURL, err := ai.NormalizeImagePayload(req.Image)
	if err != nil {
		return nil, err
	}
	target := targetLanguageName(req.Language)
	system := fmt.Sprintf("You translate restaurant menus for foreign visitors. Read the menu image and return a "+
		"VALID JSON array of objects: [{\"name\": original dish name, \"translation\": %s name, "+
		"\"notes\": one short sentence in %s about culture/dietary tips or empty}]. No markdown fences.",
		target, target)
	raw, err := s.vision.VisionChat(ctx, system, "Translate this menu.", dataURL)
	if err != nil {
		return nil, err
	}
	var items []TranslatedDish
	if err := json.Unmarshal([]byte(ai.StripFences(raw)), &items); err != nil || len(items) == 0 {
		return nil, common.BadGateway("Could not parse AI translation response")
	}
	return &MenuTranslateResponse{Items: items}, nil
}

// targetLanguageName maps the client language code to a name the vision model
// understands; unknown codes fall back to English.
func targetLanguageName(code string) string {
	switch strings.ToLower(code) {
	case "zh", "zh-cn", "zh_cn":
		return "Chinese"
	case "ja":
		return "Japanese"
	case "ko":
		return "Korean"
	case "es":
		return "Spanish"
	case "fr":
		return "French"
	default:
		return "English"
	}
}

// --- dish explain / culture card ------------------------------------------

type DishExplainRequest struct {
	Name     string `json:"name"`
	Kind     string `json:"kind"` // card | explain
	Language string `json:"language"`
}

type DishExplainResponse struct {
	Title   string   `json:"title"`
	Content string   `json:"content"`
	Tags    []string `json:"tags,omitempty"`
}

func (s *AiService) DishExplain(ctx context.Context, req DishExplainRequest) (*DishExplainResponse, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, common.BadRequest("name is required")
	}
	if !s.qwen.Enabled() {
		return nil, common.BadGateway("AI provider not configured (DASHSCOPE_API_KEY missing)")
	}
	lang := normalizeLang(req.Language)
	kind := normalizeChoice(req.Kind, []string{"card", "explain"}, "card")

	var system string
	if kind == "card" {
		if lang == "zh" {
			system = "你是粤菜文化讲解员。为菜品生成文化卡片，返回合法 JSON：" +
				"{\"title\":\"菜品名\",\"content\":\"80-150字的文化典故与风味介绍\",\"tags\":[\"标签1\",\"标签2\"]}。" +
				"不要输出 markdown 代码块。"
		} else {
			system = "You are a Cantonese cuisine docent. Return a VALID JSON object " +
				"{\"title\": dish name, \"content\": 60-100 words of culinary history and flavor notes, " +
				"\"tags\": [\"tag1\",\"tag2\"]}. No markdown fences."
		}
	} else if lang == "zh" {
		system = "你是粤菜服务员，用两三句通俗中文介绍这道菜的味道、做法与吃法建议，直接输出正文。"
	} else {
		system = "You are a Cantonese waiter. Explain the dish's flavor, preparation and how to enjoy it " +
			"in 2-3 short sentences. Output the text only."
	}

	raw, err := s.qwen.Chat(ctx, s.qwen.Model(), []ai.QwenMessage{
		{Role: "system", Content: system},
		{Role: "user", Content: req.Name},
	})
	if err != nil {
		return nil, err
	}
	raw = strings.TrimSpace(raw)
	if kind == "card" {
		var out DishExplainResponse
		if err := json.Unmarshal([]byte(ai.StripFences(raw)), &out); err != nil || out.Title == "" {
			// Fall back to a plain card instead of failing the demo flow.
			return &DishExplainResponse{Title: req.Name, Content: raw}, nil
		}
		return &out, nil
	}
	return &DishExplainResponse{Title: req.Name, Content: raw}, nil
}

func normalizeLang(lang string) string {
	if strings.HasPrefix(strings.ToLower(lang), "zh") {
		return "zh"
	}
	return "en"
}
