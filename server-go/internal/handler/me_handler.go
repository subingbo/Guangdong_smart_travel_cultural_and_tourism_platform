package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/travelgd/server-go/internal/common"
	"github.com/travelgd/server-go/internal/model"
	"github.com/travelgd/server-go/internal/service"
)

type MeHandler struct {
	me *service.MeService
}

func NewMeHandler(me *service.MeService) *MeHandler { return &MeHandler{me: me} }

func (h *MeHandler) Profile(c *gin.Context) {
	id, ok := userIDRequired(c)
	if !ok {
		return
	}
	user, err := h.me.GetProfile(c.Request.Context(), id)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, user)
}

func (h *MeHandler) Update(c *gin.Context) {
	id, ok := userIDRequired(c)
	if !ok {
		return
	}
	var req service.ProfileUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	user, err := h.me.UpdateProfile(c.Request.Context(), id, req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, user)
}

func (h *MeHandler) Stats(c *gin.Context) {
	id, ok := userIDRequired(c)
	if !ok {
		return
	}
	stats, err := h.me.GetStats(c.Request.Context(), id)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, stats)
}

type CatalogHandler struct {
	catalog *service.CatalogService
}

func NewCatalogHandler(catalog *service.CatalogService) *CatalogHandler {
	return &CatalogHandler{catalog: catalog}
}

func (h *CatalogHandler) ListScenics(c *gin.Context) {
	list, err := h.catalog.ListScenics(c.Request.Context())
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, list)
}

func (h *CatalogHandler) GetScenic(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	scenic, err := h.catalog.GetScenic(c.Request.Context(), id)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, scenic)
}

func (h *CatalogHandler) ListPlaces(c *gin.Context) {
	list, err := h.catalog.ListPlaces(c.Request.Context(), c.Query("tag"))
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, list)
}

func (h *CatalogHandler) GetPlace(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	place, err := h.catalog.GetPlace(c.Request.Context(), id)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, place)
}

type FavoriteHandler struct {
	favorites *service.FavoriteService
}

func NewFavoriteHandler(favorites *service.FavoriteService) *FavoriteHandler {
	return &FavoriteHandler{favorites: favorites}
}

func (h *FavoriteHandler) List(c *gin.Context) {
	id, ok := userIDRequired(c)
	if !ok {
		return
	}
	list, err := h.favorites.List(c.Request.Context(), id)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, list)
}

func (h *FavoriteHandler) Add(c *gin.Context) {
	id, ok := userIDRequired(c)
	if !ok {
		return
	}
	var req struct {
		PlaceID int64 `json:"placeId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.PlaceID == 0 {
		common.Fail(c, common.BadRequest("placeId is required"))
		return
	}
	if err := h.favorites.Add(c.Request.Context(), id, req.PlaceID); err != nil {
		common.Fail(c, err)
		return
	}
	common.Created(c, nil)
}

func (h *FavoriteHandler) Remove(c *gin.Context) {
	id, ok := userIDRequired(c)
	if !ok {
		return
	}
	placeID, _ := strconv.ParseInt(c.Param("placeId"), 10, 64)
	if err := h.favorites.Remove(c.Request.Context(), id, placeID); err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, nil)
}

var _ = model.Place{}
