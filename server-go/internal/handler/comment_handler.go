package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/travelgd/server-go/internal/common"
	"github.com/travelgd/server-go/internal/service"
)

type CommentHandler struct {
	comments *service.CommentService
}

func NewCommentHandler(comments *service.CommentService) *CommentHandler {
	return &CommentHandler{comments: comments}
}

func (h *CommentHandler) List(c *gin.Context) {
	blogID, ok := idParam(c, "id")
	if !ok {
		return
	}
	page, pageSize := pageParams(c)
	res, err := h.comments.List(c.Request.Context(), blogID, page, pageSize)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, res)
}

func (h *CommentHandler) Create(c *gin.Context) {
	userID, ok := userIDRequired(c)
	if !ok {
		return
	}
	blogID, ok := idParam(c, "id")
	if !ok {
		return
	}
	var req struct {
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	comment, err := h.comments.Create(c.Request.Context(), userID, blogID, req.Content)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.Created(c, comment)
}

func (h *CommentHandler) Delete(c *gin.Context) {
	userID, ok := userIDRequired(c)
	if !ok {
		return
	}
	commentID, ok := idParam(c, "id")
	if !ok {
		return
	}
	if err := h.comments.Delete(c.Request.Context(), userID, commentID); err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, nil)
}
