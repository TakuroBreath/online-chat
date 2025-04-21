package sqlite

import (
	"context"
	"fmt"
	"time"

	"auth.service/internal/repository"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type SqliteUserRepository struct {
	db *sqlx.DB
}

type SqliteSessionRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *SqliteUserRepository {
	return &SqliteUserRepository{
		db: db,
	}
}

func NewSessionRepository(db *sqlx.DB) *SqliteSessionRepository {
	return &SqliteSessionRepository{
		db: db,
	}
}

func (r *SqliteUserRepository) CreateUser(ctx context.Context, user *repository.User) error {
	var op = "repository.SqliteUserRepository.CreateUser"

	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	query := `
		INSERT INTO users (id, username, password_hash, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query, user.ID, user.Username, user.PasswordHash, user.CreatedAt, user.UpdatedAt)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *SqliteUserRepository) UserByID(ctx context.Context, id string) (*repository.User, error) {
	var op = "repository.SqliteUserRepository.UserByID"

	user := &repository.User{}

	query := `
		SELECT id, username, password_hash, created_at, updated_at
		FROM users
		WHERE id = ?
	`

	err := r.db.GetContext(ctx, user, query, id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (r *SqliteUserRepository) UserByUsername(ctx context.Context, username string) (*repository.User, error) {
	var op = "repository.SqliteUserRepository.UserByUsername"

	user := &repository.User{}

	query := `
		SELECT id, username, password_hash, created_at, updated_at
		FROM users
		WHERE username = ?
	`

	err := r.db.GetContext(ctx, user, query, username)
	if err != nil {
		switch {
		case err.Error() == "sql: no rows in result set":
			return nil, repository.ErrUserNotFound
		default:
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	return user, nil
}

func (r *SqliteUserRepository) UpdateUser(ctx context.Context, user *repository.User) error {
	var op = "repository.SqliteUserRepository.UpdateUser"

	user.UpdatedAt = time.Now()

	query := `
		UPDATE users
		SET username = ?, password_hash = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query, user.Username, user.PasswordHash, user.UpdatedAt, user.ID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%s: %s", op, "user not found")
	}

	return nil
}

func (r *SqliteUserRepository) DeleteUser(ctx context.Context, id string) error {
	var op = "repository.SqliteUserRepository.DeleteUser"

	query := `
		DELETE FROM users
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%s: %s", op, "user not found")
	}

	return nil
}

func (r *SqliteSessionRepository) CreateSession(ctx context.Context, session *repository.Session) error {
	var op = "repository.SqliteSessionRepository.CreateSession"

	if session.ID == "" {
		session.ID = uuid.New().String()
	}

	now := time.Now()
	session.CreatedAt = now
	session.ExpiresAt = now.Add(24 * time.Hour)

	query := `
		INSERT INTO sessions (id, user_id, refresh_token, expires_at, created_at, expires_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query, session.ID, session.UserID, session.RefreshToken, session.ExpiresAt, session.CreatedAt, session.ExpiresAt)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *SqliteSessionRepository) GetByRefreshToken(ctx context.Context, refreshToken string) (*repository.Session, error) {
	var op = "repository.SqliteSessionRepository.GetByRefreshToken"

	session := &repository.Session{}

	query := `
		SELECT id, user_id, refresh_token, expires_at, created_at
		FROM sessions
		WHERE refresh_token = ?
	`

	err := r.db.GetContext(ctx, session, query, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return session, nil
}

func (r *SqliteSessionRepository) DeleteSession(ctx context.Context, id string) error {
	var op = "repository.SqliteSessionRepository.DeleteSession"

	query := `
		DELETE FROM sessions
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%s: %s", op, "session not found")
	}

	return nil
}

func (r *SqliteSessionRepository) DeleteByUserID(ctx context.Context, userID string) error {
	var op = "repository.SqliteSessionRepository.DeleteByUserID"

	query := `
		DELETE FROM sessions
		WHERE user_id = ?
	`

	result, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%s: %s", op, "session not found")
	}

	return nil
}
