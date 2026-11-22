package model

import (
	"time"
)

// Run status values mirror the Java assistant-run state machine.
const (
	RunQueued      = "queued"
	RunRunning     = "running"
	RunCanceling   = "canceling"
	RunSucceeded   = "succeeded"
	RunFailed      = "failed"
	RunCanceled    = "canceled"
	RunInterrupted = "interrupted"
)

// Message status values used by guide_messages placeholders.
const (
	MessageStreaming   = "streaming"
	MessageDone        = "done"
	MessageFailed      = "failed"
	MessageCanceled    = "canceled"
	MessageInterrupted = "interrupted"
)

type AgentRun struct {
	ID               int64      `db:"id" json:"id"`
	ConversationID   int64      `db:"conversation_id" json:"conversationId"`
	UserID           int64      `db:"user_id" json:"-"`
	ClientRequestID  *string    `db:"client_request_id" json:"clientRequestId,omitempty"`
	Status           string     `db:"status" json:"status"`
	Model            string     `db:"model" json:"model"`
	Input            string     `db:"input" json:"input,omitempty"`
	Output           string     `db:"output" json:"output"`
	ErrorCode        *string    `db:"error_code" json:"errorCode,omitempty"`
	ErrorMessage     *string    `db:"error_message" json:"errorMessage,omitempty"`
	StepCount        int        `db:"step_count" json:"stepCount"`
	PromptTokens     int        `db:"prompt_tokens" json:"promptTokens"`
	CompletionTokens int        `db:"completion_tokens" json:"completionTokens"`
	PlanJSON         JSONText   `db:"plan_json" json:"plan,omitempty"`
	PlanVersion      int        `db:"plan_version" json:"planVersion"`
	RouteID          *int64     `db:"route_id" json:"routeId,omitempty"`
	CreatedAt        time.Time  `db:"created_at" json:"createdAt"`
	StartedAt        *time.Time `db:"started_at" json:"startedAt,omitempty"`
	FinishedAt       *time.Time `db:"finished_at" json:"finishedAt,omitempty"`
	UpdatedAt        time.Time  `db:"updated_at" json:"updatedAt"`
}

type PlanCommit struct {
	ID             int64     `db:"id" json:"id"`
	UserID         int64     `db:"user_id" json:"userId"`
	RunID          int64     `db:"run_id" json:"runId"`
	IdempotencyKey string    `db:"idempotency_key" json:"idempotencyKey"`
	RouteID        int64     `db:"route_id" json:"routeId"`
	CreatedAt      time.Time `db:"created_at" json:"createdAt"`
}

type AgentStep struct {
	ID        int64     `db:"id" json:"id"`
	RunID     int64     `db:"run_id" json:"runId"`
	StepNo    int       `db:"step_no" json:"stepNo"`
	Type      string    `db:"type" json:"type"`
	Name      string    `db:"name" json:"name"`
	Arguments string    `db:"arguments" json:"arguments,omitempty"`
	Result    string    `db:"result" json:"result,omitempty"`
	LatencyMS int       `db:"latency_ms" json:"latencyMs"`
	Error     *string   `db:"error" json:"error,omitempty"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
}
