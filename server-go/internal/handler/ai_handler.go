package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/travelgd/server-go/internal/common"
	"github.com/travelgd/server-go/internal/service"
)

type AiHandler struct {
	ai *service.AiService
}

func NewAiHandler(ai *service.AiService) *AiHandler { return &AiHandler{ai: ai} }

func (h *AiHandler) Copy(c *gin.Context) {
	var req service.CopyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	resp, err := h.ai.Copy(c.Request.Context(), req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, resp)
}

func (h *AiHandler) CreatePoster(c *gin.Context) {
	var req service.PosterCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	resp, err := h.ai.CreatePoster(c.Request.Context(), req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, resp)
}

func (h *AiHandler) PosterTask(c *gin.Context) {
	task, err := h.ai.PosterTask(c.Request.Context(), c.Param("taskId"))
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, task)
}

func (h *AiHandler) MenuOCR(c *gin.Context) {
	var req service.MenuImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	resp, err := h.ai.MenuOCR(c.Request.Context(), req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, resp)
}

func (h *AiHandler) MenuTranslate(c *gin.Context) {
	var req service.MenuImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	resp, err := h.ai.MenuTranslate(c.Request.Context(), req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, resp)
}

func (h *AiHandler) DishExplain(c *gin.Context) {
	var req service.DishExplainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	resp, err := h.ai.DishExplain(c.Request.Context(), req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, resp)
}
