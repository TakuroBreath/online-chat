package root

import (
	"fmt"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "chatik",
	Short: "chat is a simple chat application",
	Long: `chat is a simple chat application that allows you to chat with other users.
	It is written in Go and uses the Cobra library for command line parsing.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Добро пожаловать в чат-приложение!")
		fmt.Println("Для начала работы используйте одну из доступных команд:")
		cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(registerCmd)
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(connectCmd)
	rootCmd.AddCommand(createChatCmd)
}

func Execute() error {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println("Ошибка при выполнении команды:", err)
	}
	return err
}
