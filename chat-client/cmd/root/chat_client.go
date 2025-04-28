package root

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"chat.client/internal/chat_client"
	pb "chat.service/api/proto"
	"github.com/spf13/cobra"
)

var (
	chatID   string
	chatName string
	token    string
)

var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "connect to a chat",
	Long: `connect to a chat with the given chat ID.
	It is written in Go and uses the Cobra library for command line parsing.`,
	Run: func(cmd *cobra.Command, args []string) {
		var chatServiceAddr string

		if chatID == "" && chatName == "" {
			cmd.Help()
			return
		}

		if token == "" {
			cmd.Println("You must provide a token. Use login command to get a token.")
			return
		}

		if addr, ok := os.LookupEnv("CHAT_SERVICE_ADDR"); !ok {
			cmd.Println("CHAT_SERVICE_ADDR environment variable is not set")
			return
		} else {
			chatServiceAddr = addr
		}

		client, err := chat_client.NewChatClient(chatServiceAddr, token)
		if err != nil {
			cmd.Printf("Failed to create chat client: %v\n", err)
			return
		}
		defer client.Close()

		// Если указано имя чата, но не указан ID, создаем новый чат
		if chatID == "" && chatName != "" {
			chatID, err = client.CreateChat(chatName)
			if err != nil {
				cmd.Printf("Failed to create chat: %v\n", err)
				return
			}
			cmd.Printf("Created new chat with ID: %s\n", chatID)
		}

		// Создаем контекст с возможностью отмены
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Подключаемся к чату
		cmd.Printf("Connecting to chat with ID: %s\n", chatID)
		stream, err := client.ConnectToChat(ctx, chatID)
		if err != nil {
			cmd.Printf("Failed to connect to chat: %v\n", err)
			return
		}

		// Обработка сигналов для корректного завершения
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		// Обработка входящих сообщений
		client.ProcessChatMessages(
			stream,
			// Обработчик сообщений
			func(message *pb.ChatMessage) {
				fmt.Printf("%s: %s\n", message.GetUsername(), message.GetText())
			},
			// Обработчик ошибок
			func(err error) {
				fmt.Printf("Error receiving message: %v\n", err)
				cancel()
			},
		)

		cmd.Println("Connected to chat. Type your messages and press Enter to send. Press Ctrl+C to exit.")

		// Чтение сообщений от пользователя и отправка их в чат
		go func() {
			reader := bufio.NewReader(os.Stdin)
			for {
				input, err := reader.ReadString('\n')
				if err != nil {
					fmt.Printf("Error reading input: %v\n", err)
					continue
				}

				// Удаляем символ новой строки в конце
				input = strings.TrimSpace(input)

				if input != "" {
					err = client.SendMessage(chatID, input)
					if err != nil {
						fmt.Printf("Error sending message: %v\n", err)
					}
				}
			}
		}()

		// Ожидание сигнала завершения
		<-sigCh
		cmd.Println("\nDisconnecting from chat...")
	},
}

var createChatCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new chat",
	Long: `create a new chat with the given name.
	It is written in Go and uses the Cobra library for command line parsing.`,
	Run: func(cmd *cobra.Command, args []string) {
		var chatServiceAddr string

		if chatName == "" {
			cmd.Help()
			return
		}

		if token == "" {
			cmd.Println("You must provide a token. Use login command to get a token.")
			return
		}

		if addr, ok := os.LookupEnv("CHAT_SERVICE_ADDR"); !ok {
			cmd.Println("CHAT_SERVICE_ADDR environment variable is not set")
			return
		} else {
			chatServiceAddr = addr
		}

		client, err := chat_client.NewChatClient(chatServiceAddr, token)
		if err != nil {
			cmd.Printf("Failed to create chat client: %v\n", err)
			return
		}
		defer client.Close()

		chatID, err := client.CreateChat(chatName)
		if err != nil {
			cmd.Printf("Failed to create chat: %v\n", err)
			return
		}

		cmd.Printf("Chat created successfully with ID: %s\n", chatID)
	},
}

func init() {
	connectCmd.Flags().StringVarP(&chatID, "id", "i", "", "chat ID")
	connectCmd.Flags().StringVarP(&chatName, "name", "n", "", "chat name")
	connectCmd.Flags().StringVarP(&token, "token", "t", "", "auth token")

	createChatCmd.Flags().StringVarP(&chatName, "name", "n", "", "chat name")
	createChatCmd.Flags().StringVarP(&token, "token", "t", "", "auth token")
}
