package chat_client

import (
	"context"

	pb "chat.service/api/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ChatClient struct {
	chatClient pb.ChatServiceClient
	conn       *grpc.ClientConn
	token      string
}

func NewChatClient(chatServiceAddr string, token string) (*ChatClient, error) {
	// Создаем перехватчики для аутентификации
	authInterceptor := AuthInterceptor(token)
	streamAuthInterceptor := StreamAuthInterceptor(token)

	// Устанавливаем соединение с перехватчиками
	conn, err := grpc.Dial(
		chatServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(authInterceptor),
		grpc.WithStreamInterceptor(streamAuthInterceptor),
	)
	if err != nil {
		return nil, err
	}

	chatClient := pb.NewChatServiceClient(conn)

	return &ChatClient{
		chatClient: chatClient,
		conn:       conn,
		token:      token,
	}, nil
}

func (c *ChatClient) Close() error {
	return c.conn.Close()
}

func (c *ChatClient) CreateChat(name string) (string, error) {
	res, err := c.chatClient.CreateChat(context.Background(), &pb.CreateChatRequest{
		Name: name,
	})

	if err != nil {
		return "", err
	}

	return res.GetChatId(), nil
}

// ConnectToChat подключается к чату и возвращает стрим сообщений и функцию для обработки сообщений
func (c *ChatClient) ConnectToChat(ctx context.Context, chatID string) (pb.ChatService_ConnectChatClient, error) {
	// Создаем запрос на подключение к чату
	stream, err := c.chatClient.ConnectChat(ctx, &pb.ConnectChatRequest{
		ChatId: chatID,
	})

	if err != nil {
		return nil, err
	}

	return stream, nil
}

// ProcessChatMessages обрабатывает входящие сообщения из стрима
// messageHandler - функция, которая будет вызываться для каждого полученного сообщения
// errorHandler - функция, которая будет вызываться при возникновении ошибки
func (c *ChatClient) ProcessChatMessages(stream pb.ChatService_ConnectChatClient,
	messageHandler func(*pb.ChatMessage),
	errorHandler func(error)) {

	go func() {
		for {
			// Получаем сообщение из стрима
			message, err := stream.Recv()
			if err != nil {
				errorHandler(err)
				return
			}

			// Обрабатываем полученное сообщение
			messageHandler(message)
		}
	}()
}

func (c *ChatClient) SendMessage(chatID, text string) error {
	_, err := c.chatClient.SendMessage(context.Background(), &pb.SendMessageRequest{
		ChatId: chatID,
		Text:   text,
	})

	return err
}
