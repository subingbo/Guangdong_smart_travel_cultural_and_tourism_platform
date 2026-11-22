package handler

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/travelgd/server-go/internal/auth"
	"github.com/travelgd/server-go/internal/common"
	"github.com/travelgd/server-go/internal/repo"
)

// RequireAuth rejects requests without a valid bearer token.
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

// OptionalAuth sets userID when a token is present but never aborts.
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

// userID returns the authenticated user id, or 0 for anonymous requests.
func userID(c *gin.Context) int64 {
	if v, ok := c.Get("userID"); ok {
		if id, ok := v.(int64); ok {
			return id
		}
	}
	return 0
}

// userIDRequired is userID plus a 401 response when the user is anonymous.
func userIDRequired(c *gin.Context) (int64, bool) {
	id := userID(c)
	if id == 0 {
		common.Fail(c, common.Unauthorized("Missing bearer token"))
		return 0, false
	}
	return id, true
}

// RequireRole checks the current user's role against the database on every
// request, so a demoted user cannot keep admin access with an old JWT.
func RequireRole(users *repo.UserRepo, roles ...string) gin.HandlerFunc {
	allowed := make(map[string]bool, len(roles))
	for _, role := range roles {
		allowed[role] = true
	}
	return func(c *gin.Context) {
		id := userID(c)
		if id == 0 {
			common.Fail(c, common.Unauthorized("Missing bearer token"))
			c.Abort()
			return
		}
		user, err := users.FindByID(c.Request.Context(), id)
		if err != nil {
			common.Fail(c, err)
			c.Abort()
			return
		}
		if user == nil {
			common.Fail(c, common.Unauthorized("User no longer exists"))
			c.Abort()
			return
		}
		if !allowed[user.Role] {
			common.Fail(c, common.Forbidden("Insufficient permissions"))
			c.Abort()
			return
		}
		c.Set("role", user.Role)
		c.Next()
	}
}
