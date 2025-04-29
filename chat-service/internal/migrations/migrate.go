package migrations

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jmoiron/sqlx"
	// Важно: нужно импортировать драйвер источника миграций (в данном случае 'file')
	_ "github.com/golang-migrate/migrate/v4/source/file" // Используем _ для регистрации драйвера
	// Если бы мы создавали источник вручную через sourcefile.New(), импорт был бы таким:
	// sourcefile "github.com/golang-migrate/migrate/v4/source/file"
)

// migrateLogger реализует интерфейс migrate.Logger для интеграции с log.Printf
type migrateLogger struct{}

func (l *migrateLogger) Printf(format string, v ...interface{}) {
	log.Printf("migrate: "+format, v...)
}

func (l *migrateLogger) Verbose() bool {
	// Установите true, если хотите видеть подробные логи от golang-migrate
	return true // Или false, по вашему усмотрению
}

// RunMigrations выполняет миграции базы данных с использованием golang-migrate
func RunMigrations(db *sqlx.DB, migrationsPath string) error {
	sqlDB := db.DB
	log.Println("Запуск миграций базы данных...")

	// Проверяем существование директории с миграциями
	fileInfo, err := os.Stat(migrationsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("директория миграций не существует: %s", migrationsPath)
		}
		return fmt.Errorf("ошибка проверки директории миграций %s: %w", migrationsPath, err)
	}
	if !fileInfo.IsDir() {
		return fmt.Errorf("указанный путь миграций не является директорией: %s", migrationsPath)
	}

	// Получаем абсолютный путь к директории миграций
	absPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		return fmt.Errorf("ошибка получения абсолютного пути для '%s': %w", migrationsPath, err)
	}

	// Формируем URL для источника файлов.
	// golang-migrate ожидает URL в формате file://<полный_путь>
	sourceURL := fmt.Sprintf("file://%s", absPath)
	log.Printf("Источник миграций: %s\n", sourceURL) // Логируем для отладки

	// Создаем драйвер для postgres
	// Можно установить таймауты и другие параметры здесь
	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{
		MigrationsTable:  "schema_migrations", // Имя таблицы для хранения версий миграций (по умолчанию)
		DatabaseName:     "postgres",          // Можно указать имя БД для логов/ошибок, если нужно
		StatementTimeout: 60 * time.Second,    // Таймаут выполнения одного SQL стейтмента миграции
	})
	if err != nil {
		return fmt.Errorf("ошибка создания драйвера postgres: %w", err)
	}
	// Важно убедиться, что драйвер будет закрыт, хотя WithInstance обычно не создает новое соединение
	// defer driver.Close() // Не нужно для WithInstance, т.к. он использует существующий *sql.DB

	// Создаем экземпляр migrate, используя URL источника и драйвер базы данных
	// Вместо NewWithInstance используем NewWithDatabaseInstance, он удобнее, когда есть URL источника
	m, err := migrate.NewWithDatabaseInstance(
		sourceURL,  // Источник миграций (например, "file://./migrations")
		"postgres", // Имя базы данных (для логов/идентификации)
		driver,     // Драйвер базы данных
	)
	if err != nil {
		// Частая ошибка здесь - неверный sourceURL или проблемы с доступом к файлам
		return fmt.Errorf("ошибка создания экземпляра migrate: %w", err)
	}

	// Устанавливаем кастомный логгер (опционально)
	m.Log = &migrateLogger{}

	// Устанавливаем таймаут на всю операцию миграции (если нужно, но StatementTimeout в драйвере часто достаточно)
	// m.LockTimeout = 15 * time.Second // Таймаут на получение блокировки БД

	log.Println("Выполнение миграций (Up)...")
	// Выполняем миграции "вверх"
	err = m.Up() // Метод WithTimeout() отсутствует, таймауты настраиваются в драйвере или через LockTimeout

	// Обрабатываем результат
	if err != nil {
		// migrate.ErrNoChange - это не ошибка, а индикатор, что все миграции уже применены
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("Миграции не требуются, база данных в актуальном состоянии.")
		} else {
			// В случае реальной ошибки, возвращаем её
			// НЕ закрываем соединение с БД, так как оно будет использоваться дальше
			log.Printf("Ошибка выполнения миграций: %v\n", err)
			return fmt.Errorf("ошибка выполнения миграций: %w", err)
		}
	} else {
		log.Println("Миграции успешно выполнены.")
	}

	log.Println("Миграции завершены, продолжаем работу с тем же соединением БД")

	return nil // Возвращаем nil, если ошибок не было или была только migrate.ErrNoChange
}
