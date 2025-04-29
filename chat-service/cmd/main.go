package main

import (
	"context"
	"log"
	"os"

	"chat.service/internal/app"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // PostgreSQL драйвер
)

func main() {
	// Загружаем переменные окружения из .env файла
	err := godotenv.Load()
	if err != nil {
		log.Println("Ошибка загрузки .env файла, используем переменные окружения")
	}

	ctx := context.Background()

	// Получаем строку подключения к базе данных из переменных окружения
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatalf("Не указана строка подключения к базе данных (DATABASE_URL)")
	}

	// Получаем адрес сервиса аутентификации
	authServiceAddr := os.Getenv("AUTH_SERVICE_ADDR")
	if authServiceAddr == "" {
		authServiceAddr = "localhost:50051"
	}

	// Подключаемся к базе данных PostgreSQL
	db, err := sqlx.Connect("postgres", dbURL)
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer db.Close()

	// Создаем и запускаем приложение с PostgreSQL
	application, err := app.NewPostgresApp(ctx, db, authServiceAddr)
	if err != nil {
		log.Fatalf("Ошибка создания приложения: %v", err)
	}

	if err := application.Run(ctx); err != nil {
		log.Fatalf("Ошибка запуска приложения: %v", err)
	}
}
