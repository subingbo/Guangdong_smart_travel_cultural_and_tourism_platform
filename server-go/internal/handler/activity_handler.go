package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/travelgd/server-go/internal/common"
	"github.com/travelgd/server-go/internal/service"
)

type ActivityHandler struct {
	activities *service.ActivityService
}

func NewActivityHandler(activities *service.ActivityService) *ActivityHandler {
	return &ActivityHandler{activities: activities}
}

func (h *ActivityHandler) List(c *gin.Context) {
	page, pageSize := pageParams(c)
	res, err := h.activities.List(c.Request.Context(), c.DefaultQuery("status", "upcoming"), c.Query("q"), page, pageSize)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, res)
}

func (h *ActivityHandler) Detail(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	res, err := h.activities.Get(c.Request.Context(), id)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, res)
}
