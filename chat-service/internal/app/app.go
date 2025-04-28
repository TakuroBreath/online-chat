package app

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	pb "chat.service/api/proto"
	"chat.service/internal/api"
	"chat.service/internal/middleware"
	"chat.service/internal/repository/sqlite"
	"chat.service/internal/service/auth_client"
	"chat.service/internal/service/chat_service"
	"github.com/jmoiron/sqlx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// App представляет приложение чат-сервиса
type App struct {
	chatRepo    *sqlite.ChatRepository
	messageRepo *sqlite.MessageRepository
	authClient  *auth_client.AuthClient
	grpcServer  *grpc.Server
	port        string
}

// NewApp создает новый экземпляр приложения
func NewApp(ctx context.Context, db *sqlx.DB, authServiceAddr string) (*App, error) {
	port := getEnv("GRPC_PORT", "50052")

	// Создаем репозитории
	chatRepo := sqlite.NewChatRepository(db)
	messageRepo := sqlite.NewMessageRepository(db)

	// Создаем клиент для сервиса аутентификации
	authClient, err := auth_client.NewAuthClient(authServiceAddr)
	if err != nil {
		return nil, err
	}

	return &App{
		chatRepo:    chatRepo,
		messageRepo: messageRepo,
		authClient:  authClient,
		port:        port,
	}, nil
}

// Run запускает приложение
func (a *App) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Создаем сервис чата
	chatService := chat_service.NewChatService(a.chatRepo, a.messageRepo, a.authClient)

	// Создаем обработчик API
	chatHandler := api.NewChatServiceHandler(chatService)

	// Создаем аутентификационный перехватчик
	authInterceptor := middleware.NewAuthInterceptor(a.authClient)

	// Создаем gRPC сервер с перехватчиками
	a.grpcServer = grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor.UnaryInterceptor),
		grpc.StreamInterceptor(authInterceptor.StreamInterceptor),
	)

	// Регистрируем сервисы
	pb.RegisterChatServiceServer(a.grpcServer, chatHandler)

	// Включаем reflection для удобства отладки
	reflection.Register(a.grpcServer)

	// Запускаем сервер в отдельной горутине
	go func() {
		lis, err := net.Listen("tcp", ":"+a.port)
		if err != nil {
			log.Fatalf("Ошибка при запуске сервера: %v", err)
		}

		log.Printf("Запуск gRPC сервера на порту %s", a.port)
		if err := a.grpcServer.Serve(lis); err != nil {
			log.Fatalf("Ошибка при запуске сервера: %v", err)
		}
	}()

	return a.GracefulShutdown(ctx)
}

// GracefulShutdown обеспечивает корректное завершение работы приложения
func (a *App) GracefulShutdown(ctx context.Context) error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ctx.Done():
		log.Println("Завершение работы по контексту")
	case <-quit:
		log.Println("Завершение работы по сигналу")
	}

	log.Println("Остановка gRPC сервера...")
	a.grpcServer.GracefulStop()
	log.Println("Сервер остановлен")

	// Закрываем соединение с сервисом аутентификации
	if a.authClient != nil {
		a.authClient.Close()
	}

	return nil
}

// getEnv возвращает значение переменной окружения или значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}
