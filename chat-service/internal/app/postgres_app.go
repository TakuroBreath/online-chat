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
	"chat.service/internal/repository"
	"chat.service/internal/repository/postgres"
	"chat.service/internal/service/auth_client"
	"chat.service/internal/service/chat_service"
	"github.com/jmoiron/sqlx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// PostgresApp представляет приложение чат-сервиса с PostgreSQL
type PostgresApp struct {
	chatRepo    repository.ChatRepository
	messageRepo repository.MessageRepository
	authClient  *auth_client.AuthClient
	grpcServer  *grpc.Server
	port        string
}

// NewPostgresApp создает новый экземпляр приложения с PostgreSQL
func NewPostgresApp(ctx context.Context, db *sqlx.DB, authServiceAddr string) (*PostgresApp, error) {
	if err := InitMigrations(db); err != nil {
		log.Printf("Ошибка при выполнении миграций: %v", err)
	}
	port := getEnv("GRPC_PORT", "50052")

	// Создаем репозитории PostgreSQL
	chatRepo := postgres.NewChatRepository(db)
	messageRepo := postgres.NewMessageRepository(db)

	// Создаем клиент для сервиса аутентификации
	authClient, err := auth_client.NewAuthClient(authServiceAddr)
	if err != nil {
		return nil, err
	}

	return &PostgresApp{
		chatRepo:    chatRepo,
		messageRepo: messageRepo,
		authClient:  authClient,
		port:        port,
	}, nil
}

// Run запускает приложение
func (a *PostgresApp) Run(ctx context.Context) error {
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
func (a *PostgresApp) GracefulShutdown(ctx context.Context) error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ctx.Done():
		log.Println("Завершение работы по контексту")
	case sig := <-quit:
		log.Printf("Получен сигнал: %s", sig.String())
	}

	log.Println("Завершение работы сервера...")
	a.grpcServer.GracefulStop()
	log.Println("Сервер остановлен")

	return nil
}
