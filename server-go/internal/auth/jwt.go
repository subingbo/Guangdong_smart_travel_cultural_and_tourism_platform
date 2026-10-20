package auth

import (
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JwtService struct {
	key         []byte
	expireHours time.Duration
}

func NewJwtService(secret string, expireHours int) *JwtService {
	return &JwtService{key: []byte(secret), expireHours: time.Duration(expireHours) * time.Hour}
}

func (s *JwtService) Issue(userId int64, username string) (string, error) {
	claims := jwt.MapClaims{
		"sub":      strconv.FormatInt(userId, 10),
		"username": username,
		"iat":      time.Now().Unix(),
		"exp":      time.Now().Add(s.expireHours).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.key)
}

/** Returns the user id embedded in the token, or an error if invalid/expired. */
func (s *JwtService) ParseUserID(token string) (int64, error) {
	parsed, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.key, nil
	})
	if err != nil || !parsed.Valid {
		return 0, errors.New("invalid or expired token")
	}
	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid claims")
	}
	sub, ok := claims["sub"].(string)
	if !ok {
		return 0, errors.New("missing subject")
	}
	id, err := strconv.ParseInt(sub, 10, 64)
	if err != nil {
		return 0, errors.New("bad subject")
	}
	return id, nil
}
