package sqlite

import (
	"context"
	"fmt"
	"time"

	"auth.service/internal/models"
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

func (r *SqliteUserRepository) CreateUser(ctx context.Context, user *models.User) error {
	var op = "repository.SqliteUserRepository.CreateUser"

	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	query := `
		INSERT INTO users (id, username, password_hach, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.ExecContext(ctx, query, user.ID, user.Username, user.Password, user.CreatedAt, user.UpdatedAt)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *SqliteUserRepository) UserByID(ctx context.Context, id string) (*models.User, error) {
	var op = "repository.SqliteUserRepository.UserByID"

	user := &models.User{}

	query := `
		SELECT id, username, password_hash, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	err := r.db.GetContext(ctx, user, query, id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (r *SqliteUserRepository) UserByUsername(ctx context.Context, username string) (*models.User, error) {
	var op = "repository.SqliteUserRepository.UserByUsername"

	user := &models.User{}

	query := `
		SELECT id, username, password_hash, created_at, updated_at
		FROM users
		WHERE username = $1
	`

	err := r.db.GetContext(ctx, user, query, username)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (r *SqliteUserRepository) UpdateUser(ctx context.Context, user *models.User) error {
	var op = "repository.SqliteUserRepository.UpdateUser"

	user.UpdatedAt = time.Now()

	query := `
		UPDATE users
		SET username = $1, password_hash = $2, updated_at = $3
		WHERE id = $4
	`

	result, err := r.db.ExecContext(ctx, query, user.Username, user.Password, user.UpdatedAt, user.ID)
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
		WHERE id = $1
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
