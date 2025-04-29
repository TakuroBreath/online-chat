package main

import (
	"context"
	"log"
	"os"
	"time"

	"auth.service/internal/app"
	repo "auth.service/internal/repository/postgres"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	ctx := context.Background()

	sqdsn := os.Getenv("DATABASE_URL")

	db, err := sqlx.Connect("postgres", sqdsn)
	if err != nil {
		log.Fatalln(err)
	}
	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verify database connection
	if err = db.PingContext(ctx); err != nil {
		log.Fatalf("Unable to establish connection to database: %v", err)
	}

	userRepo := repo.NewUserRepository(db)
	sessionRepo := repo.NewSessionRepository(db)

	application := app.NewApp(ctx, userRepo, sessionRepo, db)
	if err := application.Run(ctx); err != nil {
		log.Fatalln(err)
	}
}
