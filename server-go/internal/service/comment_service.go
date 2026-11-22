package service

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/travelgd/server-go/internal/common"
	"github.com/travelgd/server-go/internal/model"
	"github.com/travelgd/server-go/internal/repo"
)

const maxCommentRunes = 1000

type CommentService struct {
	comments      *repo.CommentRepo
	blogs         *repo.BlogRepo
	notifications *NotificationService
}

func NewCommentService(comments *repo.CommentRepo, blogs *repo.BlogRepo, notifications *NotificationService) *CommentService {
	return &CommentService{comments: comments, blogs: blogs, notifications: notifications}
}

func (s *CommentService) List(ctx context.Context, blogID int64, page, pageSize int) (*CommentPage, error) {
	if authorID, err := s.blogs.AuthorID(ctx, blogID); err != nil {
		return nil, err
	} else if authorID == 0 {
		return nil, common.NotFound("Blog not found")
	}
	page, pageSize, offset := normalizePage(page, pageSize)
	items, total, err := s.comments.ListByBlog(ctx, blogID, offset, pageSize)
	if err != nil {
		return nil, err
	}
	for i := range items {
		items[i].Author = &model.Author{
			ID:        items[i].UserID,
			Nickname:  items[i].AuthorNickname,
			AvatarURL: items[i].AuthorAvatarURL,
		}
	}
	return &CommentPage{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *CommentService) Create(ctx context.Context, userID, blogID int64, content string) (*model.Comment, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, common.BadRequest("comment content is required")
	}
	if utf8.RuneCountInString(content) > maxCommentRunes {
		return nil, common.BadRequest("comment is too long (max 1000 characters)")
	}
	authorID, err := s.blogs.AuthorID(ctx, blogID)
	if err != nil {
		return nil, err
	}
	if authorID == 0 {
		return nil, common.NotFound("Blog not found")
	}
	c := &model.Comment{BlogID: blogID, UserID: userID, Content: content}
	if err := s.comments.Insert(ctx, c); err != nil {
		return nil, err
	}
	_ = s.notifications.Notify(ctx, authorID, userID, NotifyComment, blogID, c.ID)
	created, err := s.comments.FindByID(ctx, c.ID)
	if err != nil || created == nil {
		return created, err
	}
	created.Author = &model.Author{
		ID:        created.UserID,
		Nickname:  created.AuthorNickname,
		AvatarURL: created.AuthorAvatarURL,
	}
	return created, nil
}

func (s *CommentService) Delete(ctx context.Context, userID, commentID int64) error {
	c, err := s.comments.FindByID(ctx, commentID)
	if err != nil {
		return err
	}
	if c == nil {
		return common.NotFound("Comment not found")
	}
	blogAuthorID, err := s.blogs.AuthorID(ctx, c.BlogID)
	if err != nil {
		return err
	}
	if c.UserID != userID && blogAuthorID != userID {
		return common.Forbidden("Only the comment author or the blog author can delete this comment")
	}
	return s.comments.Delete(ctx, commentID)
}
