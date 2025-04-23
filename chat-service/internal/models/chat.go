package models

import (
	"time"
)

// Chat представляет модель чата
type Chat struct {
	ID          string    `db:"id"`
	Name        string    `db:"name"`
	CreatedAt   time.Time `db:"created_at"`
	CreatedByID string    `db:"created_by_id"`
}

// ChatParticipant представляет участника чата
type ChatParticipant struct {
	ChatID   string    `db:"chat_id"`
	UserID   string    `db:"user_id"`
	JoinedAt time.Time `db:"joined_at"`
}

// Message представляет сообщение в чате
type Message struct {
	ID        string    `db:"id"`
	ChatID    string    `db:"chat_id"`
	UserID    string    `db:"user_id"`
	Username  string    `db:"username"`
	Text      string    `db:"text"`
	CreatedAt time.Time `db:"created_at"`
}
