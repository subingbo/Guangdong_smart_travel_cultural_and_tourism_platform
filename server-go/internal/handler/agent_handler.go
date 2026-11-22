package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/travelgd/server-go/internal/agent"
	"github.com/travelgd/server-go/internal/common"
	"github.com/travelgd/server-go/internal/service"
)

type AgentHandler struct {
	runs  *service.AgentRunService
	plans *service.AgentPlanService
	guide *service.GuideService
	debug bool
}

func NewAgentHandler(runs *service.AgentRunService, plans *service.AgentPlanService, guide *service.GuideService, debug bool) *AgentHandler {
	return &AgentHandler{runs: runs, plans: plans, guide: guide, debug: debug}
}

// CreateRun starts a run and returns its snapshot. The client then opens the
// SSE endpoint; polling GET /runs/:id is the fallback.
func (h *AgentHandler) CreateRun(c *gin.Context) {
	userID, ok := userIDRequired(c)
	if !ok {
		return
	}
	var req service.AgentRunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest("invalid request body"))
		return
	}
	snapshot, err := h.runs.Submit(c.Request.Context(), userID, req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.JSON(http.StatusAccepted, common.Envelope{Success: true, Data: snapshot})
}

func (h *AgentHandler) GetRun(c *gin.Context) {
	userID, ok := userIDRequired(c)
	if !ok {
		return
	}
	runID, ok := idParam(c, "id")
	if !ok {
		return
	}
	snapshot, err := h.runs.GetSnapshot(c.Request.Context(), userID, runID, true)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, snapshot)
}

func (h *AgentHandler) Cancel(c *gin.Context) {
	userID, ok := userIDRequired(c)
	if !ok {
		return
	}
	runID, ok := idParam(c, "id")
	if !ok {
		return
	}
	snapshot, err := h.runs.Cancel(c.Request.Context(), userID, runID)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, snapshot)
}

// Stream subscribes to a run. It writes one snapshot event first, then live
// events, with a heartbeat comment every 15 seconds.
func (h *AgentHandler) Stream(c *gin.Context) {
	userID, ok := userIDRequired(c)
	if !ok {
		return
	}
	runID, ok := idParam(c, "id")
	if !ok {
		return
	}
	ch, cancel, snapshot, err := h.runs.Subscribe(c.Request.Context(), userID, runID)
	if err != nil {
		common.Fail(c, err)
		return
	}
	defer cancel()

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)

	if snapshot != nil {
		if !writeSSE(c.Writer, agent.Event{
			Type: agent.EventSnapshot, Seq: snapshot.Seq, RunID: runID, Data: snapshot,
		}) {
			return
		}
	}

	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()
	for {
		select {
		case event, open := <-ch:
			if !open {
				return
			}
			if !writeSSE(c.Writer, event) {
				return
			}
			if event.Type == agent.EventDone || event.Type == agent.EventError {
				return
			}
		case <-heartbeat.C:
			if _, err := fmt.Fprint(c.Writer, ": ping\n\n"); err != nil {
				return
			}
			c.Writer.Flush()
		case <-c.Request.Context().Done():
			return
		}
	}
}

func (h *AgentHandler) ConfirmPlan(c *gin.Context) {
	userID, ok := userIDRequired(c)
	if !ok {
		return
	}
	runID, ok := idParam(c, "id")
	if !ok {
		return
	}
	var req struct {
		Version        int    `json:"version"`
		IdempotencyKey string `json:"idempotencyKey"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Version <= 0 || req.IdempotencyKey == "" {
		common.Fail(c, common.BadRequest("version and idempotencyKey are required"))
		return
	}
	result, err := h.plans.Confirm(c.Request.Context(), userID, runID, req.Version, req.IdempotencyKey)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, result)
}

func (h *AgentHandler) ListConversations(c *gin.Context) {
	userID, ok := userIDRequired(c)
	if !ok {
		return
	}
	page, pageSize := pageParams(c)
	result, err := h.runs.ListConversations(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, result)
}

func (h *AgentHandler) GetConversation(c *gin.Context) {
	userID, ok := userIDRequired(c)
	if !ok {
		return
	}
	conversationID, ok := idParam(c, "id")
	if !ok {
		return
	}
	session, err := h.guide.GetSession(c.Request.Context(), userID, conversationID)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, session)
}

func (h *AgentHandler) Metrics(c *gin.Context) {
	if !h.debug {
		common.Fail(c, common.NotFound("Not found"))
		return
	}
	common.OK(c, h.runs.Metrics())
}

func writeSSE(w gin.ResponseWriter, event agent.Event) bool {
	raw, err := json.Marshal(event)
	if err != nil {
		return false
	}
	if _, err := fmt.Fprintf(w, "id: %d\nevent: %s\ndata: %s\n\n", event.Seq, event.Type, raw); err != nil {
		return false
	}
	w.Flush()
	return true
}
