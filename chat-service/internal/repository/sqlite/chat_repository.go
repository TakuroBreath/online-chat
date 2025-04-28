package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"chat.service/internal/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var (
	ErrChatNotFound  = errors.New("чат не найден")
	ErrUserNotInChat = errors.New("пользователь не является участником чата")
)

type ChatRepository struct {
	db *sqlx.DB
}

func NewChatRepository(db *sqlx.DB) *ChatRepository {
	return &ChatRepository{db: db}
}

func (r *ChatRepository) CreateChat(ctx context.Context, chat *models.Chat) (string, error) {
	if chat.ID == "" {
		chat.ID = uuid.New().String()
	}

	chat.CreatedAt = time.Now()

	query := `INSERT INTO chats (id, name, created_at, created_by_id) VALUES (?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, chat.ID, chat.Name, chat.CreatedAt, chat.CreatedByID)
	if err != nil {
		return "", err
	}

	return chat.ID, nil
}

func (r *ChatRepository) AddParticipant(ctx context.Context, chatID, userID string) error {
	// Проверяем существование чата
	_, err := r.GetChatByID(ctx, chatID)
	if err != nil {
		return err
	}

	// Проверяем, не является ли пользователь уже участником
	exists, err := r.CheckUserInChat(ctx, chatID, userID)
	if err != nil {
		return err
	}

	if exists {
		return nil // Пользователь уже участник
	}

	query := `INSERT INTO chat_participants (chat_id, user_id, joined_at) VALUES (?, ?, ?)`
	_, err = r.db.ExecContext(ctx, query, chatID, userID, time.Now())
	return err
}

func (r *ChatRepository) GetChatByID(ctx context.Context, chatID string) (*models.Chat, error) {
	var chat models.Chat

	query := `SELECT id, name, created_at, created_by_id FROM chats WHERE id = ?`
	err := r.db.GetContext(ctx, &chat, query, chatID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrChatNotFound
		}
		return nil, err
	}

	return &chat, nil
}

func (r *ChatRepository) GetChatParticipants(ctx context.Context, chatID string) ([]string, error) {
	var userIDs []string

	query := `SELECT user_id FROM chat_participants WHERE chat_id = ?`
	err := r.db.SelectContext(ctx, &userIDs, query, chatID)
	if err != nil {
		return nil, err
	}

	return userIDs, nil
}

func (r *ChatRepository) CheckUserInChat(ctx context.Context, chatID, userID string) (bool, error) {
	var count int

	query := `SELECT COUNT(*) FROM chat_participants WHERE chat_id = ? AND user_id = ?`
	err := r.db.GetContext(ctx, &count, query, chatID, userID)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// MessageRepository реализует интерфейс repository.MessageRepository
type MessageRepository struct {
	db *sqlx.DB
}

func NewMessageRepository(db *sqlx.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) SaveMessage(ctx context.Context, message *models.Message) (string, error) {
	if message.ID == "" {
		message.ID = uuid.New().String()
	}

	message.CreatedAt = time.Now()

	query := `INSERT INTO messages (id, chat_id, user_id, username, text, created_at) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(
		ctx,
		query,
		message.ID,
		message.ChatID,
		message.UserID,
		message.Username,
		message.Text,
		message.CreatedAt,
	)
	if err != nil {
		return "", err
	}

	return message.ID, nil
}

func (r *MessageRepository) GetChatMessages(ctx context.Context, chatID string, limit, offset int) ([]*models.Message, error) {
	var messages []*models.Message

	query := `SELECT id, chat_id, user_id, username, text, created_at FROM messages WHERE chat_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`
	err := r.db.SelectContext(ctx, &messages, query, chatID, limit, offset)
	if err != nil {
		return nil, err
	}

	return messages, nil
}
