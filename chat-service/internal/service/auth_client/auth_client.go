package auth_client

import (
	"context"
	"errors"
	"log"

	authpb "auth.service/api/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	ErrUserNotFound = errors.New("пользователь не найден")
	ErrInvalidToken = errors.New("недействительный токен")
)

// AuthClient предоставляет методы для взаимодействия с сервисом аутентификации
type AuthClient struct {
	accessClient authpb.AccessServiceClient
	userClient   authpb.UserServiceClient
	conn         *grpc.ClientConn
}

// NewAuthClient создает новый клиент для сервиса аутентификации
func NewAuthClient(authServiceAddr string) (*AuthClient, error) {
	// Устанавливаем соединение с сервисом аутентификации
	conn, err := grpc.NewClient(authServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	// Создаем клиенты для сервисов аутентификации
	accessClient := authpb.NewAccessServiceClient(conn)
	userClient := authpb.NewUserServiceClient(conn)

	return &AuthClient{
		accessClient: accessClient,
		userClient:   userClient,
		conn:         conn,
	}, nil
}

// Close закрывает соединение с сервисом аутентификации
func (c *AuthClient) Close() error {
	return c.conn.Close()
}

// ValidateToken проверяет токен доступа и возвращает ID пользователя
func (c *AuthClient) ValidateToken(ctx context.Context, token string) (string, error) {
	// Вызываем метод проверки токена
	resp, err := c.accessClient.Check(ctx, &authpb.CheckAccessRequest{
		AccessToken: token,
	})
	if err != nil {
		log.Printf("Ошибка при проверке токена: %v", err)
		return "", ErrInvalidToken
	}

	if !resp.IsValid {
		return "", ErrInvalidToken
	}

	return resp.UserId, nil
}

// GetUserByID возвращает имя пользователя по ID
func (c *AuthClient) GetUserByID(ctx context.Context, userID string) (string, error) {
	// Вызываем метод получения пользователя
	resp, err := c.userClient.GetUser(ctx, &authpb.GetUserRequest{
		UserId: userID,
	})

	if err != nil {
		log.Printf("Ошибка при получении пользователя: %v", err)
		return "", ErrUserNotFound
	}

	return resp.Username, nil
}
