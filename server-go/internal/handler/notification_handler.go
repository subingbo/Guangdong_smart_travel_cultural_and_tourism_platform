package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/travelgd/server-go/internal/common"
	"github.com/travelgd/server-go/internal/service"
)

type NotificationHandler struct {
	notifications *service.NotificationService
}

func NewNotificationHandler(notifications *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{notifications: notifications}
}

func (h *NotificationHandler) List(c *gin.Context) {
	userID, ok := userIDRequired(c)
	if !ok {
		return
	}
	page, pageSize := pageParams(c)
	unreadOnly := c.Query("filter") == "unread"
	res, err := h.notifications.List(c.Request.Context(), userID, unreadOnly, page, pageSize)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, res)
}

func (h *NotificationHandler) MarkRead(c *gin.Context) {
	userID, ok := userIDRequired(c)
	if !ok {
		return
	}
	var req struct {
		IDs []int64 `json:"ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	if err := h.notifications.MarkRead(c.Request.Context(), userID, req.IDs); err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, nil)
}

func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	userID, ok := userIDRequired(c)
	if !ok {
		return
	}
	if err := h.notifications.MarkAllRead(c.Request.Context(), userID); err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, nil)
}

func (h *NotificationHandler) UnreadCount(c *gin.Context) {
	userID, ok := userIDRequired(c)
	if !ok {
		return
	}
	n, err := h.notifications.UnreadCount(c.Request.Context(), userID)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, gin.H{"unread": n})
}
