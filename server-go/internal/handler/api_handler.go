package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/travelgd/server-go/internal/common"
	"github.com/travelgd/server-go/internal/service"
)

type RouteHandler struct {
	routes *service.RouteService
}

func NewRouteHandler(routes *service.RouteService) *RouteHandler {
	return &RouteHandler{routes: routes}
}

func (h *RouteHandler) Generate(c *gin.Context) {
	id, ok := userIDRequired(c)
	if !ok {
		return
	}
	var req service.RouteGenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	route, err := h.routes.Generate(c.Request.Context(), id, req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, route)
}

func (h *RouteHandler) List(c *gin.Context) {
	id, ok := userIDRequired(c)
	if !ok {
		return
	}
	list, err := h.routes.ListMine(c.Request.Context(), id)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, list)
}

func (h *RouteHandler) Detail(c *gin.Context) {
	id, ok := userIDRequired(c)
	if !ok {
		return
	}
	routeID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	route, err := h.routes.GetMine(c.Request.Context(), id, routeID)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, route)
}

type BlogHandler struct {
	blogs *service.BlogService
}

func NewBlogHandler(blogs *service.BlogService) *BlogHandler { return &BlogHandler{blogs: blogs} }

func (h *BlogHandler) Feed(c *gin.Context) {
	authorID, _ := strconv.ParseInt(c.DefaultQuery("authorId", "0"), 10, 64)
	page, pageSize := pageParams(c)
	res, err := h.blogs.Feed(c.Request.Context(), userID(c), authorID, c.Query("q"), page, pageSize)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, res)
}

func (h *BlogHandler) MyBlogs(c *gin.Context) {
	id, ok := userIDRequired(c)
	if !ok {
		return
	}
	page, pageSize := pageParams(c)
	res, err := h.blogs.Feed(c.Request.Context(), id, id, "", page, pageSize)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, res)
}

func (h *BlogHandler) Detail(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	blog, err := h.blogs.GetBlog(c.Request.Context(), id, userID(c))
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, blog)
}

func (h *BlogHandler) Publish(c *gin.Context) {
	id, ok := userIDRequired(c)
	if !ok {
		return
	}
	var req service.BlogCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	blog, err := h.blogs.Publish(c.Request.Context(), id, req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.Created(c, blog)
}

func (h *BlogHandler) Update(c *gin.Context) {
	id, ok := userIDRequired(c)
	if !ok {
		return
	}
	blogID, ok := idParam(c, "id")
	if !ok {
		return
	}
	var req service.BlogCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	blog, err := h.blogs.Update(c.Request.Context(), id, blogID, req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, blog)
}

func (h *BlogHandler) Delete(c *gin.Context) {
	id, ok := userIDRequired(c)
	if !ok {
		return
	}
	blogID, ok := idParam(c, "id")
	if !ok {
		return
	}
	if err := h.blogs.Delete(c.Request.Context(), id, blogID); err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, nil)
}

func (h *BlogHandler) Draft(c *gin.Context) {
	if _, ok := userIDRequired(c); !ok {
		return
	}
	var req struct {
		Topic string `json:"topic"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Topic == "" {
		common.Fail(c, common.BadRequest("topic is required"))
		return
	}
	draft, err := h.blogs.Draft(c.Request.Context(), req.Topic)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, draft)
}

func (h *BlogHandler) Like(c *gin.Context) {
	id, ok := userIDRequired(c)
	if !ok {
		return
	}
	blogID, ok := idParam(c, "id")
	if !ok {
		return
	}
	res, err := h.blogs.Like(c.Request.Context(), id, blogID)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, res)
}

func (h *BlogHandler) Unlike(c *gin.Context) {
	id, ok := userIDRequired(c)
	if !ok {
		return
	}
	blogID, ok := idParam(c, "id")
	if !ok {
		return
	}
	res, err := h.blogs.Unlike(c.Request.Context(), id, blogID)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, res)
}

type GuideHandler struct {
	guide *service.GuideService
}

func NewGuideHandler(guide *service.GuideService) *GuideHandler { return &GuideHandler{guide: guide} }

func (h *GuideHandler) Chat(c *gin.Context) {
	id, ok := userIDRequired(c)
	if !ok {
		return
	}
	var req service.GuideChatRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Question == "" {
		common.Fail(c, common.BadRequest("question is required"))
		return
	}
	resp, err := h.guide.Chat(c.Request.Context(), id, req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, resp)
}

func (h *GuideHandler) Session(c *gin.Context) {
	id, ok := userIDRequired(c)
	if !ok {
		return
	}
	sessionID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	session, err := h.guide.GetSession(c.Request.Context(), id, sessionID)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, session)
}
