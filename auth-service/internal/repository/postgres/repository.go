package postgres

import (
	"context"
	"fmt"
	"time"

	"auth.service/internal/repository"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type PostgresUserRepository struct {
	db *sqlx.DB
}

type PostgresSessionRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *PostgresUserRepository {
	return &PostgresUserRepository{
		db: db,
	}
}

func NewSessionRepository(db *sqlx.DB) *PostgresSessionRepository {
	return &PostgresSessionRepository{
		db: db,
	}
}

func (r *PostgresUserRepository) CreateUser(ctx context.Context, user *repository.User) error {
	var op = "repository.PostgresUserRepository.CreateUser"

	if user.ID == "" {
		uuid, err := uuid.Parse(uuid.New().String())
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		user.ID = uuid.String()
	}

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	query := `
		INSERT INTO users (id, username, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.ExecContext(ctx, query, user.ID, user.Username, user.PasswordHash, user.CreatedAt, user.UpdatedAt)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *PostgresUserRepository) UserByID(ctx context.Context, id string) (*repository.User, error) {
	var op = "repository.PostgresUserRepository.UserByID"

	query := `
		SELECT id, username, password_hash, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user repository.User
	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &user, nil
}

func (r *PostgresUserRepository) UserByUsername(ctx context.Context, username string) (*repository.User, error) {
	var op = "repository.PostgresUserRepository.UserByUsername"

	query := `
		SELECT id, username, password_hash, created_at, updated_at
		FROM users
		WHERE username = $1
	`

	user := &repository.User{}
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

func (r *PostgresUserRepository) UpdateUser(ctx context.Context, user *repository.User) error {
	var op = "repository.PostgresUserRepository.UpdateUser"

	user.UpdatedAt = time.Now()

	query := `
		UPDATE users
		SET username = $1, password_hash = $2, updated_at = $3
		WHERE id = $4
	`

	_, err := r.db.ExecContext(ctx, query, user.Username, user.PasswordHash, user.UpdatedAt, user.ID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *PostgresUserRepository) DeleteUser(ctx context.Context, id string) error {
	var op = "repository.PostgresUserRepository.DeleteUser"

	query := `
		DELETE FROM users
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *PostgresSessionRepository) CreateSession(ctx context.Context, session *repository.Session) error {
	var op = "repository.PostgresSessionRepository.CreateSession"

	if session.ID == "" {
		uuid, err := uuid.Parse(uuid.New().String())
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		session.ID = uuid.String()
	}

	query := `
		INSERT INTO sessions (id, user_id, refresh_token, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		session.ID,
		session.UserID,
		session.RefreshToken,
		session.ExpiresAt,
		session.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *PostgresSessionRepository) GetByRefreshToken(ctx context.Context, refreshToken string) (*repository.Session, error) {
	var op = "repository.PostgresSessionRepository.GetByRefreshToken"

	query := `
		SELECT id, user_id, refresh_token, expires_at, created_at
		FROM sessions
		WHERE refresh_token = $1
	`

	var session repository.Session
	err := r.db.GetContext(ctx, &session, query, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &session, nil
}

func (r *PostgresSessionRepository) DeleteSession(ctx context.Context, id string) error {
	var op = "repository.PostgresSessionRepository.DeleteSession"

	query := `
		DELETE FROM sessions
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *PostgresSessionRepository) DeleteByUserID(ctx context.Context, userID string) error {
	var op = "repository.PostgresSessionRepository.DeleteByUserID"

	query := `
		DELETE FROM sessions
		WHERE user_id = $1
	`

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
