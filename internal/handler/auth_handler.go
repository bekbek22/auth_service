package handler

import (
	"context"

	pb "github.com/bekbek22/auth_service/api/proto"
	"github.com/bekbek22/auth_service/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthHandler struct {
	pb.UnimplementedAuthServiceServer
	service *service.AuthService
}

func NewAuthHandler(s *service.AuthService) *AuthHandler {
	return &AuthHandler{service: s}
}

func (c *AuthHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	err := c.service.Register(ctx, req.Email, req.Password)
	if err != nil {
		return nil, err
	}
	return &pb.RegisterResponse{
		Message: "Registration successful",
	}, nil
}

func (c *AuthHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	token, err := c.service.Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Login failed: %v", err)
	}

	return &pb.LoginResponse{
		AccessToken: token,
	}, nil
}
