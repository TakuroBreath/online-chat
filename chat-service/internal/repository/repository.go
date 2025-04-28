package repository

import (
	"context"

	"chat.service/internal/models"
)

// ChatRepository определяет интерфейс для работы с чатами
type ChatRepository interface {
	// CreateChat создает новый чат
	CreateChat(ctx context.Context, chat *models.Chat) (string, error)
	// AddParticipant добавляет участника в чат
	AddParticipant(ctx context.Context, chatID, userID string) error
	// GetChatByID возвращает чат по ID
	GetChatByID(ctx context.Context, chatID string) (*models.Chat, error)
	// GetChatParticipants возвращает список участников чата
	GetChatParticipants(ctx context.Context, chatID string) ([]string, error)
	// CheckUserInChat проверяет, является ли пользователь участником чата
	CheckUserInChat(ctx context.Context, chatID, userID string) (bool, error)
}

// MessageRepository определяет интерфейс для работы с сообщениями
type MessageRepository interface {
	// SaveMessage сохраняет сообщение в базе данных
	SaveMessage(ctx context.Context, message *models.Message) (string, error)
	// GetChatMessages возвращает сообщения чата
	GetChatMessages(ctx context.Context, chatID string, limit, offset int) ([]*models.Message, error)
}
