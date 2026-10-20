package handler

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/travelgd/server-go/internal/auth"
	"github.com/travelgd/server-go/internal/common"
)

/** Middleware that requires a valid bearer token and stores the user id in the context. */
func RequireAuth(jwt *auth.JwtService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := requireUserID(c, jwt)
		if err != nil {
			common.Fail(c, err)
			c.Abort()
			return
		}
		c.Set("userID", id)
		c.Next()
	}
}

/** Sets userID in the context when a valid token is present; never aborts. */
func OptionalAuth(jwt *auth.JwtService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if id, err := requireUserID(c, jwt); err == nil {
			c.Set("userID", id)
		}
		c.Next()
	}
}

func requireUserID(c *gin.Context, jwt *auth.JwtService) (int64, error) {
	header := c.GetHeader("Authorization")
	if header == "" || !strings.HasPrefix(header, "Bearer ") {
		return 0, common.Unauthorized("Missing bearer token")
	}
	id, err := jwt.ParseUserID(header[7:])
	if err != nil {
		return 0, common.Unauthorized("Invalid or expired token")
	}
	return id, nil
}

/** userID returns the authenticated user id, or 0 for anonymous requests. */
func userID(c *gin.Context) int64 {
	if v, ok := c.Get("userID"); ok {
		if id, ok := v.(int64); ok {
			return id
		}
	}
	return 0
}

/** userIDRequired is like userID but fails the request with 401 when absent. */
func userIDRequired(c *gin.Context) (int64, bool) {
	id := userID(c)
	if id == 0 {
		common.Fail(c, common.Unauthorized("Missing bearer token"))
		return 0, false
	}
	return id, true
}
