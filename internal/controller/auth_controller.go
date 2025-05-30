package controller

import (
	"github.com/bekbek22/auth_service/internal/service"

	pb "github.com/bekbek22/auth_service/api/proto"
)

type AuthController struct {
	pb.UnimplementedAuthServiceServer
	service *service.AuthService
}

func NewAuthController(s *service.AuthService) *AuthController {
	return &AuthController{
		service: s,
	}
}
