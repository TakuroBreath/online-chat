module chat.client

go 1.24.2

require (
	auth.service v0.0.0
	chat.service v0.0.0
	github.com/spf13/cobra v1.9.1
)

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/joho/godotenv v1.5.1
	github.com/spf13/pflag v1.0.6 // indirect
	golang.org/x/net v0.35.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
	golang.org/x/text v0.24.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250218202821-56aae31c358a // indirect
	google.golang.org/grpc v1.72.0
	google.golang.org/protobuf v1.36.6 // indirect
)

replace auth.service => ../auth-service

replace chat.service => ../chat-service
