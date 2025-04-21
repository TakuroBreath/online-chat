package api

import (
	"context"
	"log"

	pb "auth.service/api/proto"
	"auth.service/internal/service"
	"auth.service/internal/service/auth_service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthServiceHandler struct {
	pb.UnimplementedAuthServiceServer
	authService service.AuthService
}

func NewAuthServiceHandler(authService service.AuthService) *AuthServiceHandler {
	return &AuthServiceHandler{
		authService: authService,
	}
}

func (h *AuthServiceHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	if req.Username == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "Username and password are required")
	}

	tokens, err := h.authService.Login(ctx, req.Username, req.Password)
	if err != nil {
		log.Printf("Error logging in: %v", err)
		switch err {
		case auth_service.ErrInvalidCredentials:
			return nil, status.Error(codes.Unauthenticated, "Invalid credentials")
		default:
			return nil, status.Error(codes.Internal, "Internal server error")
		}
	}

	return &pb.LoginResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		UserId:       tokens.UserID,
	}, nil
}

func (h *AuthServiceHandler) GetAccessToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.AccessTokenResponse, error) {
	if req.RefreshToken == "" {
		return nil, status.Error(codes.InvalidArgument, "Refresh token is required")
	}

	tokens, err := h.authService.RefreshTokens(ctx, req.RefreshToken)
	if err != nil {
		log.Printf("Error refreshing tokens: %v", err)
		switch err {
		case auth_service.ErrInvalidToken:
			return nil, status.Error(codes.Unauthenticated, "Invalid refresh token")
		case auth_service.ErrExpiredToken:
			return nil, status.Error(codes.Unauthenticated, "Refresh token has expired")
		case auth_service.ErrTokenNotFound:
			return nil, status.Error(codes.Unauthenticated, "Refresh token not found")
		default:
			return nil, status.Error(codes.Internal, "Internal server error")
		}
	}
	
	return &pb.AccessTokenResponse{
		AccessToken: tokens.AccessToken,
	}, nil
}
