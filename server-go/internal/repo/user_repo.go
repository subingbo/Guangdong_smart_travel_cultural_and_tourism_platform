package repo

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/travelgd/server-go/internal/model"
)

type UserRepo struct{ db *sqlx.DB }

func NewUserRepo(db *sqlx.DB) *UserRepo { return &UserRepo{db: db} }

func (r *UserRepo) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	var u model.User
	err := r.db.GetContext(ctx, &u, "SELECT * FROM users WHERE username = ?", username)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &u, err
}

func (r *UserRepo) FindByID(ctx context.Context, id int64) (*model.User, error) {
	var u model.User
	err := r.db.GetContext(ctx, &u, "SELECT * FROM users WHERE id = ?", id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &u, err
}

func (r *UserRepo) CountByUsername(ctx context.Context, username string) (int, error) {
	var n int
	err := r.db.GetContext(ctx, &n, "SELECT COUNT(*) FROM users WHERE username = ?", username)
	return n, err
}

func (r *UserRepo) Insert(ctx context.Context, u *model.User) error {
	res, err := r.db.ExecContext(ctx,
		"INSERT INTO users (username, password_hash, nickname, avatar_url, language) VALUES (?, ?, ?, ?, ?)",
		u.Username, u.PasswordHash, u.Nickname, u.AvatarURL, u.Language)
	if err != nil {
		return err
	}
	u.ID, err = res.LastInsertId()
	return err
}

func (r *UserRepo) UpdateProfile(ctx context.Context, u *model.User) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE users SET nickname = ?, avatar_url = ?, language = ? WHERE id = ?",
		u.Nickname, u.AvatarURL, u.Language, u.ID)
	return err
}
