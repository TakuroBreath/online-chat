package chat_client

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// AuthInterceptor создает новый перехватчик для добавления токена аутентификации
func AuthInterceptor(token string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// Создаем новый контекст с метаданными, содержащими токен
		md := metadata.New(map[string]string{
			"authorization": "Bearer " + token,
		})
		ctx = metadata.NewOutgoingContext(ctx, md)

		// Вызываем оригинальный метод с обновленным контекстом
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// StreamAuthInterceptor создает новый перехватчик для добавления токена аутентификации в стримы
func StreamAuthInterceptor(token string) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		// Создаем новый контекст с метаданными, содержащими токен
		md := metadata.New(map[string]string{
			"authorization": "Bearer " + token,
		})
		ctx = metadata.NewOutgoingContext(ctx, md)

		// Вызываем оригинальный стример с обновленным контекстом
		return streamer(ctx, desc, cc, method, opts...)
	}
}
