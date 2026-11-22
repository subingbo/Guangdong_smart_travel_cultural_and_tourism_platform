package service

import (
	"context"

	"github.com/travelgd/server-go/internal/common"
	"github.com/travelgd/server-go/internal/model"
	"github.com/travelgd/server-go/internal/repo"
)

// UserService exposes public profiles and the follow graph.
type UserService struct {
	users         *repo.UserRepo
	blogs         *repo.BlogRepo
	follows       *repo.FollowRepo
	blogService   *BlogService
	notifications *NotificationService
}

func NewUserService(users *repo.UserRepo, blogs *repo.BlogRepo, follows *repo.FollowRepo,
	blogService *BlogService, notifications *NotificationService) *UserService {
	return &UserService{
		users: users, blogs: blogs, follows: follows,
		blogService: blogService, notifications: notifications,
	}
}

func (s *UserService) Profile(ctx context.Context, targetID, viewerID int64) (*model.UserProfile, error) {
	u, err := s.users.FindByID(ctx, targetID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, common.NotFound("User not found")
	}
	posts, err := s.blogs.CountByUser(ctx, targetID)
	if err != nil {
		return nil, err
	}
	followers, err := s.follows.CountFollowers(ctx, targetID)
	if err != nil {
		return nil, err
	}
	following, err := s.follows.CountFollowing(ctx, targetID)
	if err != nil {
		return nil, err
	}
	likes, err := s.blogs.SumLikesByUser(ctx, targetID)
	if err != nil {
		return nil, err
	}
	isFollowing, err := s.follows.IsFollowing(ctx, viewerID, targetID)
	if err != nil {
		return nil, err
	}
	return &model.UserProfile{
		User: model.UserPublic{
			ID:        u.ID,
			Username:  u.Username,
			Nickname:  u.Nickname,
			AvatarURL: u.AvatarURL,
			CreatedAt: u.CreatedAt,
		},
		Stats: model.ProfileStats{
			Posts:         posts,
			Followers:     followers,
			Following:     following,
			LikesReceived: likes,
		},
		IsFollowing: isFollowing,
	}, nil
}

func (s *UserService) Blogs(ctx context.Context, targetID, viewerID int64, page, pageSize int) (*BlogPage, error) {
	if u, err := s.users.FindByID(ctx, targetID); err != nil {
		return nil, err
	} else if u == nil {
		return nil, common.NotFound("User not found")
	}
	return s.blogService.Feed(ctx, viewerID, targetID, "", page, pageSize)
}

func (s *UserService) Follow(ctx context.Context, followerID, targetID int64) error {
	if followerID == targetID {
		return common.BadRequest("You cannot follow yourself")
	}
	if u, err := s.users.FindByID(ctx, targetID); err != nil {
		return err
	} else if u == nil {
		return common.NotFound("User not found")
	}
	isNew, err := s.follows.Follow(ctx, followerID, targetID)
	if err != nil {
		return err
	}
	if isNew {
		_ = s.notifications.Notify(ctx, targetID, followerID, NotifyFollow, 0, 0)
	}
	return nil
}

func (s *UserService) Unfollow(ctx context.Context, followerID, targetID int64) error {
	_, err := s.follows.Unfollow(ctx, followerID, targetID)
	return err
}

func (s *UserService) Followers(ctx context.Context, targetID int64, page, pageSize int) (*UserPage, error) {
	page, pageSize, offset := normalizePage(page, pageSize)
	items, total, err := s.follows.ListFollowers(ctx, targetID, offset, pageSize)
	if err != nil {
		return nil, err
	}
	return &UserPage{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *UserService) Following(ctx context.Context, targetID int64, page, pageSize int) (*UserPage, error) {
	page, pageSize, offset := normalizePage(page, pageSize)
	items, total, err := s.follows.ListFollowing(ctx, targetID, offset, pageSize)
	if err != nil {
		return nil, err
	}
	return &UserPage{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}
