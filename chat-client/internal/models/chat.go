package models

import (
	"time"
)

type Chat struct {
	ID          string    `db:"id"`
	Name        string    `db:"name"`
	CreatedAt   time.Time `db:"created_at"`
	CreatedByID string    `db:"created_by_id"`
}

type ChatParticipant struct {
	ChatID   string    `db:"chat_id"`
	UserID   string    `db:"user_id"`
	JoinedAt time.Time `db:"joined_at"`
}

type Message struct {
	ID        string    `db:"id"`
	ChatID    string    `db:"chat_id"`
	UserID    string    `db:"user_id"`
	Username  string    `db:"username"`
	Text      string    `db:"text"`
	CreatedAt time.Time `db:"created_at"`
}
