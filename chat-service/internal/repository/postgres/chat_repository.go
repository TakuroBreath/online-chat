package postgres

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
		uuid, err := uuid.Parse(uuid.New().String())
		if err != nil {
			return "", err
		}
		chat.ID = uuid.String()
	}

	chat.CreatedAt = time.Now()

	query := `INSERT INTO chats (id, name, created_at, created_by_id) VALUES ($1, $2, $3, $4)`
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

	// Добавляем пользователя в чат
	query := `INSERT INTO chat_participants (chat_id, user_id, joined_at) VALUES ($1, $2, $3)`
	_, err = r.db.ExecContext(ctx, query, chatID, userID, time.Now())
	if err != nil {
		return err
	}

	return nil
}

func (r *ChatRepository) GetChatByID(ctx context.Context, chatID string) (*models.Chat, error) {
	query := `SELECT id, name, created_at, created_by_id FROM chats WHERE id = $1`

	var chat models.Chat
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
	query := `SELECT user_id FROM chat_participants WHERE chat_id = $1`

	var userIDs []string
	err := r.db.SelectContext(ctx, &userIDs, query, chatID)
	if err != nil {
		return nil, err
	}

	return userIDs, nil
}

func (r *ChatRepository) CheckUserInChat(ctx context.Context, chatID, userID string) (bool, error) {
	query := `SELECT COUNT(*) FROM chat_participants WHERE chat_id = $1 AND user_id = $2`

	var count int
	err := r.db.GetContext(ctx, &count, query, chatID, userID)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

type MessageRepository struct {
	db *sqlx.DB
}

func NewMessageRepository(db *sqlx.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) SaveMessage(ctx context.Context, message *models.Message) (string, error) {
	if message.ID == "" {
		uuid, err := uuid.Parse(uuid.New().String())
		if err != nil {
			return "", err
		}
		message.ID = uuid.String()
	}

	message.CreatedAt = time.Now()

	query := `INSERT INTO messages (id, chat_id, user_id, username, text, created_at) VALUES ($1, $2, $3, $4, $5, $6)`
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
	query := `
		SELECT id, chat_id, user_id, username, text, created_at 
		FROM messages 
		WHERE chat_id = $1 
		ORDER BY created_at DESC 
		LIMIT $2 OFFSET $3
	`

	var messages []*models.Message
	err := r.db.SelectContext(ctx, &messages, query, chatID, limit, offset)
	if err != nil {
		return nil, err
	}

	return messages, nil
}
