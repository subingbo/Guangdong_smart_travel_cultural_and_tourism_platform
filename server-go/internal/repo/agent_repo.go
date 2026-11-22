package repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	"github.com/travelgd/server-go/internal/model"
)

type AgentRepo struct{ db *sqlx.DB }

func NewAgentRepo(db *sqlx.DB) *AgentRepo { return &AgentRepo{db: db} }

func (r *AgentRepo) CreateRun(ctx context.Context, run *model.AgentRun) error {
	var clientRequestID interface{}
	if run.ClientRequestID != nil && *run.ClientRequestID != "" {
		clientRequestID = *run.ClientRequestID
	}
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO ai_runs (conversation_id, user_id, client_request_id, status, model, input, output)
		VALUES (?, ?, ?, ?, ?, ?, '')`,
		run.ConversationID, run.UserID, clientRequestID, run.Status, run.Model, run.Input)
	if err != nil {
		return err
	}
	run.ID, err = res.LastInsertId()
	return err
}

func (r *AgentRepo) CreateRunTx(ctx context.Context, tx *sqlx.Tx, run *model.AgentRun) error {
	var clientRequestID interface{}
	if run.ClientRequestID != nil && *run.ClientRequestID != "" {
		clientRequestID = *run.ClientRequestID
	}
	res, err := tx.ExecContext(ctx, `
		INSERT INTO ai_runs (conversation_id, user_id, client_request_id, status, model, input, output)
		VALUES (?, ?, ?, ?, ?, ?, '')`,
		run.ConversationID, run.UserID, clientRequestID, run.Status, run.Model, run.Input)
	if err != nil {
		return err
	}
	run.ID, err = res.LastInsertId()
	return err
}

func (r *AgentRepo) FindActiveRunTx(ctx context.Context, tx *sqlx.Tx, conversationID int64) (*model.AgentRun, error) {
	var run model.AgentRun
	err := tx.GetContext(ctx, &run, `
		SELECT * FROM ai_runs
		WHERE conversation_id = ? AND status IN ('queued', 'running', 'canceling')
		ORDER BY id DESC LIMIT 1`, conversationID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &run, err
}

func (r *AgentRepo) FindRun(ctx context.Context, id, userID int64) (*model.AgentRun, error) {
	var run model.AgentRun
	err := r.db.GetContext(ctx, &run, "SELECT * FROM ai_runs WHERE id = ? AND user_id = ?", id, userID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &run, err
}

func (r *AgentRepo) FindRunByClientRequestID(ctx context.Context, userID int64, clientRequestID string) (*model.AgentRun, error) {
	var run model.AgentRun
	err := r.db.GetContext(ctx, &run,
		"SELECT * FROM ai_runs WHERE user_id = ? AND client_request_id = ?", userID, clientRequestID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &run, err
}

func (r *AgentRepo) FindActiveRun(ctx context.Context, conversationID int64) (*model.AgentRun, error) {
	var run model.AgentRun
	err := r.db.GetContext(ctx, &run, `
		SELECT * FROM ai_runs
		WHERE conversation_id = ? AND status IN ('queued', 'running', 'canceling')
		ORDER BY id DESC LIMIT 1`, conversationID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &run, err
}

func (r *AgentRepo) MarkRunning(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE ai_runs SET status = 'running', started_at = NOW()
		WHERE id = ? AND status = 'queued'`, id)
	return err
}

func (r *AgentRepo) UpdateOutput(ctx context.Context, id int64, output string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE ai_runs SET output = ?
		WHERE id = ? AND status IN ('running', 'canceling')`, output, id)
	return err
}

type FinishRunParams struct {
	ID               int64
	Status           string
	Output           string
	ErrorCode        string
	ErrorMessage     string
	StepCount        int
	PromptTokens     int
	CompletionTokens int
	PlanJSON         json.RawMessage
	PlanVersion      int
}

func (r *AgentRepo) FinishRun(ctx context.Context, p FinishRunParams) error {
	var plan interface{}
	if len(p.PlanJSON) > 0 {
		plan = []byte(p.PlanJSON)
	}
	_, err := r.db.ExecContext(ctx, `
		UPDATE ai_runs SET
			status = ?,
			output = ?,
			error_code = NULLIF(?, ''),
			error_message = NULLIF(?, ''),
			step_count = ?,
			prompt_tokens = ?,
			completion_tokens = ?,
			plan_json = COALESCE(?, plan_json),
			plan_version = IF(? > 0, ?, plan_version),
			finished_at = NOW()
		WHERE id = ?`,
		p.Status, p.Output, p.ErrorCode, p.ErrorMessage, p.StepCount,
		p.PromptTokens, p.CompletionTokens, plan, p.PlanVersion, p.PlanVersion, p.ID)
	return err
}

func (r *AgentRepo) AppendStep(ctx context.Context, step *model.AgentStep) error {
	var errValue interface{}
	if step.Error != nil && *step.Error != "" {
		errValue = *step.Error
	}
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO ai_steps (run_id, step_no, type, name, arguments, result, latency_ms, error)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		step.RunID, step.StepNo, step.Type, step.Name, step.Arguments, step.Result, step.LatencyMS, errValue)
	if err != nil {
		return err
	}
	step.ID, err = res.LastInsertId()
	return err
}

func (r *AgentRepo) ListSteps(ctx context.Context, runID int64, limit int) ([]model.AgentStep, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	list := make([]model.AgentStep, 0)
	err := r.db.SelectContext(ctx, &list, `
		SELECT * FROM ai_steps WHERE run_id = ?
		ORDER BY step_no ASC, id ASC LIMIT ?`, runID, limit)
	return list, err
}

// SaveDraft stores a server-built plan on both the run and the conversation.
func (r *AgentRepo) SaveDraft(ctx context.Context, runID, conversationID int64, planJSON []byte, version int) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
		UPDATE ai_runs SET plan_json = ?, plan_version = ? WHERE id = ?`,
		planJSON, version, runID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE guide_sessions
		SET plan_json = ?, plan_version = ?, last_run_id = ?, updated_at = NOW()
		WHERE id = ?`, planJSON, version, runID, conversationID); err != nil {
		return err
	}
	return tx.Commit()
}

type ConfirmPlanParams struct {
	UserID         int64
	RunID          int64
	ConversationID int64
	IdempotencyKey string
	PlanJSON       []byte
	PlanVersion    int
	Route          *model.Route
}

// ConfirmPlan writes a route, the idempotency record and the run/session
// pointers in one transaction. A duplicate idempotency key returns the
// previously created route instead of creating a second one.
func (r *AgentRepo) ConfirmPlan(ctx context.Context, p ConfirmPlanParams) (int64, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var existing int64
	err = tx.GetContext(ctx, &existing, `
		SELECT route_id FROM ai_plan_commits
		WHERE user_id = ? AND idempotency_key = ?`, p.UserID, p.IdempotencyKey)
	if err == nil {
		return existing, tx.Commit()
	}
	if err != sql.ErrNoRows {
		return 0, err
	}

	res, err := tx.ExecContext(ctx, `
		INSERT INTO routes (user_id, route_name, duration_days, total_distance, mode, pace, interests, waypoints, polyline)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.Route.UserID, p.Route.RouteName, p.Route.DurationDays, p.Route.TotalDistance,
		p.Route.Mode, p.Route.Pace, p.Route.Interests, p.Route.Waypoints, p.Route.Polyline)
	if err != nil {
		return 0, err
	}
	routeID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO ai_plan_commits (user_id, run_id, idempotency_key, route_id)
		VALUES (?, ?, ?, ?)`, p.UserID, p.RunID, p.IdempotencyKey, routeID); err != nil {
		if isDuplicateKey(err) {
			tx.Rollback()
			return r.findCommitRoute(ctx, p.UserID, p.IdempotencyKey)
		}
		return 0, err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE ai_runs SET route_id = ?, plan_json = ?, plan_version = ?
		WHERE id = ? AND user_id = ?`,
		routeID, p.PlanJSON, p.PlanVersion, p.RunID, p.UserID); err != nil {
		return 0, err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE guide_sessions
		SET plan_json = ?, plan_version = ?, last_run_id = ?, updated_at = NOW()
		WHERE id = ?`, p.PlanJSON, p.PlanVersion, p.RunID, p.ConversationID); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	p.Route.ID = routeID
	return routeID, nil
}

func (r *AgentRepo) findCommitRoute(ctx context.Context, userID int64, key string) (int64, error) {
	var routeID int64
	err := r.db.GetContext(ctx, &routeID, `
		SELECT route_id FROM ai_plan_commits
		WHERE user_id = ? AND idempotency_key = ?`, userID, key)
	return routeID, err
}

// RecoverInterrupted marks runs left behind by a process exit. It never
// replays tools: the client is expected to resend the message.
func (r *AgentRepo) RecoverInterrupted(ctx context.Context) (int64, error) {
	res, err := r.db.ExecContext(ctx, `
		UPDATE ai_runs
		SET status = 'interrupted',
			error_code = 'server_restart',
			error_message = 'server restarted before the run finished',
			finished_at = NOW()
		WHERE status IN ('queued', 'running', 'canceling')`)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	if _, err := r.db.ExecContext(ctx, `
		UPDATE guide_messages SET status = 'interrupted' WHERE status = 'streaming'`); err != nil {
		return n, err
	}
	return n, nil
}

func isDuplicateKey(err error) bool {
	var me *mysql.MySQLError
	if errors.As(err, &me) {
		return me.Number == 1062
	}
	return false
}
