CREATE TABLE IF NOT EXISTS chats (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    created_by_id UUID NOT NULL
);

-- Таблица участников чата
CREATE TABLE IF NOT EXISTS chat_participants (
    chat_id UUID NOT NULL,
    user_id UUID NOT NULL,
    joined_at TIMESTAMP NOT NULL,
    PRIMARY KEY (chat_id, user_id),
    FOREIGN KEY (chat_id) REFERENCES chats (id) ON DELETE CASCADE
);

-- Таблица сообщений
CREATE TABLE IF NOT EXISTS messages (
    id UUID PRIMARY KEY,
    chat_id UUID NOT NULL,
    user_id UUID NOT NULL,
    username VARCHAR(255) NOT NULL,
    text TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    FOREIGN KEY (chat_id) REFERENCES chats (id) ON DELETE CASCADE
);

-- Индексы для ускорения запросов
CREATE INDEX IF NOT EXISTS idx_messages_chat_id ON messages (chat_id);

CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages (created_at);

CREATE INDEX IF NOT EXISTS idx_chat_participants_user_id ON chat_participants (user_id);