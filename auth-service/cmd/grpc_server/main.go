package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"auth.service/internal/app"
	repo "auth.service/internal/repository/postgres"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	ctx := context.Background()

	// Database configuration
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	db, err := sqlx.Connect("postgres", dbURL)
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verify database connection
	if err = db.PingContext(ctx); err != nil {
		log.Fatalf("Unable to establish connection to database: %v", err)
	}

	// Initialize repositories
	userRepo := repo.NewUserRepository(db)
	sessionRepo := repo.NewSessionRepository(db)

	// Initialize application
	application := app.NewApp(ctx, userRepo, sessionRepo, db)

	// Get gRPC server port
	port := os.Getenv("GRPC_SERVER_PORT")
	if port == "" {
		port = "50051"
	}

	// Create listener
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Create gRPC server
	grpcServer := grpc.NewServer()

	// Register services
	application.RegisterServices(grpcServer)

	// Register reflection service on gRPC server
	reflection.Register(grpcServer)

	// Channel to listen for errors coming from the listener.
	serverErrors := make(chan error, 1)

	// Start the service listening for requests.
	go func() {
		log.Printf("Starting gRPC server on port %s", port)
		serverErrors <- grpcServer.Serve(lis)
	}()

	// Channel to listen for an interrupt or terminate signal from the OS.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Blocking main and waiting for shutdown.
	select {
	case err := <-serverErrors:
		log.Fatalf("gRPC server error: %v", err)
	case sig := <-shutdown:
		log.Printf("gRPC server shutting down: %v", sig)
		grpcServer.GracefulStop()
	}
}
