package service

import (
	"context"
	"errors"
	"regexp"

	"github.com/bekbek22/auth_service/internal/model"
	"github.com/bekbek22/auth_service/internal/repository"
	"github.com/bekbek22/auth_service/internal/utils"

	"github.com/bekbek22/auth_service/config"
)

type AuthService struct {
	repo *repository.UserRepository
	cfg  *config.Config
}

func NewAuthService(repo *repository.UserRepository, cfg *config.Config) *AuthService {
	return &AuthService{repo: repo, cfg: cfg}
}

func isValidEmail(email string) bool {
	regex := `^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`
	re := regexp.MustCompile(regex)
	return re.MatchString(email)
}

func isStrongPassword(pw string) bool {
	return len(pw) >= 8
}

func (s *AuthService) Register(ctx context.Context, email, password string) error {
	// Check email format
	if !isValidEmail(email) {
		return errors.New("invalid email format")
	}

	// Check password strength
	if !isStrongPassword(password) {
		return errors.New("password too weak (min 8 characters, mix letters & numbers)")
	}

	// Check if email already exists
	existingUser, _ := s.repo.FindByEmail(ctx, email)
	if existingUser != nil {
		return errors.New("email already registered")
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return errors.New("failed to hash password")
	}

	// Create user
	user := &model.User{
		Email:    email,
		Password: hashedPassword,
	}
	return s.repo.CreateUser(ctx, user)
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return "", errors.New("user not found")
	}

	if !utils.CheckPasswordHash(password, user.Password) {
		return "", errors.New("invalid password")
	}

	token, err := utils.GenerateJWT(user.ID.Hex(), s.cfg.JWTSecret)
	if err != nil {
		return "", errors.New("failed to generate token")
	}

	return token, nil
}
