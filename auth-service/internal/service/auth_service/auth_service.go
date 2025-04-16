package auth_service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"auth.service/internal/repository"
	"auth.service/internal/service"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken  = errors.New("Invalid token")
	ErrExpiredToken  = errors.New("Expired token")
	ErrTokenNotFound = errors.New("Token not found")
)

type CustomClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type AuthServiceImpl struct {
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository
	jwtSecret   []byte
	accessTTL   time.Duration
	refreshTTL  time.Duration
}

func NewAuthService(userRepo repository.UserRepository, sessionRepo repository.SessionRepository, jwtSecret []byte, accessTTL time.Duration, refreshTTL time.Duration) *AuthServiceImpl {
	return &AuthServiceImpl{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		jwtSecret:   []byte(getEnv("JWT_SECRET_KEY", "capybara_enjoys")),
		accessTTL:   parseDuration(getEnv("ACCESS_TOKEN_TTL", "15m")),
		refreshTTL:  parseDuration(getEnv("REFRESH_TOKEN_TTL", "24h")),
	}
}

func (s *AuthServiceImpl) Login(ctx context.Context, username, password string) (*service.TokenPair, error) {
	op := "AuthService.Login"

	user, err := s.userRepo.UserByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	if !checkPasswordHash(password, user.PasswordHash) {
		return nil, ErrInvalidCredentials
	}

	u := &service.User{
		ID:       user.ID,
		Username: user.Username,
	}

	return s.createTokens(ctx, u)
}

func (s *AuthServiceImpl) RefreshTokens(ctx context.Context, refreshToken string) (*service.TokenPair, error) {
	op := "AuthService.RefreshTokens"

	session, err := s.sessionRepo.GetByRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if time.Now().After(session.ExpiresAt) {
		_ = s.sessionRepo.DeleteSession(ctx, session.ID)
	}

	user, err := s.userRepo.UserByID(ctx, session.UserID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_ = s.sessionRepo.DeleteSession(ctx, session.ID)

	u := &service.User{
		ID:       user.ID,
		Username: user.Username,
	}

	return s.createTokens(ctx, u)
}

func (s *AuthServiceImpl) ValidateToken(ctx context.Context, accessToken string) (*service.TokenClaims, error) {
	token, err := jwt.ParseWithClaims(accessToken, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		if time.Now().Unix() > claims.ExpiresAt.Time.Unix() {
			return nil, ErrExpiredToken
		}

		return &service.TokenClaims{
			UserID:    claims.UserID,
			Username:  claims.Username,
			ExpiresAt: claims.ExpiresAt.Time,
		}, nil
	}

	return nil, ErrInvalidToken
}

func (s *AuthServiceImpl) createTokens(ctx context.Context, user *service.User) (*service.TokenPair, error) {
	op := "AuthService.createTokens"

	now := time.Now()

	accessToken, err := s.generateAccessToken(user, now)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	refresh_token := uuid.New().String()

	session := &repository.Session{
		UserID:       user.ID,
		RefreshToken: refresh_token,
		ExpiresAt:    now.Add(s.refreshTTL),
	}

	if err := s.sessionRepo.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &service.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refresh_token,
		UserID:       user.ID,
	}, nil
}

func (s *AuthServiceImpl) generateAccessToken(user *service.User, now time.Time) (string, error) {
	claims := CustomClaims{
		UserID:   user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "auth.service",
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultValue
}

func parseDuration(value string) time.Duration {
	duration, err := time.ParseDuration(value)
	if err != nil {
		if hours, err := strconv.Atoi(value); err == nil {
			return time.Duration(hours) * time.Hour
		}
		return 15 * time.Minute
	}

	return duration
}
