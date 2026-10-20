package service

import (
	"context"
	"fmt"
	"math/rand"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/travelgd/server-go/internal/auth"
	"github.com/travelgd/server-go/internal/common"
	"github.com/travelgd/server-go/internal/model"
	"github.com/travelgd/server-go/internal/repo"
)

type AuthService struct {
	users *repo.UserRepo
	jwt   *auth.JwtService
}

func NewAuthService(users *repo.UserRepo, jwt *auth.JwtService) *AuthService {
	return &AuthService{users: users, jwt: jwt}
}

type AuthResponse struct {
	Token string     `json:"token"`
	User  model.User `json:"user"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (s *AuthService) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	if len(req.Username) < 3 || len(req.Username) > 50 {
		return nil, common.BadRequest("username: must be 3-50 characters")
	}
	if len(req.Password) < 6 {
		return nil, common.BadRequest("password: must be at least 6 characters")
	}
	if n, err := s.users.CountByUsername(ctx, req.Username); err != nil {
		return nil, err
	} else if n > 0 {
		return nil, common.Conflict("Username already taken")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, common.Internal("Failed to hash password")
	}
	nickname := strings.TrimSpace(req.Nickname)
	if nickname == "" {
		nickname = "Guest User"
	}
	user := &model.User{
		Username:     req.Username,
		PasswordHash: string(hash),
		Nickname:     nickname,
		AvatarURL:    fmt.Sprintf("https://i.pravatar.cc/150?img=%d", rand.Intn(70)+1),
		Language:     "en",
	}
	if err := s.users.Insert(ctx, user); err != nil {
		return nil, err
	}
	return s.tokenFor(user)
}

func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	user, err := s.users.FindByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if user == nil || bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash), []byte(req.Password)) != nil {
		return nil, common.Unauthorized("Invalid username or password")
	}
	return s.tokenFor(user)
}

func (s *AuthService) tokenFor(user *model.User) (*AuthResponse, error) {
	token, err := s.jwt.Issue(user.ID, user.Username)
	if err != nil {
		return nil, common.Internal("Failed to issue token")
	}
	return &AuthResponse{Token: token, User: *user}, nil
}

/** Returns the authenticated user or nil when the token is missing/invalid. */
func (s *AuthService) CurrentUser(ctx context.Context, userID int64) (*model.User, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, common.Unauthorized("User no longer exists")
	}
	return user, nil
}
