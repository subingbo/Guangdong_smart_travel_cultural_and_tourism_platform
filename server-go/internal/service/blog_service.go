package service

import (
	"context"
	"encoding/json"
	"strings"
	"unicode/utf8"

	"github.com/travelgd/server-go/internal/ai"
	"github.com/travelgd/server-go/internal/common"
	"github.com/travelgd/server-go/internal/model"
	"github.com/travelgd/server-go/internal/repo"
)

const (
	maxTitleRunes    = 200
	maxContentRunes  = 20000
	maxImagesPerBlog = 9
	maxImageURLLen   = 500
)

type BlogService struct {
	blogs         *repo.BlogRepo
	qwen          *ai.QwenClient
	notifications *NotificationService
}

func NewBlogService(blogs *repo.BlogRepo, qwen *ai.QwenClient, notifications *NotificationService) *BlogService {
	return &BlogService{blogs: blogs, qwen: qwen, notifications: notifications}
}

type BlogCreateRequest struct {
	Title   string   `json:"title"`
	Content string   `json:"content"`
	Images  []string `json:"images"`
}

type LikeResult struct {
	Likes     int  `json:"likes"`
	LikedByMe bool `json:"likedByMe"`
}

// Feed returns a page of public blogs; authorID and query are optional filters.
func (s *BlogService) Feed(ctx context.Context, viewerID, authorID int64, query string, page, pageSize int) (*BlogPage, error) {
	page, pageSize, offset := normalizePage(page, pageSize)
	items, total, err := s.blogs.FindPage(ctx, viewerID, authorID, query, offset, pageSize)
	if err != nil {
		return nil, err
	}
	for i := range items {
		attachBlogAuthor(&items[i])
	}
	return &BlogPage{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *BlogService) GetBlog(ctx context.Context, id, viewerID int64) (*model.TravelBlog, error) {
	b, err := s.blogs.FindByID(ctx, id, viewerID)
	if err != nil {
		return nil, err
	}
	if b == nil {
		return nil, common.NotFound("Blog not found")
	}
	attachBlogAuthor(b)
	return b, nil
}

func (s *BlogService) Publish(ctx context.Context, userID int64, req BlogCreateRequest) (*model.TravelBlog, error) {
	title, content, images, err := validateBlogInput(req.Title, req.Content, req.Images)
	if err != nil {
		return nil, err
	}
	blog := &model.TravelBlog{
		UserID:  userID,
		Title:   title,
		Content: content,
		Images:  images,
	}
	if err := s.blogs.Insert(ctx, blog); err != nil {
		return nil, err
	}
	return s.GetBlog(ctx, blog.ID, userID)
}

func (s *BlogService) Update(ctx context.Context, userID, id int64, req BlogCreateRequest) (*model.TravelBlog, error) {
	if err := s.requireOwner(ctx, id, userID); err != nil {
		return nil, err
	}
	title, content, images, err := validateBlogInput(req.Title, req.Content, req.Images)
	if err != nil {
		return nil, err
	}
	affected, err := s.blogs.Update(ctx, id, userID, title, content, images)
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		return nil, common.NotFound("Blog not found")
	}
	return s.GetBlog(ctx, id, userID)
}

func (s *BlogService) Delete(ctx context.Context, userID, id int64) error {
	if err := s.requireOwner(ctx, id, userID); err != nil {
		return err
	}
	affected, err := s.blogs.Delete(ctx, id, userID)
	if err != nil {
		return err
	}
	if affected == 0 {
		return common.NotFound("Blog not found")
	}
	return nil
}

func (s *BlogService) Like(ctx context.Context, userID, id int64) (*LikeResult, error) {
	blog, err := s.GetBlog(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	isNew, likes, err := s.blogs.SetLike(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if isNew {
		_ = s.notifications.Notify(ctx, blog.UserID, userID, NotifyLike, id, 0)
	}
	return &LikeResult{Likes: likes, LikedByMe: true}, nil
}

func (s *BlogService) Unlike(ctx context.Context, userID, id int64) (*LikeResult, error) {
	if _, err := s.GetBlog(ctx, id, userID); err != nil {
		return nil, err
	}
	_, likes, err := s.blogs.RemoveLike(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	return &LikeResult{Likes: likes, LikedByMe: false}, nil
}

func (s *BlogService) requireOwner(ctx context.Context, id, userID int64) error {
	authorID, err := s.blogs.AuthorID(ctx, id)
	if err != nil {
		return err
	}
	if authorID == 0 {
		return common.NotFound("Blog not found")
	}
	if authorID != userID {
		return common.Forbidden("Only the author can modify this blog")
	}
	return nil
}

func attachBlogAuthor(b *model.TravelBlog) {
	if b == nil {
		return
	}
	b.Author = &model.Author{ID: b.UserID, Nickname: b.AuthorNickname, AvatarURL: b.AuthorAvatarURL}
}

func validateBlogInput(title, content string, images []string) (string, string, model.StringList, error) {
	title = strings.TrimSpace(title)
	content = strings.TrimSpace(content)
	if title == "" || content == "" {
		return "", "", nil, common.BadRequest("title and content are required")
	}
	if utf8.RuneCountInString(title) > maxTitleRunes {
		return "", "", nil, common.BadRequest("title is too long (max 200 characters)")
	}
	if utf8.RuneCountInString(content) > maxContentRunes {
		return "", "", nil, common.BadRequest("content is too long (max 20000 characters)")
	}
	if len(images) > maxImagesPerBlog {
		return "", "", nil, common.BadRequest("too many images (max 9)")
	}
	clean := make(model.StringList, 0, len(images))
	for _, raw := range images {
		img := strings.TrimSpace(raw)
		if img == "" {
			continue
		}
		if len(img) > maxImageURLLen {
			return "", "", nil, common.BadRequest("image URL is too long")
		}
		if !strings.HasPrefix(img, "/uploads/") &&
			!strings.HasPrefix(img, "http://") &&
			!strings.HasPrefix(img, "https://") {
			return "", "", nil, common.BadRequest("image must be an uploaded file or an http(s) URL")
		}
		clean = append(clean, img)
	}
	return title, content, clean, nil
}

type BlogDraft struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

// Draft asks the model for a title and body from a topic or rough notes.
func (s *BlogService) Draft(ctx context.Context, topic string) (*BlogDraft, error) {
	systemPrompt := `You are a travel blogger writing about Guangzhou and Lingnan culture. Based on the user's topic or notes, produce a short blog post (200-300 words) with a friendly, vivid tone. Return ONLY valid JSON: { "title": string, "content": string }. Do NOT output markdown.`
	raw, err := s.qwen.ChatDefault(ctx, []ai.QwenMessage{
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
