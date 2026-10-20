package service

import (
	"context"
	"encoding/json"

	"github.com/travelgd/server-go/internal/ai"
	"github.com/travelgd/server-go/internal/common"
	"github.com/travelgd/server-go/internal/model"
	"github.com/travelgd/server-go/internal/repo"
)

type BlogService struct {
	blogs *repo.BlogRepo
	qwen  *ai.QwenClient
}

func NewBlogService(blogs *repo.BlogRepo, qwen *ai.QwenClient) *BlogService {
	return &BlogService{blogs: blogs, qwen: qwen}
}

type BlogCreateRequest struct {
	Title   string   `json:"title"`
	Content string   `json:"content"`
	Images  []string `json:"images"`
}

/** Public feed; likedByMe is only true for the authenticated viewer's likes. */
func (s *BlogService) ListFeed(ctx context.Context, viewerID int64) ([]model.TravelBlog, error) {
	return s.blogs.FindAll(ctx, viewerID)
}

func (s *BlogService) GetBlog(ctx context.Context, id, viewerID int64) (*model.TravelBlog, error) {
	b, err := s.blogs.FindByID(ctx, id, viewerID)
	if err != nil {
		return nil, err
	}
	if b == nil {
		return nil, common.NotFound("Blog not found")
	}
	return b, nil
}

func (s *BlogService) Publish(ctx context.Context, userID int64, req BlogCreateRequest) (*model.TravelBlog, error) {
	blog := &model.TravelBlog{
		UserID:  userID,
		Title:   req.Title,
		Content: req.Content,
		Images:  req.Images,
	}
	if err := s.blogs.Insert(ctx, blog); err != nil {
		return nil, err
	}
	return s.GetBlog(ctx, blog.ID, userID)
}

func (s *BlogService) Like(ctx context.Context, id int64) (int, error) {
	if _, err := s.GetBlog(ctx, id, 0); err != nil {
		return 0, err
	}
	if err := s.blogs.IncrementLikes(ctx, id); err != nil {
		return 0, err
	}
	b, err := s.GetBlog(ctx, id, 0)
	if err != nil {
		return 0, err
	}
	return b.Likes, nil
}

type BlogDraft struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

/** AI-drafts a blog post from a topic or rough notes. */
func (s *BlogService) Draft(ctx context.Context, topic string) (*BlogDraft, error) {
	systemPrompt := `You are a travel blogger writing about Guangzhou and Lingnan culture. Based on the user's topic or notes, produce a short blog post (200-300 words) in English with a friendly, vivid tone. Return ONLY valid JSON: { "title": string, "content": string }. Do NOT output markdown.`
	raw, err := s.qwen.Chat(ctx, "qwen-plus", []ai.QwenMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: topic},
	})
	if err != nil {
		return nil, err
	}
	var draft BlogDraft
	if err := json.Unmarshal([]byte(ai.StripFences(raw)), &draft); err != nil {
		return &BlogDraft{Title: topic, Content: raw}, nil
	}
	return &draft, nil
}
