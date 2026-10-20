package service

import (
	"context"
	"strings"

	"github.com/travelgd/server-go/internal/common"
	"github.com/travelgd/server-go/internal/model"
	"github.com/travelgd/server-go/internal/repo"
)

type MeService struct {
	auth      *AuthService
	users     *repo.UserRepo
	favorites *repo.FavoriteRepo
	routes    *repo.RouteRepo
	blogs     *repo.BlogRepo
}

func NewMeService(auth *AuthService, users *repo.UserRepo, favorites *repo.FavoriteRepo,
	routes *repo.RouteRepo, blogs *repo.BlogRepo) *MeService {
	return &MeService{auth: auth, users: users, favorites: favorites, routes: routes, blogs: blogs}
}

type Stats struct {
	Favorites int `json:"favorites"`
	Routes    int `json:"routes"`
	Posts     int `json:"posts"`
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
	if req.Nickname != nil && strings.TrimSpace(*req.Nickname) != "" {
		user.Nickname = *req.Nickname
	}
	if req.AvatarURL != nil {
		user.AvatarURL = *req.AvatarURL
	}
	if req.Language != nil {
		lang := strings.ToLower(*req.Language)
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
	return &Stats{Favorites: favs, Routes: rts, Posts: posts}, nil
}
