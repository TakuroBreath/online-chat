package chat_service

import (
	"context"
	"errors"
	"log"
	"time"

	"chat.service/internal/models"
	"chat.service/internal/repository"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	ErrChatNotFound   = errors.New("чат не найден")
	ErrUserNotInChat  = errors.New("пользователь не является участником чата")
	ErrUserNotFound   = errors.New("пользователь не найден")
	ErrInvalidChatID  = errors.New("некорректный ID чата")
	ErrInvalidUserID  = errors.New("некорректный ID пользователя")
	ErrInvalidMessage = errors.New("некорректное сообщение")
	ErrSubscription   = errors.New("ошибка подписки на обновления чата")
)

// ChatService предоставляет методы для работы с чатами
type ChatService struct {
	chatRepo    repository.ChatRepository
	messageRepo repository.MessageRepository
	authClient  AuthClient           // Клиент для взаимодействия с сервисом аутентификации
	subManager  *SubscriptionManager // Менеджер подписок для real-time обновлений
}

// AuthClient определяет интерфейс для взаимодействия с сервисом аутентификации
type AuthClient interface {
	// GetUserByID возвращает информацию о пользователе по ID
	GetUserByID(ctx context.Context, userID string) (string, error)
	// ValidateToken проверяет токен доступа и возвращает ID пользователя
	ValidateToken(ctx context.Context, token string) (string, error)
}

// NewChatService создает новый экземпляр сервиса чатов
func NewChatService(chatRepo repository.ChatRepository, messageRepo repository.MessageRepository, authClient AuthClient) *ChatService {
	return &ChatService{
		chatRepo:    chatRepo,
		messageRepo: messageRepo,
		authClient:  authClient,
		subManager:  NewSubscriptionManager(),
	}
}

// CreateChat создает новый чат и добавляет в него создателя и указанных участников
func (s *ChatService) CreateChat(ctx context.Context, name string, creatorID string, participantIDs []string) (string, error) {
	if creatorID == "" {
		return "", ErrInvalidUserID
	}

	// Создаем чат
	chat := &models.Chat{
		Name:        name,
		CreatedByID: creatorID,
	}

	chatID, err := s.chatRepo.CreateChat(ctx, chat)
	if err != nil {
		return "", err
	}

	// Добавляем создателя как участника
	err = s.chatRepo.AddParticipant(ctx, chatID, creatorID)
	if err != nil {
		return "", err
	}

	// Добавляем других участников
	for _, userID := range participantIDs {
		// Проверяем, что пользователь существует через сервис аутентификации
		_, err := s.authClient.GetUserByID(ctx, userID)
		if err != nil {
			// Пропускаем несуществующих пользователей
			continue
		}

		// Добавляем пользователя в чат
		_ = s.chatRepo.AddParticipant(ctx, chatID, userID)
	}

	return chatID, nil
}

// SendMessage отправляет сообщение в чат
func (s *ChatService) SendMessage(ctx context.Context, chatID, userID, text string) (string, time.Time, error) {
	if chatID == "" {
		log.Printf("Ошибка: пустой ID чата")
		return "", time.Time{}, ErrInvalidChatID
	}

	if userID == "" {
		log.Printf("Ошибка: пустой ID пользователя")
		return "", time.Time{}, ErrInvalidUserID
	}

	if text == "" {
		log.Printf("Ошибка: пустой текст сообщения")
		return "", time.Time{}, ErrInvalidMessage
	}

	// Проверяем, что пользователь является участником чата
	// isParticipant, err := s.chatRepo.CheckUserInChat(ctx, chatID, userID)
	// if err != nil {
	// 	log.Printf("Ошибка при проверке участия пользователя в чате: %v", err)
	// 	return "", time.Time{}, err
	// }

	// if !isParticipant {
	// 	log.Printf("Пользователь %s не является участником чата %s", userID, chatID)
	// 	return "", time.Time{}, ErrUserNotInChat
	// }

	// Получаем имя пользователя через сервис аутентификации
	username, err := s.authClient.GetUserByID(ctx, userID)
	if err != nil {
		log.Printf("Не удалось получить имя пользователя %s: %v, используем ID", userID, err)
		// Если не удалось получить имя пользователя, используем ID
		username = userID
	}

	// Создаем сообщение
	message := &models.Message{
		ChatID:   chatID,
		UserID:   userID,
		Username: username,
		Text:     text,
	}

	// Сохраняем сообщение
	messageID, err := s.messageRepo.SaveMessage(ctx, message)
	if err != nil {
		log.Printf("Ошибка при сохранении сообщения: %v", err)
		return "", time.Time{}, err
	}

	// Устанавливаем ID сообщения
	message.ID = messageID

	// Публикуем сообщение для всех подписчиков
	s.subManager.PublishMessage(chatID, message)
	log.Printf("Сообщение %s успешно отправлено в чат %s пользователем %s", messageID, chatID, userID)

	return messageID, message.CreatedAt, nil
}

// GetChatMessages возвращает сообщения чата
func (s *ChatService) GetChatMessages(ctx context.Context, chatID, userID string, limit, offset int) ([]*models.Message, error) {
	// Проверяем, что пользователь является участником чата
	// isParticipant, err := s.chatRepo.CheckUserInChat(ctx, chatID, userID)
	// if err != nil {
	// 	return nil, err
	// }

	// if !isParticipant {
	// 	return nil, ErrUserNotInChat
	// }

	// Получаем сообщения
	return s.messageRepo.GetChatMessages(ctx, chatID, limit, offset)
}

// ConvertMessageToProto конвертирует модель сообщения в protobuf формат
func ConvertMessageToProto(message *models.Message) *ChatMessage {
	return &ChatMessage{
		MessageId: message.ID,
		ChatId:    message.ChatID,
		UserId:    message.UserID,
		Username:  message.Username,
		Text:      message.Text,
		Timestamp: timestamppb.New(message.CreatedAt),
	}
}

// SubscribeToChat подписывает клиента на обновления чата
func (s *ChatService) SubscribeToChat(ctx context.Context, chatID, userID string) (<-chan *models.Message, string, error) {
	log.Printf("Попытка подписки пользователя %s на обновления чата %s", userID, chatID)

	// Проверяем, что пользователь является участником чата
	// isParticipant, err := s.chatRepo.CheckUserInChat(ctx, chatID, userID)
	// if err != nil {
	// 	log.Printf("Ошибка при проверке участия пользователя %s в чате %s: %v", userID, chatID, err)
	// 	return nil, "", err
	// }

	// if !isParticipant {
	// 	log.Printf("Пользователь %s не является участником чата %s", userID, chatID)
	// 	return nil, "", ErrUserNotInChat
	// }

	// Создаем подписку
	messageChan, subscriptionID := s.subManager.Subscribe(chatID)
	log.Printf("Пользователь %s успешно подписан на обновления чата %s, ID подписки: %s", userID, chatID, subscriptionID)

	return messageChan, subscriptionID, nil
}

// UnsubscribeFromChat отписывает клиента от обновлений чата
func (s *ChatService) UnsubscribeFromChat(chatID, subscriptionID string) {
	log.Printf("Отписка от обновлений чата %s, ID подписки: %s", chatID, subscriptionID)
	s.subManager.Unsubscribe(chatID, subscriptionID)
}

// ChatMessage представляет protobuf сообщение (для удобства конвертации)
type ChatMessage struct {
	MessageId string
	ChatId    string
	UserId    string
	Username  string
	Text      string
	Timestamp *timestamppb.Timestamp
}
