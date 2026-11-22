package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/travelgd/server-go/internal/ai"
	"github.com/travelgd/server-go/internal/common"
	"github.com/travelgd/server-go/internal/model"
	"github.com/travelgd/server-go/internal/repo"
)

// GuideService is the legacy /guide/chat path: a keyword knowledge search whose
// hits are appended to the system prompt.
type GuideService struct {
	sessions  *repo.GuideRepo
	knowledge *repo.KnowledgeRepo
	qwen      *ai.QwenClient
}

func NewGuideService(sessions *repo.GuideRepo, knowledge *repo.KnowledgeRepo, qwen *ai.QwenClient) *GuideService {
	return &GuideService{sessions: sessions, knowledge: knowledge, qwen: qwen}
}

type GuideChatRequest struct {
	SessionID int64  `json:"sessionId"`
	Question  string `json:"question"`
}

type GuideChatResponse struct {
	SessionID int64  `json:"sessionId"`
	Reply     string `json:"reply"`
}

func (s *GuideService) Chat(ctx context.Context, userID int64, req GuideChatRequest) (*GuideChatResponse, error) {
	var session *model.GuideSession
	var err error
	if req.SessionID == 0 {
		session = &model.GuideSession{UserID: userID, Title: truncateRunes(req.Question, 60)}
		if err := s.sessions.InsertSession(ctx, session); err != nil {
			return nil, err
		}
	} else {
		session, err = s.sessions.FindSession(ctx, req.SessionID, userID)
		if err != nil {
			return nil, err
		}
		if session == nil {
			return nil, common.NotFound("Session not found")
		}
	}

	userMsg := &model.GuideMessage{SessionID: session.ID, Role: "user", Content: req.Question}
	if err := s.sessions.InsertMessage(ctx, userMsg); err != nil {
		return nil, err
	}

	systemPrompt := s.buildSystemPrompt(ctx, req.Question)
	history, err := s.sessions.FindRecentMessages(ctx, session.ID, 20)
	if err != nil {
		return nil, err
	}
	messages := []ai.QwenMessage{{Role: "system", Content: systemPrompt}}
	for _, m := range history {
		messages = append(messages, ai.QwenMessage{Role: m.Role, Content: m.Content})
	}

	reply, err := s.qwen.ChatDefault(ctx, messages)
	if err != nil {
		return nil, err
	}
	botMsg := &model.GuideMessage{SessionID: session.ID, Role: "assistant", Content: reply}
	if err := s.sessions.InsertMessage(ctx, botMsg); err != nil {
		return nil, err
	}
	if err := s.sessions.TouchSession(ctx, session.ID); err != nil {
		return nil, err
	}
	return &GuideChatResponse{SessionID: session.ID, Reply: reply}, nil
}

func (s *GuideService) buildSystemPrompt(ctx context.Context, question string) string {
	sb := strings.Builder{}
	sb.WriteString("You are a knowledgeable AI tour guide for Guangdong / Lingnan culture. " +
		"Answer the user's question about travel, history, or culture. " +
		"Keep answers concise (under 100 words) and engaging.")
	keyword := truncateRunes(question, 60)
	hits, err := s.knowledge.SearchByKeyword(ctx, keyword, 3)
	if err == nil && len(hits) > 0 {
		sb.WriteString(" Relevant cultural background:")
		for _, k := range hits {
			fmt.Fprintf(&sb, " [%s: %s]", k.TitleEn, k.ContentEn)
		}
	}
	return sb.String()
}

func truncateRunes(s string, n int) string {
	if n <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}

func (s *GuideService) GetSession(ctx context.Context, userID, id int64) (*model.GuideSession, error) {
	session, err := s.sessions.FindSession(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, common.NotFound("Session not found")
	}
	messages, err := s.sessions.FindMessages(ctx, id)
	if err != nil {
		return nil, err
	}
	session.Messages = messages
	return session, nil
}
