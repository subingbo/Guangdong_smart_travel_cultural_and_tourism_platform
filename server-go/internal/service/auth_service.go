package service

import (
	"context"
	cryptorand "crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/travelgd/server-go/internal/auth"
	"github.com/travelgd/server-go/internal/common"
	"github.com/travelgd/server-go/internal/model"
	"github.com/travelgd/server-go/internal/repo"
)

type AuthService struct {
	users    *repo.UserRepo
	jwt      *auth.JwtService
	wxAppID  string
	wxSecret string
}

func NewAuthService(users *repo.UserRepo, jwt *auth.JwtService, wxAppID, wxSecret string) *AuthService {
	return &AuthService{users: users, jwt: jwt, wxAppID: wxAppID, wxSecret: wxSecret}
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
		Role:         "user",
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

type WechatLoginRequest struct {
	Code     string `json:"code"`
	DeviceID string `json:"deviceId"`
	Nickname string `json:"nickname"`
}

// WechatLogin exchanges a wx.login code for a session-backed user.
//
// With WX_APPID/WX_SECRET configured the code goes through jscode2session.
// Without them (local development) the client's stable deviceId is hashed
// into a pseudo openid, so repeated logins in a demo resolve to the same
// account without any WeChat secret on the machine.
func (s *AuthService) WechatLogin(ctx context.Context, req WechatLoginRequest) (*AuthResponse, error) {
	openID, err := s.resolveOpenID(ctx, req)
	if err != nil {
		return nil, err
	}
	// users.username is 50 chars: "wx_" + 16 hex chars of the openid digest.
	digest := sha256.Sum256([]byte(openID))
	username := "wx_" + hex.EncodeToString(digest[:])[:16]

	user, err := s.users.FindByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		random := make([]byte, 24)
		if _, err := cryptorand.Read(random); err != nil {
			return nil, common.Internal("Failed to create user")
		}
		hash, err := bcrypt.GenerateFromPassword(random, bcrypt.DefaultCost)
		if err != nil {
			return nil, common.Internal("Failed to hash password")
		}
		nickname := strings.TrimSpace(req.Nickname)
		if nickname == "" {
			nickname = "微信用户"
		}
		user = &model.User{
			Username:     username,
			PasswordHash: string(hash),
			Nickname:     nickname,
			AvatarURL:    fmt.Sprintf("https://i.pravatar.cc/150?img=%d", rand.Intn(70)+1),
			Language:     "zh",
			Role:         "user",
		}
		if err := s.users.Insert(ctx, user); err != nil {
			return nil, err
		}
	}
	return s.tokenFor(user)
}

func (s *AuthService) resolveOpenID(ctx context.Context, req WechatLoginRequest) (string, error) {
	if s.wxAppID != "" && s.wxSecret != "" {
		if strings.TrimSpace(req.Code) == "" {
			return "", common.BadRequest("code is required")
		}
		endpoint := fmt.Sprintf(
			"https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
			url.QueryEscape(s.wxAppID), url.QueryEscape(s.wxSecret), url.QueryEscape(req.Code))
		httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return "", common.Internal("WeChat request build failed")
		}
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(httpReq)
		if err != nil {
			return "", common.BadGateway("WeChat service unreachable: " + err.Error())
		}
		defer resp.Body.Close()
		var out struct {
			OpenID  string `json:"openid"`
			ErrCode int    `json:"errcode"`
			ErrMsg  string `json:"errmsg"`
		}
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
		if err := json.Unmarshal(body, &out); err != nil {
			return "", common.BadGateway("Unexpected WeChat response")
		}
		if out.ErrCode != 0 || out.OpenID == "" {
			return "", common.Unauthorized("WeChat login failed: "+out.ErrMsg)
		}
		return out.OpenID, nil
	}

	// Local development: derive a stable identity from the device id.
	if strings.TrimSpace(req.DeviceID) == "" {
		return "", common.BadRequest("deviceId is required (or configure WX_APPID/WX_SECRET)")
	}
	return "dev_" + strings.TrimSpace(req.DeviceID), nil
}

func (s *AuthService) tokenFor(user *model.User) (*AuthResponse, error) {
	token, err := s.jwt.Issue(user.ID, user.Username)
	if err != nil {
		return nil, common.Internal("Failed to issue token")
	}
	return &AuthResponse{Token: token, User: *user}, nil
}

// CurrentUser loads the user behind the token or fails when the row is gone.
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
