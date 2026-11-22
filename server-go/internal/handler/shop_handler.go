package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/travelgd/server-go/internal/common"
	"github.com/travelgd/server-go/internal/service"
)

type ShopHandler struct {
	shop *service.ShopService
}

func NewShopHandler(shop *service.ShopService) *ShopHandler { return &ShopHandler{shop: shop} }

func (h *ShopHandler) Products(c *gin.Context) {
	page, pageSize := pageParams(c)
	res, err := h.shop.ListProducts(c.Request.Context(), c.Query("category"), c.Query("q"), page, pageSize)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, res)
}

func (h *ShopHandler) Product(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	res, err := h.shop.GetProduct(c.Request.Context(), id)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, res)
}

func (h *ShopHandler) Cart(c *gin.Context) {
	userID, ok := userIDRequired(c)
	if !ok {
		return
	}
	res, err := h.shop.GetCart(c.Request.Context(), userID)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, res)
}

type cartAddRequest struct {
	ProductID int64 `json:"productId"`
	Quantity  int   `json:"quantity"`
}

func (h *ShopHandler) AddCart(c *gin.Context) {
	userID, ok := userIDRequired(c)
	if !ok {
		return
	}
	var req cartAddRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.ProductID <= 0 {
		common.Fail(c, common.BadRequest("productId and quantity are required"))
		return
	}
	if req.Quantity == 0 {
		req.Quantity = 1
	}
	res, err := h.shop.AddToCart(c.Request.Context(), userID, req.ProductID, req.Quantity)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, res)
}

type cartUpdateRequest struct {
	Quantity int `json:"quantity"`
}

func (h *ShopHandler) UpdateCart(c *gin.Context) {
	userID, ok := userIDRequired(c)
	if !ok {
		return
	}
	productID, ok := idParam(c, "productId")
	if !ok {
		return
	}
	var req cartUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest("quantity is required"))
		return
	}
	res, err := h.shop.UpdateCart(c.Request.Context(), userID, productID, req.Quantity)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, res)
}

func (h *ShopHandler) RemoveCart(c *gin.Context) {
	userID, ok := userIDRequired(c)
	if !ok {
		return
	}
	productID, ok := idParam(c, "productId")
	if !ok {
		return
	}
	res, err := h.shop.RemoveFromCart(c.Request.Context(), userID, productID)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, res)
}

func (h *ShopHandler) Checkout(c *gin.Context) {
	userID, ok := userIDRequired(c)
	if !ok {
		return
	}
	order, err := h.shop.Checkout(c.Request.Context(), userID, c.GetHeader("Idempotency-Key"))
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.Created(c, order)
}

func (h *ShopHandler) Orders(c *gin.Context) {
	userID, ok := userIDRequired(c)
	if !ok {
		return
	}
	page, pageSize := pageParams(c)
	res, err := h.shop.ListOrders(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, res)
}

func (h *ShopHandler) Order(c *gin.Context) {
	userID, ok := userIDRequired(c)
	if !ok {
		return
	}
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	res, err := h.shop.GetOrder(c.Request.Context(), userID, id)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, res)
}

func (h *ShopHandler) Reviews(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	res, err := h.shop.Reviews(c.Request.Context(), id)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, res)
}
