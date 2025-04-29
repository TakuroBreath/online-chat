#!/bin/sh
set -e

echo "Ожидание готовности PostgreSQL..."
sleep 5

echo "Применение миграций базы данных..."

# Создаем базу данных, если она не существует
PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -U $POSTGRES_USER -tc "SELECT 1 FROM pg_database WHERE datname = '$POSTGRES_DB'" | grep -q 1 || PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -U $POSTGRES_USER -c "CREATE DATABASE $POSTGRES_DB"

# Применяем миграции с помощью golang-migrate
migrate -path /app/internal/migrations -database "${DATABASE_URL}" -verbose up

# Проверяем результат выполнения миграций
if [ $? -ne 0 ]; then
  echo "Ошибка при применении миграций!"
  exit 1
fi

echo "Миграции успешно применены."

# Запуск основного приложения
exec /app/auth-service