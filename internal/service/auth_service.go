package service

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/bekbek22/auth_service/internal/middleware"
	"github.com/bekbek22/auth_service/internal/model"
	"github.com/bekbek22/auth_service/internal/repository"
	"github.com/bekbek22/auth_service/internal/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/bekbek22/auth_service/config"
)

type IAuthService interface {
	Register(ctx context.Context, name, email, password string) error
	Login(ctx context.Context, email, password string) (string, error)
	Logout(ctx context.Context, token string) error
	ListUsers(ctx context.Context, name, email string, page, limit int32) ([]model.User, int32, error)
	GetProfile(ctx context.Context, userID string) (*model.User, error)
	UpdateProfile(ctx context.Context, userID, name, email string) error
	DeleteProfile(ctx context.Context, userID string) error
	RequestPasswordReset(ctx context.Context, email string) (string, error)
	ResetPassword(ctx context.Context, token, newPassword string) error
}

type AuthService struct {
	repo              *repository.UserRepository
	tokenRepo         *repository.TokenRepository
	passwordResetRepo *repository.PasswordResetRepository
	Cfg               *config.Config
	rateLimiter       *middleware.RateLimiter
}

func NewAuthService(userRepo *repository.UserRepository, tokenRepo *repository.TokenRepository, passwordResetRepo *repository.PasswordResetRepository, cfg *config.Config) *AuthService {
	rl := middleware.NewRateLimiter(5, 60)
	return &AuthService{
		repo:              userRepo,
		tokenRepo:         tokenRepo,
		passwordResetRepo: passwordResetRepo,
		Cfg:               cfg,
		rateLimiter:       rl,
	}
}

func isValidEmail(email string) bool {
	regex := `^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`
	re := regexp.MustCompile(regex)
	return re.MatchString(email)
}

func isStrongPassword(pw string) bool {
	return len(pw) >= 8
}

func (s *AuthService) Register(ctx context.Context, name, email, password string) error {
	//Check name format
	if strings.TrimSpace(name) == "" {
		return errors.New("name is required")
	}

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
		Name:     name,
		Email:    email,
		Role:     "user", // default
		Password: hashedPassword,
	}
	return s.repo.CreateUser(ctx, user)
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	if !s.rateLimiter.Allow(email) {
		return "", errors.New("too many login attempts, please wait")
	}

	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return "", errors.New("user not found")
	}

	if !utils.CheckPasswordHash(password, user.Password) {
		return "", errors.New("invalid password")
	}

	token, err := utils.GenerateJWT(user.ID.Hex(), user.Role, s.Cfg.JWTSecret)
	if err != nil {
		return "", errors.New("failed to generate token")
	}

	return token, nil
}

func (s *AuthService) Logout(ctx context.Context, token string) error {
	parsed, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return []byte(s.Cfg.JWTSecret), nil
	})
	if err != nil || !parsed.Valid {
		return errors.New("invalid token")
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return errors.New("invalid claims")
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return errors.New("missing exp")
	}

	return s.tokenRepo.BlacklistToken(ctx, token, int64(exp))
}

func (s *AuthService) ListUsers(ctx context.Context, name, email string, page, limit int32) ([]model.User, int32, error) {
	return s.repo.FindUsers(ctx, name, email, page, limit)
}

func (s *AuthService) GetProfile(ctx context.Context, userID string) (*model.User, error) {
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}
	return s.repo.FindByID(ctx, oid)
}

func (s *AuthService) UpdateProfile(ctx context.Context, userID, name, email string) error {
	if strings.TrimSpace(name) == "" || strings.TrimSpace(email) == "" {
		return errors.New("name and email must not be empty")
	}

	if !isValidEmail(email) {
		return errors.New("invalid email format")
	}

	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("invalid user ID")
	}

	updates := bson.M{
		"name":  name,
		"email": email,
	}

	return s.repo.UpdateUserByID(ctx, oid, updates)
}

func (s *AuthService) DeleteProfile(ctx context.Context, userID string) error {
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("invalid user ID")
	}
	return s.repo.SoftDeleteUserByID(ctx, oid)
}

func (s *AuthService) RequestPasswordReset(ctx context.Context, email string) (string, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil || user.IsDeleted {
		return "", errors.New("user not found")
	}

	token := uuid.NewString()
	exp := time.Now().Add(15 * time.Minute).Unix()

	err = s.passwordResetRepo.SaveToken(ctx, email, token, exp)
	if err != nil {
		return "", errors.New("failed to save reset token")
	}
	return token, nil
}

func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	if len(newPassword) < 8 {
		return errors.New("password too short")
	}

	email, err := s.passwordResetRepo.GetEmailByToken(ctx, token)
	if err != nil {
		return errors.New("invalid or expired token")
	}

	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return errors.New("user not found")
	}

	hashed, err := utils.HashPassword(newPassword)
	if err != nil {
		return errors.New("failed to hash password")
	}

	err = s.repo.UpdateUserByID(ctx, user.ID, bson.M{"password": hashed})
	if err != nil {
		return err
	}

	return s.passwordResetRepo.DeleteToken(ctx, token)
}
