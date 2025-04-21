package main

import (
	"context"
	"log"
	"os"

	"auth.service/internal/app"
	repo "auth.service/internal/repository/sqlite"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	ctx := context.Background()

	sqdsn := os.Getenv("SQLITE_PATH")

	db, err := sqlx.Connect("sqlite3", sqdsn)
	if err != nil {
		log.Fatalln(err)
	}

	userRepo := repo.NewUserRepository(db)
	sessionRepo := repo.NewSessionRepository(db)

	application := app.NewApp(ctx, userRepo, sessionRepo)
	if err := application.Run(ctx); err != nil {
		log.Fatalln(err)
	}
}
