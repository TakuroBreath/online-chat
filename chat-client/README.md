# Chat Client

Это клиент командной строки для взаимодействия с микросервисами онлайн-чата (`auth-service` и `chat-service`).

## Функциональность

*   Регистрация нового пользователя (`register`).
*   Вход пользователя в систему (`login`) для получения токена аутентификации.
*   Создание нового чата (`create`).
*   Подключение к существующему чату по ID (`connect`).
*   Отправка и получение сообщений в реальном времени.

## Использование

Клиент использует библиотеку Cobra для управления командами.

1.  **Установите зависимости:**
    ```bash
    go mod tidy
    ```
2.  **Соберите исполняемый файл (опционально):**
    ```bash
    go build -o chatik main.go
    ```
    Или используйте `go run main.go`.

3.  **Настройте переменные окружения:**
    Создайте файл `.env` в корне каталога `chat-client`:
    ```dotenv
    CHAT_AUTH_SERVICE_ADDR=localhost:50051 # Адрес сервиса аутентификации
    CHAT_SERVICE_ADDR=localhost:50052    # Адрес сервиса чатов
    ```

4.  **Запустите клиент с нужной командой:**

    *   **Регистрация:**
        ```bash
        ./chatik register -u <username> -p <password>
        # или go run main.go register -u <username> -p <password>
        ```
    *   **Вход:**
        ```bash
        ./chatik login -u <username> -p <password>
        # Запомните полученный токен!
        ```
    *   **Создание чата:**
        ```bash
        ./chatik create -n "<chat_name>" -t <your_auth_token>
        # Запомните полученный ID чата!
        ```
    *   **Подключение к чату:**
        ```bash
        ./chatik connect -i <chat_id> -t <your_auth_token>
        ```
        После подключения вы можете отправлять сообщения, вводя их в консоль и нажимая Enter. Для выхода нажмите Ctrl+C.

## Зависимости

*   Работающие экземпляры `auth-service` и `chat-service`.

## Конфигурация

Клиент конфигурируется с помощью переменных окружения:

*   `CHAT_AUTH_SERVICE_ADDR`: Адрес и порт gRPC сервера `auth-service`.
*   `CHAT_SERVICE_ADDR`: Адрес и порт gRPC сервера `chat-service`.