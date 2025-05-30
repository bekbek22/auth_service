package service

import (
	"github.com/bekbek22/auth_service/config"
	"github.com/bekbek22/auth_service/internal/repository"
)

type AuthService struct {
	repo *repository.UserRepository
	cfg  *config.Config
}

func NewAuthService(repo *repository.UserRepository, cfg *config.Config) *AuthService {
	return &AuthService{
		repo: repo,
		cfg:  cfg,
	}
}
