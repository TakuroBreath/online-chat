package root

import (
	"os"

	"chat.client/internal/user_client"
	"github.com/spf13/cobra"
)

var (
	username string
	password string
)

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "register a new user",
	Long: `register a new user with the given username and password.
	It is written in Go and uses the Cobra library for command line parsing.`,
	Run: func(cmd *cobra.Command, args []string) {
		var authServiceAddr string

		if username == "" || password == "" {
			cmd.Help()
			return
		}

		if addr, ok := os.LookupEnv("CHAT_AUTH_SERVICE_ADDR"); !ok {
			cmd.Help()
			return
		} else {
			authServiceAddr = addr
		}

		client, err := user_client.NewUserClient(authServiceAddr)
		if err != nil {
			cmd.Printf("Ошибка при подключении к серверу аутентификации: %v\n", err)
			return
		}

		err = client.Register(username, password)
		if err != nil {
			cmd.Printf("Ошибка при регистрации пользователя: %v\n", err)
			return
		}

		err = client.Close()
		if err != nil {
			cmd.Help()
			return
		}

		cmd.Println("User registered successfully")
	},
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "login a user",
	Long: `login a user with the given username and password.
	It is written in Go and uses the Cobra library for command line parsing.`,
	Run: func(cmd *cobra.Command, args []string) {
		var authServiceAddr string
		var token string

		if username == "" || password == "" {
			cmd.Help()
			return
		}

		if addr, ok := os.LookupEnv("CHAT_AUTH_SERVICE_ADDR"); !ok {
			cmd.Help()
			return
		} else {
			authServiceAddr = addr
		}

		client, err := user_client.NewUserClient(authServiceAddr)
		if err != nil {
			cmd.Printf("Ошибка при подключении к серверу аутентификации: %v\n", err)
			return
		}

		token, err = client.Login(username, password)
		if err != nil {
			cmd.Printf("Ошибка при входе в систему: %v\n", err)
			return
		}

		err = client.Close()
		if err != nil {
			cmd.Help()
			return
		}

		cmd.Printf("User logged in successfully, token: %s\n", token)
	},
}

func init() {
	registerCmd.Flags().StringVarP(&username, "username", "u", "", "username")
	registerCmd.Flags().StringVarP(&password, "password", "p", "", "password")

	loginCmd.Flags().StringVarP(&username, "username", "u", "", "username")
	loginCmd.Flags().StringVarP(&password, "password", "p", "", "password")
}
