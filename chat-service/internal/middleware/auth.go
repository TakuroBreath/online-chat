package middleware

import (
	"context"
	"strings"

	"chat.service/internal/service/auth_client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthInterceptor перехватчик для аутентификации запросов
type AuthInterceptor struct {
	authClient *auth_client.AuthClient
}

// NewAuthInterceptor создает новый перехватчик аутентификации
func NewAuthInterceptor(authClient *auth_client.AuthClient) *AuthInterceptor {
	return &AuthInterceptor{
		authClient: authClient,
	}
}

// UnaryInterceptor перехватчик для унарных RPC
func (i *AuthInterceptor) UnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Извлекаем токен из метаданных запроса
	token, err := extractToken(ctx)
	if err != nil {
		return nil, err
	}

	// Проверяем токен и получаем ID пользователя
	userID, err := i.authClient.ValidateToken(ctx, token)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "недействительный токен доступа")
	}

	// Добавляем ID пользователя в контекст
	ctx = addUserIDToContext(ctx, userID)

	// Вызываем обработчик с обновленным контекстом
	return handler(ctx, req)
}

// StreamInterceptor перехватчик для потоковых RPC
func (i *AuthInterceptor) StreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	// Извлекаем токен из метаданных запроса
	token, err := extractToken(ss.Context())
	if err != nil {
		return err
	}

	// Проверяем токен и получаем ID пользователя
	userID, err := i.authClient.ValidateToken(ss.Context(), token)
	if err != nil {
		return status.Error(codes.Unauthenticated, "недействительный токен доступа")
	}

	// Создаем обертку для ServerStream с обновленным контекстом
	wrappedStream := &wrappedServerStream{
		ServerStream: ss,
		ctx:          addUserIDToContext(ss.Context(), userID),
	}

	// Вызываем обработчик с оберткой
	return handler(srv, wrappedStream)
}

// wrappedServerStream обертка для ServerStream с обновленным контекстом
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

// Context возвращает обновленный контекст
func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

// extractToken извлекает токен из метаданных запроса
func extractToken(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "метаданные не найдены")
	}

	values := md.Get("authorization")
	if len(values) == 0 {
		return "", status.Error(codes.Unauthenticated, "токен доступа не найден")
	}

	auth := values[0]
	if !strings.HasPrefix(auth, "Bearer ") {
		return "", status.Error(codes.Unauthenticated, "неверный формат токена")
	}

	return strings.TrimPrefix(auth, "Bearer "), nil
}

// addUserIDToContext добавляет ID пользователя в контекст
func addUserIDToContext(ctx context.Context, userID string) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(map[string]string{})
	} else {
		// Создаем копию метаданных, чтобы не изменять оригинал
		md = md.Copy()
	}

	// Добавляем ID пользователя в метаданные
	md.Set("user-id", userID)

	// Создаем новый контекст с обновленными метаданными
	return metadata.NewIncomingContext(ctx, md)
}
