package repo

import (
	"context"
	"database/sql"

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

func (r *GuideRepo) FindSession(ctx context.Context, id, userID int64) (*model.GuideSession, error) {
	var s model.GuideSession
	err := r.db.GetContext(ctx, &s,
		"SELECT * FROM guide_sessions WHERE id = ? AND user_id = ?", id, userID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &s, err
}

func (r *GuideRepo) InsertMessage(ctx context.Context, m *model.GuideMessage) error {
	res, err := r.db.ExecContext(ctx,
		"INSERT INTO guide_messages (session_id, role, content) VALUES (?, ?, ?)",
		m.SessionID, m.Role, m.Content)
	if err != nil {
		return err
	}
	m.ID, err = res.LastInsertId()
	return err
}

func (r *GuideRepo) FindMessages(ctx context.Context, sessionID int64) ([]model.GuideMessage, error) {
	var list []model.GuideMessage
	err := r.db.SelectContext(ctx, &list,
		"SELECT * FROM guide_messages WHERE session_id = ? ORDER BY id ASC", sessionID)
	return list, err
}
