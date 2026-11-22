package service

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/travelgd/server-go/internal/agent"
	"github.com/travelgd/server-go/internal/common"
	"github.com/travelgd/server-go/internal/model"
	"github.com/travelgd/server-go/internal/repo"
)

type AgentRunConfig struct {
	Enabled            bool
	Model              string
	MaxConcurrentRuns  int
	QueueCapacity      int
	QueueTimeout       time.Duration
	RunTimeout         time.Duration
	HistoryMessages    int
	MaxOutputBytes     int
	RateLimitPerMinute int
	Debug              bool
}

type AgentRunRequest struct {
	ConversationID  int64  `json:"conversationId"`
	Message         string `json:"message"`
	ClientRequestID string `json:"clientRequestId"`
}

type AgentError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type AgentSnapshot struct {
	RunID          int64             `json:"runId"`
	ConversationID int64             `json:"conversationId"`
	Status         string            `json:"status"`
	Output         string            `json:"output"`
	Plan           json.RawMessage   `json:"plan,omitempty"`
	PlanVersion    int               `json:"planVersion"`
	StepCount      int               `json:"stepCount"`
	Usage          agent.Usage       `json:"usage"`
	Error          *AgentError       `json:"error,omitempty"`
	CreatedAt      time.Time         `json:"createdAt"`
	StartedAt      *time.Time        `json:"startedAt,omitempty"`
	FinishedAt     *time.Time        `json:"finishedAt,omitempty"`
	Seq            int64             `json:"seq"`
	Steps          []model.AgentStep `json:"steps,omitempty"`
}

type AgentMetrics struct {
	Goroutines      int    `json:"goroutines"`
	HeapAlloc       uint64 `json:"heapAlloc"`
	NumGC           uint32 `json:"numGC"`
	ActiveRuns      int    `json:"activeRuns"`
	ExecutorMode    string `json:"executorMode"`
	ExecutorRunning int    `json:"executorRunning"`
	ExecutorQueued  int    `json:"executorQueued"`
}

var errQueueTimeout = errors.New("agent: queued run timed out")

type runControl struct {
	runID          int64
	conversationID int64
	userID         int64
	assistantMsgID int64
	ctx            context.Context
	cancel         context.CancelFunc
	seq            int64
	output         strings.Builder
	planJSON       json.RawMessage
	planVersion    int
	stepCount      int
	usage          agent.Usage
	status         string
	startedAt      time.Time
	dirty          bool
	flushInFlight  bool
	finished       bool
	timedOut       bool
	toolCalls      []map[string]string
}

type subscriber struct {
	ch chan agent.Event
}

// AgentRunService owns the run state machine, SSE fan-out, admission callback,
// progress flushing and recovery. It is single-instance by design.
type AgentRunService struct {
	db       *sqlx.DB
	guide    *repo.GuideRepo
	runs     *repo.AgentRepo
	users    *repo.UserRepo
	runner   *agent.Runner
	executor agent.Executor
	cfg      AgentRunConfig
	now      func() time.Time
	loc      *time.Location

	mu          sync.Mutex
	active      map[int64]*runControl
	subscribers map[int64]map[int64]*subscriber
	nextSubID   int64
	rate        map[int64][]time.Time
	stopped     bool

	steps    chan model.AgentStep
	stopOnce sync.Once
	stopCh   chan struct{}
	flushSem chan struct{}
}

func NewAgentRunService(
	db *sqlx.DB,
	guide *repo.GuideRepo,
	runs *repo.AgentRepo,
	users *repo.UserRepo,
	runner *agent.Runner,
	executor agent.Executor,
	cfg AgentRunConfig,
) *AgentRunService {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.FixedZone("CST", 8*3600)
	}
	if cfg.HistoryMessages <= 0 {
		cfg.HistoryMessages = 12
	}
	if cfg.MaxOutputBytes <= 0 {
		cfg.MaxOutputBytes = 64 * 1024
	}
	if cfg.RateLimitPerMinute <= 0 {
		cfg.RateLimitPerMinute = 10
	}
	s := &AgentRunService{
		db: db, guide: guide, runs: runs, users: users,
		runner: runner, executor: executor, cfg: cfg,
		now: time.Now, loc: loc,
		active:      make(map[int64]*runControl),
		subscribers: make(map[int64]map[int64]*subscriber),
		rate:        make(map[int64][]time.Time),
		steps:       make(chan model.AgentStep, 128),
		stopCh:      make(chan struct{}),
		flushSem:    make(chan struct{}, 2),
	}
	go s.tickLoop()
	go s.stepWriter()
	return s
}

func (s *AgentRunService) Enabled() bool { return s.cfg.Enabled }

func (s *AgentRunService) RecoverInterrupted(ctx context.Context) error {
	n, err := s.runs.RecoverInterrupted(ctx)
	if err == nil && n > 0 {
		log.Printf("agent: marked %d interrupted runs", n)
	}
	return err
}

func (s *AgentRunService) Submit(ctx context.Context, userID int64, req AgentRunRequest) (*AgentSnapshot, error) {
	if !s.cfg.Enabled {
		return nil, common.ServiceUnavailable("AI agent is disabled (AGENT_ENABLED=false)")
	}
	message := strings.TrimSpace(req.Message)
	if message == "" || len([]rune(message)) > 4000 {
		return nil, common.BadRequest("message must be 1-4000 characters")
	}
	clientRequestID := strings.TrimSpace(req.ClientRequestID)
	if len(clientRequestID) > 36 {
		return nil, common.BadRequest("clientRequestId is too long")
	}

	if clientRequestID != "" {
		existing, err := s.runs.FindRunByClientRequestID(ctx, userID, clientRequestID)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			if existing.Input != message {
				return nil, common.Conflict("clientRequestId was already used with a different message")
			}
			return s.GetSnapshot(ctx, userID, existing.ID, false)
		}
	}
	if !s.allowRate(userID) {
		return nil, common.TooManyRequests("请求过于频繁，请稍后再试")
	}

	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, common.Unauthorized("User not found")
	}
	systemPrompt := BuildAgentSystemPrompt(s.now().In(s.loc), user.Language)

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	conversationID := req.ConversationID
	planVersion := 0
	history := make([]agent.Message, 0, s.cfg.HistoryMessages)
	if conversationID == 0 {
		session := &model.GuideSession{UserID: userID, Title: truncateRunes(message, 60)}
		if err := s.guide.InsertSessionTx(ctx, tx, session); err != nil {
			return nil, err
		}
		conversationID = session.ID
	} else {
		session, err := s.guide.FindSessionForUpdate(ctx, tx, conversationID, userID)
		if err != nil {
			return nil, err
		}
		if session == nil {
			return nil, common.NotFound("Conversation not found")
		}
		planVersion = session.PlanVersion
		msgs, err := s.guide.FindRecentMessagesTx(ctx, tx, conversationID, s.cfg.HistoryMessages)
		if err != nil {
			return nil, err
		}
		history = toAgentHistory(msgs)
	}

	active, err := s.runs.FindActiveRunTx(ctx, tx, conversationID)
	if err != nil {
		return nil, err
	}
	if active != nil {
		return nil, common.Conflict("当前会话已有正在生成的回复，请稍候")
	}

	var requestID *string
	if clientRequestID != "" {
		requestID = &clientRequestID
	}
	run := &model.AgentRun{
		ConversationID:  conversationID,
		UserID:          userID,
		ClientRequestID: requestID,
		Status:          model.RunQueued,
		Model:           s.cfg.Model,
		Input:           message,
	}
	if err := s.runs.CreateRunTx(ctx, tx, run); err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return nil, common.Conflict("clientRequestId was already used")
		}
		return nil, err
	}
	runID := run.ID

	userMsg := &model.GuideMessage{
		SessionID: conversationID, RunID: &runID, Role: "user",
		Content: message, Status: model.MessageDone,
	}
	if err := s.guide.InsertMessageTx(ctx, tx, userMsg); err != nil {
		return nil, err
	}
	placeholder := &model.GuideMessage{
		SessionID: conversationID, RunID: &runID, Role: "assistant",
		Content: "", Status: model.MessageStreaming, Model: s.cfg.Model,
	}
	if err := s.guide.InsertMessageTx(ctx, tx, placeholder); err != nil {
		return nil, err
	}
	if err := s.guide.TouchSessionTx(ctx, tx, conversationID); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	var runCtx context.Context
	var cancel context.CancelFunc
	if s.cfg.RunTimeout > 0 {
		runCtx, cancel = context.WithTimeout(context.Background(), s.cfg.RunTimeout)
	} else {
		runCtx, cancel = context.WithCancel(context.Background())
	}
	ctl := &runControl{
		runID: runID, conversationID: conversationID, userID: userID,
		assistantMsgID: placeholder.ID, ctx: runCtx, cancel: cancel,
		status: model.RunQueued,
	}

	s.mu.Lock()
	if s.stopped {
		s.mu.Unlock()
		cancel()
		s.finalize(ctl, nil, common.ServiceUnavailable("server is shutting down"))
		return nil, common.ServiceUnavailable("server is shutting down")
	}
	s.active[runID] = ctl
	s.mu.Unlock()

	runReq := agent.RunRequest{
		RunID:          runID,
		ConversationID: conversationID,
		UserID:         userID,
		Language:       user.Language,
		PlanVersion:    planVersion,
		Model:          s.cfg.Model,
		SystemPrompt:   systemPrompt,
		History:        history,
		UserMessage:    message,
		State:          agent.NewRunState(),
		Emit: func(t agent.EventType, data any) {
			s.publish(runID, t, data)
		},
		OnStep: func(rec agent.StepRecord) {
			s.recordStep(runID, rec)
		},
	}
	err = s.executor.Submit(runCtx, strconv.FormatInt(runID, 10), func(taskCtx context.Context) {
		s.execute(taskCtx, ctl, runReq)
	}, func() {
		s.finalize(ctl, nil, errQueueTimeout)
	})
	if err != nil {
		if errors.Is(err, agent.ErrQueueFull) {
			s.finalize(ctl, nil, agent.ErrQueueFull)
			return nil, common.TooManyRequests("助手繁忙，等待队列已满，请稍后重试")
		}
		s.finalize(ctl, nil, err)
		return nil, err
	}
	return s.GetSnapshot(ctx, userID, runID, false)
}

func (s *AgentRunService) execute(ctx context.Context, ctl *runControl, req agent.RunRequest) {
	s.mu.Lock()
	if ctl.finished {
		s.mu.Unlock()
		return
	}
	ctl.status = model.RunRunning
	ctl.startedAt = s.now()
	s.mu.Unlock()

	markCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	_ = s.runs.MarkRunning(markCtx, ctl.runID)
	cancel()

	result, err := s.runner.Run(ctx, req)
	s.finalize(ctl, result, err)
}

func (s *AgentRunService) finalize(ctl *runControl, result *agent.RunResult, runErr error) {
	s.mu.Lock()
	if ctl.finished {
		s.mu.Unlock()
		return
	}
	ctl.finished = true
	knownSteps := ctl.stepCount
	planJSON := ctl.planJSON
	planVersion := ctl.planVersion
	timedOut := ctl.timedOut
	toolCalls := toolCallsJSON(ctl.toolCalls)
	s.mu.Unlock()

	var (
		output    string
		usage     agent.Usage
		draft     *agent.PlanDraft
		stepCount int
	)
	if result != nil {
		output = result.Output
		usage = result.Usage
		draft = result.Draft
		stepCount = result.StepCount
		if stepCount < knownSteps {
			stepCount = knownSteps
		}
	} else {
		stepCount = knownSteps
		usage = ctl.usage
	}
	if len(output) > s.cfg.MaxOutputBytes {
		output = truncateUTF8(output, s.cfg.MaxOutputBytes) + "\n...[truncated]"
	}

	status := model.RunSucceeded
	errorCode, errorMessage := "", ""
	switch {
	case runErr == nil:
		// succeeded
	case errors.Is(runErr, context.Canceled) && timedOut:
		status = model.RunFailed
		errorCode, errorMessage = "run_timeout", "生成超时，请重试"
	case errors.Is(runErr, context.Canceled):
		status = model.RunCanceled
		errorCode, errorMessage = "canceled", "已取消"
	case errors.Is(runErr, context.DeadlineExceeded):
		status = model.RunFailed
		errorCode, errorMessage = "run_timeout", "生成超时，请重试"
	case errors.Is(runErr, errQueueTimeout):
		status = model.RunFailed
		errorCode, errorMessage = "queue_timeout", "排队超时，请稍后重试"
	case errors.Is(runErr, agent.ErrQueueFull):
		status = model.RunFailed
		errorCode, errorMessage = "queue_full", "助手繁忙，等待队列已满"
	case errors.Is(runErr, agent.ErrMaxSteps):
		if draft != nil {
			status = model.RunSucceeded
			output = strings.TrimSpace(output + "\n\n（已达到最大规划步数，已保留当前草稿）")
		} else {
			status = model.RunFailed
			errorCode, errorMessage = "max_steps", "规划步数超限，请缩小范围后重试"
		}
	default:
		status = model.RunFailed
		errorCode, errorMessage = "llm_error", runErr.Error()
		var apiErr *common.ApiError
		if errors.As(runErr, &apiErr) {
			errorMessage = apiErr.Message
			if apiErr.Status == 429 {
				errorCode = "rate_limited"
			}
		}
	}
	if status == model.RunSucceeded && strings.TrimSpace(output) == "" && draft == nil {
		status = model.RunFailed
		errorCode, errorMessage = "empty_output", "模型未返回内容，请重试"
	}

	if draft != nil {
		planJSON = draft.Plan
		planVersion = draft.Version
	}

	dbCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	finishErr := s.runs.FinishRun(dbCtx, repo.FinishRunParams{
		ID:               ctl.runID,
		Status:           status,
		Output:           output,
		ErrorCode:        errorCode,
		ErrorMessage:     errorMessage,
		StepCount:        stepCount,
		PromptTokens:     usage.PromptTokens,
		CompletionTokens: usage.CompletionTokens,
		PlanJSON:         planJSON,
		PlanVersion:      planVersion,
	})
	if finishErr != nil {
		log.Printf("agent: finish run %d: %v", ctl.runID, finishErr)
	}
	msgStatus := model.MessageDone
	switch status {
	case model.RunFailed:
		msgStatus = model.MessageFailed
	case model.RunCanceled:
		msgStatus = model.MessageCanceled
	}
	if err := s.guide.FinishMessage(dbCtx, ctl.assistantMsgID, output, msgStatus, toolCalls, usage.PromptTokens, usage.CompletionTokens); err != nil {
		log.Printf("agent: finish message %d: %v", ctl.assistantMsgID, err)
	}
	_ = s.guide.TouchSession(dbCtx, ctl.conversationID)

	if status == model.RunSucceeded {
		s.publish(ctl.runID, agent.EventDone, agent.DonePayload{
			Status: status, StepCount: stepCount, Usage: usage, Plan: len(planJSON) > 0,
		})
	} else {
		s.publish(ctl.runID, agent.EventError, agent.ErrorPayload{Code: errorCode, Message: errorMessage})
	}

	s.closeSubscribers(ctl.runID)
	s.mu.Lock()
	delete(s.active, ctl.runID)
	s.mu.Unlock()
	ctl.cancel()
}

func (s *AgentRunService) publish(runID int64, typ agent.EventType, data any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ctl, ok := s.active[runID]
	if !ok {
		return
	}
	ctl.seq++
	event := agent.Event{Type: typ, Seq: ctl.seq, RunID: runID, Data: data}

	switch typ {
	case agent.EventDelta:
		if payload, ok := data.(agent.DeltaPayload); ok && ctl.output.Len() < s.cfg.MaxOutputBytes {
			ctl.output.WriteString(payload.Content)
			ctl.dirty = true
		}
	case agent.EventPlan:
		if payload, ok := data.(agent.PlanPayload); ok {
			ctl.planJSON = payload.Plan
			ctl.planVersion = payload.Version
		}
	case agent.EventToolCall:
		if payload, ok := data.(agent.ToolCallPayload); ok && len(ctl.toolCalls) < 20 {
			ctl.toolCalls = append(ctl.toolCalls, map[string]string{
				"name": payload.Name, "arguments": truncateRunes(payload.Arguments, 500),
			})
		}
	}

	for id, sub := range s.subscribers[runID] {
		select {
		case sub.ch <- event:
		default:
			delete(s.subscribers[runID], id)
			close(sub.ch)
		}
	}
}

func (s *AgentRunService) closeSubscribers(runID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, sub := range s.subscribers[runID] {
		close(sub.ch)
		delete(s.subscribers[runID], id)
	}
	delete(s.subscribers, runID)
}

func (s *AgentRunService) removeSubscriber(runID, id int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if subs, ok := s.subscribers[runID]; ok {
		if sub, ok := subs[id]; ok {
			close(sub.ch)
			delete(subs, id)
		}
		if len(subs) == 0 {
			delete(s.subscribers, runID)
		}
	}
}

func (s *AgentRunService) Subscribe(ctx context.Context, userID, runID int64) (<-chan agent.Event, func(), *AgentSnapshot, error) {
	run, err := s.runs.FindRun(ctx, runID, userID)
	if err != nil {
		return nil, nil, nil, err
	}
	if run == nil {
		return nil, nil, nil, common.NotFound("Run not found")
	}

	s.mu.Lock()
	ctl := s.active[runID]
	snapshot := s.snapshotFrom(run, ctl, nil)
	if ctl == nil || ctl.finished {
		ch := make(chan agent.Event)
		close(ch)
		s.mu.Unlock()
		return ch, func() {}, snapshot, nil
	}
	sub := &subscriber{ch: make(chan agent.Event, 128)}
	s.nextSubID++
	id := s.nextSubID
	if s.subscribers[runID] == nil {
		s.subscribers[runID] = make(map[int64]*subscriber)
	}
	s.subscribers[runID][id] = sub
	s.mu.Unlock()

	cancel := func() { s.removeSubscriber(runID, id) }
	return sub.ch, cancel, snapshot, nil
}

func (s *AgentRunService) GetSnapshot(ctx context.Context, userID, runID int64, includeSteps bool) (*AgentSnapshot, error) {
	run, err := s.runs.FindRun(ctx, runID, userID)
	if err != nil {
		return nil, err
	}
	if run == nil {
		return nil, common.NotFound("Run not found")
	}
	var steps []model.AgentStep
	if includeSteps {
		steps, err = s.runs.ListSteps(ctx, runID, 50)
		if err != nil {
			return nil, err
		}
	}
	s.mu.Lock()
	ctl := s.active[runID]
	snapshot := s.snapshotFrom(run, ctl, steps)
	s.mu.Unlock()
	return snapshot, nil
}

func (s *AgentRunService) snapshotFrom(run *model.AgentRun, ctl *runControl, steps []model.AgentStep) *AgentSnapshot {
	snapshot := &AgentSnapshot{
		RunID:          run.ID,
		ConversationID: run.ConversationID,
		Status:         run.Status,
		Output:         run.Output,
		Plan:           json.RawMessage(run.PlanJSON),
		PlanVersion:    run.PlanVersion,
		StepCount:      run.StepCount,
		Usage: agent.Usage{
			PromptTokens:     run.PromptTokens,
			CompletionTokens: run.CompletionTokens,
			TotalTokens:      run.PromptTokens + run.CompletionTokens,
		},
		CreatedAt:  run.CreatedAt,
		StartedAt:  run.StartedAt,
		FinishedAt: run.FinishedAt,
		Steps:      steps,
	}
	if run.ErrorCode != nil || run.ErrorMessage != nil {
		snapshot.Error = &AgentError{}
		if run.ErrorCode != nil {
			snapshot.Error.Code = *run.ErrorCode
		}
		if run.ErrorMessage != nil {
			snapshot.Error.Message = *run.ErrorMessage
		}
	}
	if ctl != nil {
		if ctl.output.Len() > 0 {
			snapshot.Output = ctl.output.String()
		}
		if len(ctl.planJSON) > 0 {
			snapshot.Plan = ctl.planJSON
			snapshot.PlanVersion = ctl.planVersion
		}
		snapshot.Status = ctl.status
		snapshot.Seq = ctl.seq
		if ctl.stepCount > snapshot.StepCount {
			snapshot.StepCount = ctl.stepCount
		}
	}
	return snapshot
}

func (s *AgentRunService) Cancel(ctx context.Context, userID, runID int64) (*AgentSnapshot, error) {
	run, err := s.runs.FindRun(ctx, runID, userID)
	if err != nil {
		return nil, err
	}
	if run == nil {
		return nil, common.NotFound("Run not found")
	}
	if isTerminal(run.Status) {
		return s.GetSnapshot(ctx, userID, runID, false)
	}

	s.mu.Lock()
	ctl := s.active[runID]
	if ctl == nil {
		s.mu.Unlock()
		return s.GetSnapshot(ctx, userID, runID, false)
	}
	ctl.status = model.RunCanceling
	cancel := ctl.cancel
	s.mu.Unlock()

	cancel()
	if s.executor.Remove(strconv.FormatInt(runID, 10)) {
		s.finalize(ctl, nil, context.Canceled)
	}
	return s.GetSnapshot(ctx, userID, runID, false)
}

func (s *AgentRunService) ListConversations(ctx context.Context, userID int64, page, pageSize int) (*ConversationPage, error) {
	page, pageSize, offset := normalizePage(page, pageSize)
	total, err := s.guide.CountSessions(ctx, userID)
	if err != nil {
		return nil, err
	}
	items, err := s.guide.ListSessions(ctx, userID, pageSize, offset)
	if err != nil {
		return nil, err
	}
	return &ConversationPage{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *AgentRunService) Shutdown(ctx context.Context) {
	s.mu.Lock()
	s.stopped = true
	type shutdownRun struct {
		ctl    *runControl
		status string
	}
	runs := make([]shutdownRun, 0, len(s.active))
	for _, ctl := range s.active {
		runs = append(runs, shutdownRun{ctl: ctl, status: ctl.status})
	}
	s.mu.Unlock()
	for _, run := range runs {
		run.ctl.cancel()
		// A queued run never reaches execute(), so finalize it here instead of
		// waiting for the executor goroutine to drain.
		if run.status == model.RunQueued {
			s.finalize(run.ctl, nil, context.Canceled)
		}
	}

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	for {
		s.mu.Lock()
		n := len(s.active)
		s.mu.Unlock()
		if n == 0 {
			break
		}
		select {
		case <-ctx.Done():
			goto drained
		case <-ticker.C:
		}
	}
drained:
	s.executor.Shutdown(ctx)
	s.stopOnce.Do(func() {
		close(s.stopCh)
	})
	s.mu.Lock()
	for runID := range s.subscribers {
		for id, sub := range s.subscribers[runID] {
			close(sub.ch)
			delete(s.subscribers[runID], id)
		}
	}
	s.subscribers = make(map[int64]map[int64]*subscriber)
	s.mu.Unlock()
}

func (s *AgentRunService) Metrics() AgentMetrics {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	s.mu.Lock()
	active := len(s.active)
	s.mu.Unlock()
	return AgentMetrics{
		Goroutines:      runtime.NumGoroutine(),
		HeapAlloc:       mem.HeapAlloc,
		NumGC:           mem.NumGC,
		ActiveRuns:      active,
		ExecutorMode:    s.executor.Mode(),
		ExecutorRunning: s.executor.RunningCount(),
		ExecutorQueued:  s.executor.QueuedCount(),
	}
}

func (s *AgentRunService) tickLoop() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.tick()
		}
	}
}

func (s *AgentRunService) tick() {
	now := s.now()
	type flush struct {
		id     int64
		output string
		length int
	}
	flushes := make([]flush, 0, 4)

	s.mu.Lock()
	for _, ctl := range s.active {
		if ctl.status == model.RunRunning && !ctl.startedAt.IsZero() && s.cfg.RunTimeout > 0 &&
			now.Sub(ctl.startedAt) > s.cfg.RunTimeout {
			ctl.status = model.RunCanceling
			ctl.timedOut = true
			ctl.cancel()
			continue
		}
		if ctl.dirty && !ctl.flushInFlight {
			ctl.flushInFlight = true
			flushes = append(flushes, flush{id: ctl.runID, output: ctl.output.String(), length: ctl.output.Len()})
		}
	}
	if len(s.rate) > 1024 {
		cutoff := now.Add(-time.Minute)
		for id, list := range s.rate {
			if len(list) == 0 || list[len(list)-1].Before(cutoff) {
				delete(s.rate, id)
			}
		}
	}
	s.mu.Unlock()

	for _, f := range flushes {
		s.scheduleFlush(f.id, f.output, f.length)
	}
}

func (s *AgentRunService) scheduleFlush(runID int64, output string, length int) {
	select {
	case s.flushSem <- struct{}{}:
		go func() {
			defer func() { <-s.flushSem }()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			_ = s.runs.UpdateOutput(ctx, runID, output)
			cancel()
			s.mu.Lock()
			if ctl, ok := s.active[runID]; ok {
				if ctl.output.Len() == length {
					ctl.dirty = false
				}
				ctl.flushInFlight = false
			}
			s.mu.Unlock()
		}()
	default:
		s.mu.Lock()
		if ctl, ok := s.active[runID]; ok {
			ctl.flushInFlight = false
		}
		s.mu.Unlock()
	}
}

func (s *AgentRunService) recordStep(runID int64, rec agent.StepRecord) {
	step := model.AgentStep{
		RunID:     runID,
		StepNo:    rec.StepNo,
		Type:      rec.Type,
		Name:      rec.Name,
		Arguments: truncateRunes(rec.Arguments, 4000),
		Result:    truncateRunes(rec.Result, 4000),
		LatencyMS: int(rec.Latency / time.Millisecond),
	}
	if rec.Error != "" {
		errText := truncateRunes(rec.Error, 500)
		step.Error = &errText
	}
	select {
	case s.steps <- step:
	default:
		// Trace is best-effort; never block the model loop.
	}
}

func (s *AgentRunService) stepWriter() {
	for {
		select {
		case <-s.stopCh:
			return
		case step := <-s.steps:
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			if err := s.runs.AppendStep(ctx, &step); err != nil {
				log.Printf("agent: append step: %v", err)
			}
			cancel()
		}
	}
}

func (s *AgentRunService) allowRate(userID int64) bool {
	now := s.now()
	cutoff := now.Add(-time.Minute)
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.rate[userID]
	kept := list[:0]
	for _, t := range list {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	if len(kept) >= s.cfg.RateLimitPerMinute {
		s.rate[userID] = kept
		return false
	}
	s.rate[userID] = append(kept, now)
	return true
}

func toolCallsJSON(calls []map[string]string) json.RawMessage {
	if len(calls) == 0 {
		return nil
	}
	raw, err := json.Marshal(calls)
	if err != nil {
		return nil
	}
	return raw
}

func toAgentHistory(messages []model.GuideMessage) []agent.Message {
	history := make([]agent.Message, 0, len(messages))
	for _, m := range messages {
		if m.Status != model.MessageDone {
			continue
		}
		if m.Role != "user" && m.Role != "assistant" {
			continue
		}
		content := strings.TrimSpace(m.Content)
		if content == "" {
			continue
		}
		history = append(history, agent.Message{
			Role:    agent.Role(m.Role),
			Content: truncateRunes(content, 4000),
		})
	}
	return history
}

func isTerminal(status string) bool {
	switch status {
	case model.RunSucceeded, model.RunFailed, model.RunCanceled, model.RunInterrupted:
		return true
	default:
		return false
	}
}

func truncateUTF8(s string, maxBytes int) string {
	if len(s) <= maxBytes {
		return s
	}
	cut := maxBytes
	for cut > 0 && (s[cut]&0xC0) == 0x80 {
		cut--
	}
	return s[:cut]
}
