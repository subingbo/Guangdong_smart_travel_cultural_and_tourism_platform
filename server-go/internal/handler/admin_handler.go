package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/travelgd/server-go/internal/common"
	"github.com/travelgd/server-go/internal/model"
	"github.com/travelgd/server-go/internal/service"
)

type AdminHandler struct {
	admin *service.AdminService
}

func NewAdminHandler(admin *service.AdminService) *AdminHandler { return &AdminHandler{admin: admin} }

type knowledgePage struct {
	Items    []model.Knowledge `json:"items"`
	Total    int               `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"pageSize"`
}

// --- knowledge -----------------------------------------------------------

func (h *AdminHandler) ListKnowledge(c *gin.Context) {
	page, pageSize := pageParams(c)
	items, total, page, pageSize, err := h.admin.ListKnowledge(c.Request.Context(), c.Query("q"), page, pageSize)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, knowledgePage{Items: items, Total: total, Page: page, PageSize: pageSize})
}

func (h *AdminHandler) CreateKnowledge(c *gin.Context) {
	var req service.KnowledgeInput
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	res, err := h.admin.CreateKnowledge(c.Request.Context(), req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.Created(c, res)
}

func (h *AdminHandler) UpdateKnowledge(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	var req service.KnowledgeInput
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	res, err := h.admin.UpdateKnowledge(c.Request.Context(), id, req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, res)
}

func (h *AdminHandler) DeleteKnowledge(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	if err := h.admin.DeleteKnowledge(c.Request.Context(), id); err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, nil)
}

// --- products ------------------------------------------------------------

func (h *AdminHandler) ListProducts(c *gin.Context) {
	page, pageSize := pageParams(c)
	res, err := h.admin.ListProducts(c.Request.Context(), c.Query("q"), c.Query("category"), page, pageSize)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, res)
}

func (h *AdminHandler) CreateProduct(c *gin.Context) {
	var req service.ProductInput
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	res, err := h.admin.CreateProduct(c.Request.Context(), req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.Created(c, res)
}

func (h *AdminHandler) UpdateProduct(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	var req service.ProductInput
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	res, err := h.admin.UpdateProduct(c.Request.Context(), id, req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, res)
}

func (h *AdminHandler) DeleteProduct(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	if err := h.admin.DeleteProduct(c.Request.Context(), id); err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, nil)
}

// --- orders --------------------------------------------------------------

func (h *AdminHandler) ListOrders(c *gin.Context) {
	page, pageSize := pageParams(c)
	res, err := h.admin.ListOrders(c.Request.Context(), c.Query("status"), page, pageSize)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, res)
}

type orderStatusRequest struct {
	Status string `json:"status"`
}

func (h *AdminHandler) UpdateOrderStatus(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	var req orderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	if err := h.admin.UpdateOrderStatus(c.Request.Context(), id, req.Status); err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, nil)
}



