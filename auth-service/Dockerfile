FROM golang:1.24-alpine AS builder

WORKDIR /app

# Устанавливаем необходимые пакеты для CGO и PostgreSQL
# build-base включает gcc, libc-dev и т.д., что нужно для CGO
RUN apk add --no-cache build-base postgresql-dev 

# Копируем файлы модулей и скачиваем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Добавляем зависимость golang-migrate
RUN go get -u github.com/golang-migrate/migrate/v4
RUN go get -u github.com/golang-migrate/migrate/v4/database/postgres
RUN go get -u github.com/golang-migrate/migrate/v4/source/file

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=1 GOOS=linux go build -a -o auth-service ./cmd/main.go

# Устанавливаем migrate CLI для управления миграциями
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Финальный образ
# Используем конкретную версию Alpine для предсказуемости
FROM alpine:3.19

# Устанавливаем только рантайм-зависимости PostgreSQL и сертификаты
RUN apk add --no-cache ca-certificates postgresql-client binutils

# Копируем утилиту migrate из образа builder
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate

WORKDIR /app

# Директория для базы данных больше не нужна, так как используем PostgreSQL

# Копируем бинарный файл приложения и миграции
COPY --from=builder /app/auth-service /app/
COPY --from=builder /app/internal/migrations /app/internal/migrations

# Копируем скрипт инициализации
COPY scripts/init.sh /app/init.sh
RUN chmod +x /app/init.sh

# Копируем .env файл (Рассмотрите передачу через переменные окружения!)
COPY .env /app/ 

# Открываем порт
EXPOSE 50051

# Запускаем скрипт инициализации
CMD ["/app/init.sh"]