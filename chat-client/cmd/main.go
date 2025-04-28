package main

import (
	"log"

	"chat.client/cmd/root"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

	root.Execute()
}
