package repository

import (
	"auth.service/internal/models"
	"context"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	UserByID(ctx context.Context, id string) (*models.User, error)
	UserByUsername(ctx context.Context, username string) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
	DeleteUser(ctx context.Context, id string) error
}

type SessionRepository interface {
	CreateSession(ctx context.Context, session *models.Session) error
	GetByRefreshToken(ctx context.Context, refreshToken string) (*models.Session, error)
	DeleteSession(ctx context.Context, id string) error
	DeleteByUserID(ctx context.Context, userID string) error
}
