# Миграции базы данных с golang-migrate

## Структура миграций

Миграции в этом проекте используют инструмент [golang-migrate](https://github.com/golang-migrate/migrate).

Файлы миграций следуют следующему формату именования:
```
{version}_{description}.{up|down}.sql
```

Где:
- `version` - номер версии миграции (например, 000001, 000002)
- `description` - краткое описание миграции
- `up` - SQL для применения миграции
- `down` - SQL для отката миграции

## Создание новых миграций

Для создания новой миграции используйте команду:

```bash
migrate create -ext sql -dir ./internal/migrations -seq название_миграции
```

Эта команда создаст два файла:
- `{version}_{название_миграции}.up.sql` - для применения миграции
- `{version}_{название_миграции}.down.sql` - для отката миграции

## Применение миграций

Миграции применяются автоматически при запуске сервиса через скрипт `init.sh`.

Для ручного применения миграций используйте команду:

```bash
migrate -path ./internal/migrations -database "postgres://postgres:postgres@localhost:5432/chat_service?sslmode=disable" up
```

## Откат миграций

Для отката последней миграции:

```bash
migrate -path ./internal/migrations -database "postgres://postgres:postgres@localhost:5432/chat_service?sslmode=disable" down 1
```

Для отката всех миграций:

```bash
migrate -path ./internal/migrations -database "postgres://postgres:postgres@localhost:5432/chat_service?sslmode=disable" down
```