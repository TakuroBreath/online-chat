package app

import (
	"log"
	"os"
	"path/filepath"

	"chat.service/internal/migrations"
	"github.com/jmoiron/sqlx"
)

// InitMigrations инициализирует и запускает миграции базы данных
func InitMigrations(db *sqlx.DB) error {
	// Получаем путь к директории миграций из переменной окружения или используем значение по умолчанию
	migrationsDir := os.Getenv("MIGRATIONS_DIR")
	if migrationsDir == "" {
		// Если переменная окружения не установлена, используем путь по умолчанию
		migrationsDir = "/app/internal/migrations"
	}

	// Проверяем существование директории
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		log.Printf("Директория миграций %s не существует, пробуем найти относительный путь", migrationsDir)

		// Пробуем найти директорию относительно текущего рабочего каталога
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		migrationsDir = filepath.Join(cwd, "internal", "migrations")
		if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
			log.Printf("Директория миграций %s также не существует", migrationsDir)
			return err
		}
	}

	log.Printf("Запуск миграций из директории: %s", migrationsDir)
	return migrations.RunMigrations(db, migrationsDir)
}
