package api

import (
	pb "auth.service/api/proto"
	"auth.service/internal/service"
)

type AuthUserHandler struct {
	pb.UnimplementedAuthServiceServer
	authService service.AuthService
}
