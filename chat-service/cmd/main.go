package main

import (
	"context"
	"log"
	"os"

	"chat.service/internal/app"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Загружаем переменные окружения из .env файла
	err := godotenv.Load()
	if err != nil {
		log.Println("Ошибка загрузки .env файла, используем переменные окружения")
	}

	ctx := context.Background()

	// Получаем путь к базе данных из переменных окружения
	sqdsn := os.Getenv("SQLITE_PATH")
	if sqdsn == "" {
		sqdsn = "./database/chat.db"
	}

	// Получаем адрес сервиса аутентификации
	authServiceAddr := os.Getenv("AUTH_SERVICE_ADDR")
	if authServiceAddr == "" {
		authServiceAddr = "localhost:50051"
	}

	// Подключаемся к базе данных
	db, err := sqlx.Connect("sqlite3", sqdsn)
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer db.Close()

	// Создаем и запускаем приложение
	application, err := app.NewApp(ctx, db, authServiceAddr)
	if err != nil {
		log.Fatalf("Ошибка создания приложения: %v", err)
	}

	if err := application.Run(ctx); err != nil {
		log.Fatalf("Ошибка запуска приложения: %v", err)
	}
}
