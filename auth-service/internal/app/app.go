package app

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "auth.service/api/proto"
	"auth.service/internal/api"
	"auth.service/internal/repository"
	"auth.service/internal/service/auth_service"
	"github.com/jmoiron/sqlx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type App struct {
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository
	grpcServer  *grpc.Server
	port        string
	db          *sqlx.DB
}

func NewApp(ctx context.Context, userRepo repository.UserRepository, sessionRepo repository.SessionRepository, db *sqlx.DB) *App {
	// Run migrations during app initialization
	if err := InitMigrations(db); err != nil {
		log.Printf("Error executing migrations: %v", err)
	}
	port := getEnv("GRPC_PORT", "50051")

	return &App{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		port:        port,
		db:          db,
	}
}

func (a *App) RegisterServices(grpcServer *grpc.Server) {
	userService := auth_service.NewUserService(a.userRepo)
	authService := auth_service.NewAuthService(
		a.userRepo,
		a.sessionRepo,
		nil,
		time.Duration(0),
		time.Duration(0),
	)
	accessService := auth_service.NewAccessService(authService)

	userHandler := api.NewUserServiceHandler(userService)
	authHandler := api.NewAuthServiceHandler(authService)
	accessHandler := api.NewAccessServiceHandler(accessService)

	pb.RegisterUserServiceServer(grpcServer, userHandler)
	pb.RegisterAuthServiceServer(grpcServer, authHandler)
	pb.RegisterAccessServiceServer(grpcServer, accessHandler)

	reflection.Register(grpcServer)
}

func (a *App) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Проверяем соединение с базой данных перед запуском сервисов
	if err := a.db.PingContext(ctx); err != nil {
		log.Printf("Ошибка соединения с базой данных: %v", err)
		return err
	}

	a.grpcServer = grpc.NewServer()
	a.RegisterServices(a.grpcServer)

	go func() {
		lis, err := net.Listen("tcp", ":"+a.port)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}

		log.Printf("Starting gRPC server on port %s", a.port)
		if err := a.grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	return a.GracefulShutdown(ctx)
}

func (a *App) GracefulShutdown(ctx context.Context) error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ctx.Done():
		log.Println("Shutdown requested via context")
	case <-quit:
		log.Println("Shutdown requested via signal")
	}

	log.Println("Shutting down gRPC server...")
	a.grpcServer.GracefulStop()
	log.Println("Server gracefully stopped")

	log.Println("Closing database connection...")
	if err := a.db.Close(); err != nil {
		log.Printf("Error closing database connection: %v", err)
	}
	log.Println("Database connection closed")

	return nil
}

func getEnv(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}
