package repo

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/jmoiron/sqlx"
	"github.com/travelgd/server-go/internal/model"
)

type GuideRepo struct{ db *sqlx.DB }

func NewGuideRepo(db *sqlx.DB) *GuideRepo { return &GuideRepo{db: db} }

func (r *GuideRepo) InsertSession(ctx context.Context, s *model.GuideSession) error {
	res, err := r.db.ExecContext(ctx,
		"INSERT INTO guide_sessions (user_id, title) VALUES (?, ?)", s.UserID, s.Title)
	if err != nil {
		return err
	}
	s.ID, err = res.LastInsertId()
	return err
}

func (r *GuideRepo) InsertSessionTx(ctx context.Context, tx *sqlx.Tx, s *model.GuideSession) error {
	res, err := tx.ExecContext(ctx,
		"INSERT INTO guide_sessions (user_id, title) VALUES (?, ?)", s.UserID, s.Title)
	if err != nil {
		return err
	}
	s.ID, err = res.LastInsertId()
	return err
}

func (r *GuideRepo) FindSession(ctx context.Context, id, userID int64) (*model.GuideSession, error) {
	var s model.GuideSession
	err := r.db.GetContext(ctx, &s,
		"SELECT * FROM guide_sessions WHERE id = ? AND user_id = ?", id, userID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &s, err
}

// FindSessionForUpdate locks the conversation row so a user cannot start two
// active runs in the same conversation.
func (r *GuideRepo) FindSessionForUpdate(ctx context.Context, tx *sqlx.Tx, id, userID int64) (*model.GuideSession, error) {
	var s model.GuideSession
	err := tx.GetContext(ctx, &s,
		"SELECT * FROM guide_sessions WHERE id = ? AND user_id = ? FOR UPDATE", id, userID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &s, err
}

func (r *GuideRepo) InsertMessage(ctx context.Context, m *model.GuideMessage) error {
	status := m.Status
	if status == "" {
		status = model.MessageDone
	}
	var runID interface{}
	if m.RunID != nil {
		runID = *m.RunID
	}
	var toolCalls interface{}
	if len(m.ToolCalls) > 0 {
		toolCalls = []byte(m.ToolCalls)
	}
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO guide_messages
			(session_id, run_id, role, content, tool_calls, status, model, prompt_tokens, completion_tokens)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		m.SessionID, runID, m.Role, m.Content, toolCalls, status, m.Model, m.PromptTokens, m.CompletionTokens)
	if err != nil {
		return err
	}
	m.ID, err = res.LastInsertId()
	return err
}

// InsertMessageTx inserts a message inside an existing transaction.
func (r *GuideRepo) InsertMessageTx(ctx context.Context, tx *sqlx.Tx, m *model.GuideMessage) error {
	status := m.Status
	if status == "" {
		status = model.MessageDone
	}
	var runID interface{}
	if m.RunID != nil {
		runID = *m.RunID
	}
	var toolCalls interface{}
	if len(m.ToolCalls) > 0 {
		toolCalls = []byte(m.ToolCalls)
	}
	res, err := tx.ExecContext(ctx, `
		INSERT INTO guide_messages
			(session_id, run_id, role, content, tool_calls, status, model, prompt_tokens, completion_tokens)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		m.SessionID, runID, m.Role, m.Content, toolCalls, status, m.Model, m.PromptTokens, m.CompletionTokens)
	if err != nil {
		return err
	}
	m.ID, err = res.LastInsertId()
	return err
}

func (r *GuideRepo) FinishMessage(ctx context.Context, id int64, content, status string, toolCalls json.RawMessage, promptTokens, completionTokens int) error {
	var toolCallsValue interface{}
	if len(toolCalls) > 0 {
		toolCallsValue = []byte(toolCalls)
	}
	_, err := r.db.ExecContext(ctx, `
		UPDATE guide_messages
		SET content = ?, status = ?, tool_calls = ?, prompt_tokens = ?, completion_tokens = ?
		WHERE id = ?`,
		content, status, toolCallsValue, promptTokens, completionTokens, id)
	return err
}

func (r *GuideRepo) UpdateSessionPlan(ctx context.Context, sessionID int64, planJSON json.RawMessage, version int, lastRunID int64) error {
	var plan interface{}
	if len(planJSON) > 0 {
		plan = []byte(planJSON)
	}
	_, err := r.db.ExecContext(ctx, `
		UPDATE guide_sessions
		SET plan_json = ?, plan_version = ?, last_run_id = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`, plan, version, lastRunID, sessionID)
	return err
}

func (r *GuideRepo) TouchSession(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE guide_sessions SET updated_at = CURRENT_TIMESTAMP WHERE id = ?", id)
	return err
}

func (r *GuideRepo) TouchSessionTx(ctx context.Context, tx *sqlx.Tx, id int64) error {
	_, err := tx.ExecContext(ctx,
		"UPDATE guide_sessions SET updated_at = CURRENT_TIMESTAMP WHERE id = ?", id)
	return err
}

func (r *GuideRepo) FindMessages(ctx context.Context, sessionID int64) ([]model.GuideMessage, error) {
	list := make([]model.GuideMessage, 0)
	err := r.db.SelectContext(ctx, &list,
		"SELECT * FROM guide_messages WHERE session_id = ? ORDER BY id ASC", sessionID)
	return list, err
}

// FindRecentMessages returns the newest limit messages in ascending order.
func (r *GuideRepo) FindRecentMessages(ctx context.Context, sessionID int64, limit int) ([]model.GuideMessage, error) {
	if limit <= 0 {
		limit = 12
	}
	list := make([]model.GuideMessage, 0)
	err := r.db.SelectContext(ctx, &list,
		"SELECT * FROM guide_messages WHERE session_id = ? ORDER BY id DESC LIMIT ?", sessionID, limit)
	if err != nil {
		return nil, err
	}
	for i, j := 0, len(list)-1; i < j; i, j = i+1, j-1 {
		list[i], list[j] = list[j], list[i]
	}
	return list, nil
}

func (r *GuideRepo) FindRecentMessagesTx(ctx context.Context, tx *sqlx.Tx, sessionID int64, limit int) ([]model.GuideMessage, error) {
	if limit <= 0 {
		limit = 12
	}
	list := make([]model.GuideMessage, 0)
	err := tx.SelectContext(ctx, &list,
		"SELECT * FROM guide_messages WHERE session_id = ? ORDER BY id DESC LIMIT ?", sessionID, limit)
	if err != nil {
		return nil, err
	}
	for i, j := 0, len(list)-1; i < j; i, j = i+1, j-1 {
		list[i], list[j] = list[j], list[i]
	}
	return list, nil
}

func (r *GuideRepo) ListSessions(ctx context.Context, userID int64, limit, offset int) ([]model.GuideSession, error) {
	list := make([]model.GuideSession, 0)
	err := r.db.SelectContext(ctx, &list, `
		SELECT * FROM guide_sessions
		WHERE user_id = ?
		ORDER BY updated_at DESC, id DESC
		LIMIT ? OFFSET ?`, userID, limit, offset)
	return list, err
}

func (r *GuideRepo) CountSessions(ctx context.Context, userID int64) (int, error) {
	var n int
	err := r.db.GetContext(ctx, &n,
		"SELECT COUNT(*) FROM guide_sessions WHERE user_id = ?", userID)
	return n, err
}
