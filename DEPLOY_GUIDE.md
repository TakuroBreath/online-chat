# Руководство по переносу Docker-контейнеров на удаленную машину

Этот документ содержит подробные инструкции по сборке, публикации и развертыванию Docker-контейнеров приложения Online Chat на удаленной машине.

## Предварительные требования

- Docker и Docker Compose установлены на локальной машине
- Docker и Docker Compose установлены на удаленной машине
- Учетная запись на Docker Hub (или другом Docker-реестре)
- SSH-доступ к удаленной машине

## Автоматический способ (с использованием скрипта)

В проекте есть скрипт `deploy.sh`, который автоматизирует процесс сборки, публикации и развертывания контейнеров.

### Шаг 1: Сделать скрипт исполняемым

```bash
chmod +x deploy.sh
```

### Шаг 2: Запустить скрипт

```bash
# Только сборка и публикация образов
./deploy.sh your_dockerhub_username

# Сборка, публикация и развертывание на удаленной машине
./deploy.sh your_dockerhub_username user@remote-server
```

## Ручной способ

Если вы предпочитаете выполнить процесс вручную, следуйте этим инструкциям:

### Шаг 1: Сборка Docker-образов

```bash
docker-compose build
```

### Шаг 2: Тегирование образов

```bash
docker tag online-chat_auth-service your_dockerhub_username/online-chat-auth-service:latest
docker tag online-chat_chat-service your_dockerhub_username/online-chat-chat-service:latest
```

### Шаг 3: Авторизация в Docker Hub

```bash
docker login
```

### Шаг 4: Публикация образов в Docker Hub

```bash
docker push your_dockerhub_username/online-chat-auth-service:latest
docker push your_dockerhub_username/online-chat-chat-service:latest
```

### Шаг 5: Создание docker-compose.yml для удаленной машины

Создайте файл `docker-compose.remote.yml` со следующим содержимым:

```yaml
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
    image: your_dockerhub_username/online-chat-auth-service:latest
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
    image: your_dockerhub_username/online-chat-chat-service:latest
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
```

### Шаг 6: Копирование файла на удаленную машину

```bash
scp docker-compose.remote.yml user@remote-server:~/docker-compose.yml
```

### Шаг 7: Запуск контейнеров на удаленной машине

```bash
ssh user@remote-server "cd ~ && docker-compose up -d"
```

### Шаг 8: Проверка статуса контейнеров

```bash
ssh user@remote-server "docker ps"
```

## Настройка клиента для подключения к удаленным сервисам

После успешного развертывания сервисов на удаленной машине, необходимо настроить клиент для подключения к ним. Измените файл `.env` в директории `chat-client`:

```
CHAT_AUTH_SERVICE_ADDR=<IP-адрес-сервера>:50051
CHAT_SERVICE_ADDR=<IP-адрес-сервера>:50052
```

Замените `<IP-адрес-сервера>` на реальный IP-адрес вашего удаленного сервера.

## Устранение неполадок

### Проверка логов контейнеров

```bash
ssh user@remote-server "docker logs auth-service"
ssh user@remote-server "docker logs chat-service"
```

### Перезапуск контейнеров

```bash
ssh user@remote-server "docker-compose restart"
```

### Обновление образов на удаленной машине

Если вы обновили образы в Docker Hub и хотите обновить их на удаленной машине:

```bash
ssh user@remote-server "docker-compose pull && docker-compose up -d"
```

## Безопасность

- Рассмотрите возможность использования приватного Docker-реестра для конфиденциальных проектов
- Настройте брандмауэр на удаленной машине, чтобы разрешить только необходимые порты (50051 и 50052)
- Используйте SSL/TLS для защиты соединений между клиентом и сервисами