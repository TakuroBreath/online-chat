# Этап 1: Сборка приложения и утилиты migrate
FROM golang:1.24-alpine AS builder

# Устанавливаем общую рабочую директорию
WORKDIR /app

# Устанавливаем необходимые пакеты для CGO и PostgreSQL
RUN apk add --no-cache build-base postgresql-dev

# --- Оптимизация кэширования зависимостей ---
# 1. Копируем файлы модулей *обоих* сервисов
COPY auth-service/go.mod auth-service/go.sum ./auth-service/
COPY chat-service/go.mod chat-service/go.sum ./chat-service/

# Важно: Убедитесь, что в chat-service/go.mod правильный путь в replace:
# replace <module-path-auth-service> => ../auth-service
# Например: replace github.com/your-org/auth-service => ../auth-service

# 2. Переходим в директорию chat-service
WORKDIR /app/chat-service

# 3. Скачиваем зависимости для chat-service.
# На этом этапе go найдет /app/auth-service/go.mod благодаря ../auth-service
RUN go mod download

# Добавляем зависимость golang-migrate
RUN go get -u github.com/golang-migrate/migrate/v4
RUN go get -u github.com/golang-migrate/migrate/v4/database/postgres
RUN go get -u github.com/golang-migrate/migrate/v4/source/file

# Устанавливаем migrate CLI для управления миграциями
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# 4. Возвращаемся в корневую директорию сборки
WORKDIR /app
# --- Конец оптимизации кэширования ---

# 5. Копируем ВЕСЬ исходный код auth-service
COPY auth-service/ ./auth-service/

# 6. Копируем ВЕСЬ исходный код chat-service (перезапишет go.mod/go.sum, но это ок)
COPY chat-service/ ./chat-service/

# 7. Собираем приложение chat-service С ВКЛЮЧЕННЫМ CGO
# Запускаем сборку из директории chat-service
RUN cd chat-service && CGO_ENABLED=1 GOOS=linux go build -a -o chat-service ./cmd/main.go

# Этап 2: Финальный образ на базе Alpine
FROM alpine:3.19

RUN apk add --no-cache ca-certificates postgresql-client binutils

# Установка необходимых зависимостей

WORKDIR /app

# Директория для базы данных больше не нужна, так как используем PostgreSQL

# Копируем бинарник chat-service (путь в builder: /app/chat-service/chat-service)
COPY --from=builder /app/chat-service/chat-service /app/

# Копируем миграции chat-service (путь в builder: /app/chat-service/internal/migrations)
COPY --from=builder /app/chat-service/internal/migrations /app/internal/migrations

# Копируем утилиту migrate из образа builder
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate

# Копируем скрипт инициализации
COPY chat-service/scripts/init.sh /app/init.sh
RUN chmod +x /app/init.sh

# Копируем .env файл (ВНИМАНИЕ: плохая практика для секретов!)
COPY chat-service/.env /app/

EXPOSE 50052

CMD ["/app/init.sh"]