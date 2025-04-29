#!/bin/bash

# Скрипт для сборки, публикации и развертывания Docker-образов

# Проверка наличия аргументов
if [ $# -lt 1 ]; then
  echo "Использование: $0 <docker_username> [remote_host]"
  echo "Пример: $0 myusername user@remote-server.com"
  exit 1
fi

DOCKER_USERNAME=$1
REMOTE_HOST=$2

echo "=== Сборка Docker-образов ==="
docker-compose build

# Тегирование образов
echo "=== Тегирование образов ==="
docker tag online-chat_auth-service ${DOCKER_USERNAME}/online-chat-auth-service:latest
docker tag online-chat_chat-service ${DOCKER_USERNAME}/online-chat-chat-service:latest

# Публикация образов в Docker Hub
echo "=== Публикация образов в Docker Hub ==="
echo "Авторизуйтесь в Docker Hub:"
docker login

docker push ${DOCKER_USERNAME}/online-chat-auth-service:latest
docker push ${DOCKER_USERNAME}/online-chat-chat-service:latest

# Создание файла docker-compose для удаленной машины
echo "=== Создание docker-compose.remote.yml ==="
cat > docker-compose.remote.yml << EOL
version: '3.8'

services:
  postgres:
    image: postgres:14
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: postgres
    volumes:
      - postgres-data:/var/lib/postgresql/data
    ports:
      - "5432:5432"
    networks:
      - chat-network
    restart: unless-stopped

  auth-service:
    image: ${DOCKER_USERNAME}/online-chat-auth-service:latest
    container_name: auth-service
    environment:
      - DATABASE_URL=postgres://postgres:postgres@postgres:5432/auth_service?sslmode=disable
      - GOOSE_DRIVER=postgres
      - GOOSE_DBSTRING=postgres://postgres:postgres@postgres:5432/auth_service?sslmode=disable
      - GOOSE_MIGRATION_DIR=/app/internal/migrations
    ports:
      - "50051:50051"
    depends_on:
      - postgres
    networks:
      - chat-network
    restart: unless-stopped

  chat-service:
    image: ${DOCKER_USERNAME}/online-chat-chat-service:latest
    container_name: chat-service
    environment:
      - GRPC_PORT=50052
      - DATABASE_URL=postgres://postgres:postgres@postgres:5432/chat_service?sslmode=disable
      - GOOSE_DRIVER=postgres
      - GOOSE_DBSTRING=postgres://postgres:postgres@postgres:5432/chat_service?sslmode=disable
      - GOOSE_MIGRATION_DIR=/app/internal/migrations
      - AUTH_SERVICE_ADDR=auth-service:50051
    ports:
      - "50052:50052"
    depends_on:
      - postgres
      - auth-service
    networks:
      - chat-network
    restart: unless-stopped

volumes:
  postgres-data:

networks:
  chat-network:
    driver: bridge
EOL

echo "=== Создание инструкций для удаленной машины ==="
cat > remote_instructions.txt << EOL
Инструкции по развертыванию на удаленной машине:

1. Установите Docker и Docker Compose на удаленную машину:
   https://docs.docker.com/engine/install/
   https://docs.docker.com/compose/install/

2. Скопируйте файл docker-compose.remote.yml на удаленную машину:
   scp docker-compose.remote.yml user@remote-server:/path/to/destination/docker-compose.yml

3. Запустите контейнеры на удаленной машине:
   ssh user@remote-server "cd /path/to/destination && docker-compose up -d"

4. Проверьте, что контейнеры запущены:
   ssh user@remote-server "docker ps"

5. Настройте клиент для подключения к удаленным сервисам, изменив файл .env в директории chat-client:
   CHAT_AUTH_SERVICE_ADDR=<IP-адрес-сервера>:50051
   CHAT_SERVICE_ADDR=<IP-адрес-сервера>:50052
EOL

# Если указан удаленный хост, копируем файлы и запускаем контейнеры
if [ ! -z "$REMOTE_HOST" ]; then
  echo "=== Копирование файлов на удаленную машину ==="
  scp docker-compose.remote.yml ${REMOTE_HOST}:~/docker-compose.yml
  
  echo "=== Запуск контейнеров на удаленной машине ==="
  ssh ${REMOTE_HOST} "docker-compose up -d"
  
  echo "=== Проверка статуса контейнеров ==="
  ssh ${REMOTE_HOST} "docker ps"
fi

echo "=== Готово! ==="
echo "Образы опубликованы: ${DOCKER_USERNAME}/online-chat-auth-service:latest и ${DOCKER_USERNAME}/online-chat-chat-service:latest"
echo "Инструкции по развертыванию сохранены в файле remote_instructions.txt"

if [ -z "$REMOTE_HOST" ]; then
  echo "Для развертывания на удаленной машине выполните: $0 ${DOCKER_USERNAME} user@remote-server"
fi