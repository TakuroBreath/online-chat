package api

import (
	"context"
	"log"

	pb "chat.service/api/proto"
	"chat.service/internal/service/chat_service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ChatServiceHandler обрабатывает gRPC запросы к сервису чата
type ChatServiceHandler struct {
	pb.UnimplementedChatServiceServer
	chatService *chat_service.ChatService
}

// NewChatServiceHandler создает новый обработчик сервиса чата
func NewChatServiceHandler(chatService *chat_service.ChatService) *ChatServiceHandler {
	return &ChatServiceHandler{
		chatService: chatService,
	}
}

// getUserIDFromContext извлекает ID пользователя из контекста запроса
func getUserIDFromContext(ctx context.Context) (string, error) {
	// Получаем метаданные из контекста
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "метаданные не найдены")
	}

	// Получаем ID пользователя из метаданных
	userIDs := md.Get("user-id")
	if len(userIDs) == 0 {
		return "", status.Error(codes.Unauthenticated, "ID пользователя не найден")
	}

	return userIDs[0], nil
}

// CreateChat создает новый чат
func (h *ChatServiceHandler) CreateChat(ctx context.Context, req *pb.CreateChatRequest) (*pb.CreateChatResponse, error) {
	// Получаем ID пользователя из контекста
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Создаем чат
	chatID, err := h.chatService.CreateChat(ctx, req.Name, userID, req.ParticipantUserIds)
	if err != nil {
		log.Printf("Ошибка при создании чата: %v", err)
		return nil, status.Error(codes.Internal, "ошибка при создании чата")
	}

	return &pb.CreateChatResponse{
		ChatId: chatID,
	}, nil
}

// ConnectChat подключает пользователя к чату для получения сообщений
func (h *ChatServiceHandler) ConnectChat(req *pb.ConnectChatRequest, stream pb.ChatService_ConnectChatServer) error {
	// Получаем ID пользователя из контекста
	userID, err := getUserIDFromContext(stream.Context())
	if err != nil {
		return err
	}

	// Получаем последние сообщения чата
	messages, err := h.chatService.GetChatMessages(stream.Context(), req.ChatId, userID, 50, 0)
	if err != nil {
		log.Printf("Ошибка при получении сообщений чата: %v", err)
		return status.Error(codes.Internal, "ошибка при получении сообщений чата")
	}

	// Отправляем последние сообщения клиенту
	for _, msg := range messages {
		pbMsg := &pb.ChatMessage{
			MessageId: msg.ID,
			ChatId:    msg.ChatID,
			UserId:    msg.UserID,
			Username:  msg.Username,
			Text:      msg.Text,
			Timestamp: timestamppb.New(msg.CreatedAt),
		}

		if err := stream.Send(pbMsg); err != nil {
			log.Printf("Ошибка при отправке сообщения клиенту: %v", err)
			return status.Error(codes.Internal, "ошибка при отправке сообщения")
		}
	}

	// Подписываемся на обновления чата
	messageChan, subscriptionID, err := h.chatService.SubscribeToChat(stream.Context(), req.ChatId, userID)
	if err != nil {
		log.Printf("Ошибка при подписке на обновления чата: %v", err)
		return status.Error(codes.Internal, "ошибка при подписке на обновления чата")
	}

	// Обеспечиваем отписку при завершении соединения
	defer h.chatService.UnsubscribeFromChat(req.ChatId, subscriptionID)

	// Запускаем горутину для отправки новых сообщений клиенту
	for {
		select {
		case message, ok := <-messageChan:
			// Проверяем, не закрыт ли канал
			if !ok {
				return nil
			}

			// Конвертируем сообщение в protobuf формат
			pbMsg := &pb.ChatMessage{
				MessageId: message.ID,
				ChatId:    message.ChatID,
				UserId:    message.UserID,
				Username:  message.Username,
				Text:      message.Text,
				Timestamp: timestamppb.New(message.CreatedAt),
			}

			// Отправляем сообщение клиенту
			if err := stream.Send(pbMsg); err != nil {
				log.Printf("Ошибка при отправке сообщения клиенту: %v", err)
				return status.Error(codes.Internal, "ошибка при отправке сообщения")
			}

		case <-stream.Context().Done():
			// Соединение закрыто клиентом или сервером
			return nil
		}
	}
}

// SendMessage отправляет сообщение в чат
func (h *ChatServiceHandler) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error) {
	// Получаем ID пользователя из контекста
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Отправляем сообщение
	messageID, timestamp, err := h.chatService.SendMessage(ctx, req.ChatId, userID, req.Text)
	if err != nil {
		log.Printf("Ошибка при отправке сообщения: %v", err)
		switch err {
		case chat_service.ErrUserNotInChat:
			return nil, status.Error(codes.PermissionDenied, "пользователь не является участником чата")
		case chat_service.ErrChatNotFound:
			return nil, status.Error(codes.NotFound, "чат не найден")
		default:
			return nil, status.Error(codes.Internal, "ошибка при отправке сообщения")
		}
	}

	return &pb.SendMessageResponse{
		MessageId: messageID,
		Timestamp: timestamppb.New(timestamp),
	}, nil
}
