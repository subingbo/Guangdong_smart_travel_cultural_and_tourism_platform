package service

import (
	"context"

	"github.com/travelgd/server-go/internal/model"
	"github.com/travelgd/server-go/internal/repo"
)

// NotificationService creates and lists in-app notifications.
type NotificationService struct {
	notifications *repo.NotificationRepo
}

func NewNotificationService(notifications *repo.NotificationRepo) *NotificationService {
	return &NotificationService{notifications: notifications}
}

const (
	NotifyLike    = "like"
	NotifyComment = "comment"
	NotifyFollow  = "follow"
)

// Notify never notifies the actor about their own action.
func (s *NotificationService) Notify(ctx context.Context, userID, actorID int64, kind string, blogID, commentID int64) error {
	if userID <= 0 || userID == actorID {
		return nil
	}
	return s.notifications.Insert(ctx, &model.Notification{
		UserID:    userID,
		ActorID:   actorID,
		Type:      kind,
		BlogID:    blogID,
		CommentID: commentID,
	})
}

func (s *NotificationService) List(ctx context.Context, userID int64, unreadOnly bool, page, pageSize int) (*NotificationPage, error) {
	page, pageSize, offset := normalizePage(page, pageSize)
	items, total, err := s.notifications.List(ctx, userID, unreadOnly, offset, pageSize)
	if err != nil {
		return nil, err
	}
	unread, err := s.notifications.UnreadCount(ctx, userID)
	if err != nil {
		return nil, err
	}
	for i := range items {
		items[i].Actor = &model.Author{
			ID:        items[i].ActorID,
			Nickname:  items[i].ActorNickname,
			AvatarURL: items[i].ActorAvatarURL,
		}
	}
	return &NotificationPage{Items: items, Total: total, Unread: unread, Page: page, PageSize: pageSize}, nil
}

func (s *NotificationService) UnreadCount(ctx context.Context, userID int64) (int, error) {
	return s.notifications.UnreadCount(ctx, userID)
}

func (s *NotificationService) MarkRead(ctx context.Context, userID int64, ids []int64) error {
	return s.notifications.MarkRead(ctx, userID, ids)
}

func (s *NotificationService) MarkAllRead(ctx context.Context, userID int64) error {
	return s.notifications.MarkAllRead(ctx, userID)
}
