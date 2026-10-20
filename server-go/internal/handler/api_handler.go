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
	list, err := h.blogs.ListFeed(c.Request.Context(), userID(c))
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, list)
}

func (h *BlogHandler) Detail(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
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
	if err := c.ShouldBindJSON(&req); err != nil || req.Title == "" || req.Content == "" {
		common.Fail(c, common.BadRequest("title and content are required"))
		return
	}
	blog, err := h.blogs.Publish(c.Request.Context(), id, req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.Created(c, blog)
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
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	likes, err := h.blogs.Like(c.Request.Context(), id)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, map[string]int{"likes": likes})
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
