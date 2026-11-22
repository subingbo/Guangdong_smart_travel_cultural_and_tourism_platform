package service

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/travelgd/server-go/internal/common"
	"github.com/travelgd/server-go/internal/model"
	"github.com/travelgd/server-go/internal/repo"
)

const maxNicknameRunes = 50

type MeService struct {
	auth          *AuthService
	users         *repo.UserRepo
	favorites     *repo.FavoriteRepo
	routes        *repo.RouteRepo
	blogs         *repo.BlogRepo
	follows       *repo.FollowRepo
	notifications *repo.NotificationRepo
}

func NewMeService(auth *AuthService, users *repo.UserRepo, favorites *repo.FavoriteRepo,
	routes *repo.RouteRepo, blogs *repo.BlogRepo, follows *repo.FollowRepo,
	notifications *repo.NotificationRepo) *MeService {
	return &MeService{
		auth: auth, users: users, favorites: favorites, routes: routes,
		blogs: blogs, follows: follows, notifications: notifications,
	}
}

type Stats struct {
	Favorites           int `json:"favorites"`
	Routes              int `json:"routes"`
	Posts               int `json:"posts"`
	Followers           int `json:"followers"`
	Following           int `json:"following"`
	LikesReceived       int `json:"likesReceived"`
	UnreadNotifications int `json:"unreadNotifications"`
}

type ProfileUpdate struct {
	Nickname  *string `json:"nickname"`
	AvatarURL *string `json:"avatarUrl"`
	Language  *string `json:"language"`
}

func (s *MeService) GetProfile(ctx context.Context, userID int64) (*model.User, error) {
	return s.auth.CurrentUser(ctx, userID)
}

func (s *MeService) UpdateProfile(ctx context.Context, userID int64, req ProfileUpdate) (*model.User, error) {
	user, err := s.auth.CurrentUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if req.Nickname != nil {
		nickname := strings.TrimSpace(*req.Nickname)
		if nickname == "" || utf8.RuneCountInString(nickname) > maxNicknameRunes {
			return nil, common.BadRequest("nickname must be 1-50 characters")
		}
		user.Nickname = nickname
	}
	if req.AvatarURL != nil {
		avatar := strings.TrimSpace(*req.AvatarURL)
		if avatar != "" {
			if len(avatar) > maxImageURLLen {
				return nil, common.BadRequest("avatar URL is too long")
			}
			if !strings.HasPrefix(avatar, "/uploads/") &&
				!strings.HasPrefix(avatar, "http://") &&
				!strings.HasPrefix(avatar, "https://") {
				return nil, common.BadRequest("avatar must be an uploaded file or an http(s) URL")
			}
		}
		user.AvatarURL = avatar
	}
	if req.Language != nil {
		lang := strings.ToLower(strings.TrimSpace(*req.Language))
		if lang != "en" && lang != "zh" {
			return nil, common.BadRequest("language must be 'en' or 'zh'")
		}
		user.Language = lang
	}
	if err := s.users.UpdateProfile(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *MeService) GetStats(ctx context.Context, userID int64) (*Stats, error) {
	favs, err := s.favorites.CountByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	rts, err := s.routes.CountByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	posts, err := s.blogs.CountByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	followers, err := s.follows.CountFollowers(ctx, userID)
	if err != nil {
		return nil, err
	}
	following, err := s.follows.CountFollowing(ctx, userID)
	if err != nil {
		return nil, err
	}
	likes, err := s.blogs.SumLikesByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	unread, err := s.notifications.UnreadCount(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &Stats{
		Favorites:           favs,
		Routes:              rts,
		Posts:               posts,
		Followers:           followers,
		Following:           following,
		LikesReceived:       likes,
		UnreadNotifications: unread,
	}, nil
}
