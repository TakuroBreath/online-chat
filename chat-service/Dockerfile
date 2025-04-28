FROM golang:1.24-alpine AS builder

WORKDIR /app

# Копируем файлы модулей и скачиваем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=1 GOOS=linux go build -a -o chat-service ./cmd/main.go

# Финальный образ
FROM alpine:latest

RUN apk add --no-cache ca-certificates sqlite

WORKDIR /app

# Создаем директорию для базы данных
RUN mkdir -p /app/database

# Копируем бинарный файл из предыдущего этапа
COPY --from=builder /app/chat-service /app/
COPY --from=builder /app/internal/migrations /app/internal/migrations

# Копируем .env файл
COPY .env /app/

# Открываем порт
EXPOSE 50052

# Запускаем приложение
CMD ["/app/chat-service"]