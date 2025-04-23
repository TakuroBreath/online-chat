package chat_service

import (
	"sync"

	"chat.service/internal/models"
	"github.com/google/uuid"
)

// SubscriptionManager управляет подписками на обновления чатов
type SubscriptionManager struct {
	subscriptions map[string]map[string]chan *models.Message // map[chatID]map[subscriptionID]channel
	mutex         sync.RWMutex
}

// NewSubscriptionManager создает новый менеджер подписок
func NewSubscriptionManager() *SubscriptionManager {
	return &SubscriptionManager{
		subscriptions: make(map[string]map[string]chan *models.Message),
	}
}

// Subscribe создает новую подписку на обновления чата
// Возвращает канал для получения сообщений и ID подписки
func (m *SubscriptionManager) Subscribe(chatID string) (chan *models.Message, string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Создаем канал для сообщений с буфером
	messageChan := make(chan *models.Message, 100)

	// Генерируем уникальный ID подписки
	subscriptionID := generateSubscriptionID()

	// Проверяем, существует ли уже карта подписок для этого чата
	if _, ok := m.subscriptions[chatID]; !ok {
		m.subscriptions[chatID] = make(map[string]chan *models.Message)
	}

	// Добавляем подписку
	m.subscriptions[chatID][subscriptionID] = messageChan

	return messageChan, subscriptionID
}

// Unsubscribe отменяет подписку
func (m *SubscriptionManager) Unsubscribe(chatID, subscriptionID string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Проверяем, существует ли карта подписок для этого чата
	if chatSubscriptions, ok := m.subscriptions[chatID]; ok {
		// Проверяем, существует ли подписка
		if messageChan, ok := chatSubscriptions[subscriptionID]; ok {
			// Закрываем канал
			close(messageChan)
			// Удаляем подписку
			delete(chatSubscriptions, subscriptionID)
		}

		// Если подписок на чат больше нет, удаляем карту
		if len(chatSubscriptions) == 0 {
			delete(m.subscriptions, chatID)
		}
	}
}

// PublishMessage отправляет сообщение всем подписчикам чата
func (m *SubscriptionManager) PublishMessage(chatID string, message *models.Message) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Проверяем, есть ли подписчики для этого чата
	if chatSubscriptions, ok := m.subscriptions[chatID]; ok {
		// Отправляем сообщение всем подписчикам
		for _, messageChan := range chatSubscriptions {
			// Используем неблокирующую отправку, чтобы не зависать, если канал полон
			select {
			case messageChan <- message:
				// Сообщение успешно отправлено
			default:
				// Канал полон, пропускаем отправку
			}
		}
	}
}

// generateSubscriptionID генерирует уникальный ID подписки
func generateSubscriptionID() string {
	// Для простоты используем текущее время в наносекундах
	return uuid.New().String()
}
