package user_client

import (
	"context"

	authpb "auth.service/api/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type UserClient struct {
	userClient   authpb.UserServiceClient
	authClient   authpb.AuthServiceClient
	accessClient authpb.AccessServiceClient
	conn         *grpc.ClientConn
}

func NewUserClient(authServiceAddr string) (*UserClient, error) {
	conn, err := grpc.Dial(authServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	accessClient := authpb.NewAccessServiceClient(conn)
	userClient := authpb.NewUserServiceClient(conn)
	authClient := authpb.NewAuthServiceClient(conn)

	return &UserClient{
		userClient:   userClient,
		authClient:   authClient,
		accessClient: accessClient,
		conn:         conn,
	}, nil
}

func (c *UserClient) Close() error {
	return c.conn.Close()
}

func (c *UserClient) Register(username, password string) error {
	_, err := c.userClient.CreateUser(context.Background(), &authpb.CreateUserRequest{
		Username: username,
		Password: password,
	})

	return err
}

func (c *UserClient) Login(username, password string) (string, error) {
	res, err := c.authClient.Login(context.Background(), &authpb.LoginRequest{
		Username: username,
		Password: password,
	})

	if err != nil {
		return "", err
	}

	return res.AccessToken, nil
}
