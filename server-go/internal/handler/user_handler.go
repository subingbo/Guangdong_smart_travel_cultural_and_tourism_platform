package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/travelgd/server-go/internal/common"
	"github.com/travelgd/server-go/internal/service"
)

type UserHandler struct {
	users *service.UserService
}

func NewUserHandler(users *service.UserService) *UserHandler {
	return &UserHandler{users: users}
}

func (h *UserHandler) Profile(c *gin.Context) {
	targetID, ok := idParam(c, "id")
	if !ok {
		return
	}
	profile, err := h.users.Profile(c.Request.Context(), targetID, userID(c))
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, profile)
}

func (h *UserHandler) Blogs(c *gin.Context) {
	targetID, ok := idParam(c, "id")
	if !ok {
		return
	}
	page, pageSize := pageParams(c)
	res, err := h.users.Blogs(c.Request.Context(), targetID, userID(c), page, pageSize)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, res)
}

func (h *UserHandler) Follow(c *gin.Context) {
	viewerID, ok := userIDRequired(c)
	if !ok {
		return
	}
	targetID, ok := idParam(c, "id")
	if !ok {
		return
	}
	if err := h.users.Follow(c.Request.Context(), viewerID, targetID); err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, gin.H{"following": true})
}

func (h *UserHandler) Unfollow(c *gin.Context) {
	viewerID, ok := userIDRequired(c)
	if !ok {
		return
	}
	targetID, ok := idParam(c, "id")
	if !ok {
		return
	}
	if err := h.users.Unfollow(c.Request.Context(), viewerID, targetID); err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, gin.H{"following": false})
}

func (h *UserHandler) Followers(c *gin.Context) {
	targetID, ok := idParam(c, "id")
	if !ok {
		return
	}
	page, pageSize := pageParams(c)
	res, err := h.users.Followers(c.Request.Context(), targetID, page, pageSize)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, res)
}

func (h *UserHandler) Following(c *gin.Context) {
	targetID, ok := idParam(c, "id")
	if !ok {
		return
	}
	page, pageSize := pageParams(c)
	res, err := h.users.Following(c.Request.Context(), targetID, page, pageSize)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, res)
}
