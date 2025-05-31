package handler

import (
	"context"

	pb "github.com/bekbek22/auth_service/api/proto"
	"github.com/bekbek22/auth_service/internal/middleware"
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

func (h *AuthHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	err := h.service.Register(ctx, req.Name, req.Email, req.Password)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "register failed: %v", err)
	}
	return &pb.RegisterResponse{Message: "Registration successful"}, nil
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

func (h *AuthHandler) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	err := h.service.Logout(ctx, req.AccessToken)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Logout failed: %v", err)
	}
	return &pb.LogoutResponse{Message: "Logged out successfully"}, nil
}

func (h *AuthHandler) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	//Extract & Validate JWT
	tokenStr, err := middleware.ExtractTokenFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "missing or invalid token: %v", err)
	}

	claims, err := middleware.ValidateJWT(tokenStr, h.service.Cfg.JWTSecret)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "token invalid: %v", err)
	}

	//Check role
	if role, ok := claims["role"].(string); !ok || role != "admin" {
		return nil, status.Errorf(codes.PermissionDenied, "admin access only")
	}

	// Query users
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}

	users, total, err := h.service.ListUsers(ctx, req.Name, req.Email, req.Page, req.Limit)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list users: %v", err)
	}

	var items []*pb.UserItem
	for _, u := range users {
		items = append(items, &pb.UserItem{
			Id:    u.ID.Hex(),
			Name:  u.Name,
			Email: u.Email,
			Role:  u.Role,
		})
	}

	return &pb.ListUsersResponse{
		Users: items,
		Total: total,
	}, nil
}

func (h *AuthHandler) GetProfile(ctx context.Context, req *pb.GetProfileRequest) (*pb.GetProfileResponse, error) {
	tokenStr, err := middleware.ExtractTokenFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "missing token")
	}

	claims, err := middleware.ValidateJWT(tokenStr, h.service.Cfg.JWTSecret)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token")
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return nil, status.Errorf(codes.Internal, "missing user_id in token")
	}

	user, err := h.service.GetProfile(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found")
	}

	return &pb.GetProfileResponse{
		Id:    user.ID.Hex(),
		Name:  user.Name,
		Email: user.Email,
		Role:  user.Role,
	}, nil
}

func (h *AuthHandler) UpdateProfile(ctx context.Context, req *pb.UpdateProfileRequest) (*pb.UpdateProfileResponse, error) {
	//Extract token & get user_id
	tokenStr, err := middleware.ExtractTokenFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "missing token")
	}

	claims, err := middleware.ValidateJWT(tokenStr, h.service.Cfg.JWTSecret)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token")
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return nil, status.Errorf(codes.Internal, "missing user_id in token")
	}

	err = h.service.UpdateProfile(ctx, userID, req.Name, req.Email)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "update failed: %v", err)
	}

	return &pb.UpdateProfileResponse{
		Message: "Profile updated successfully",
	}, nil
}

func (h *AuthHandler) DeleteProfile(ctx context.Context, req *pb.DeleteProfileRequest) (*pb.DeleteProfileResponse, error) {
	tokenStr, err := middleware.ExtractTokenFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "missing token")
	}

	claims, err := middleware.ValidateJWT(tokenStr, h.service.Cfg.JWTSecret)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token")
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return nil, status.Errorf(codes.Internal, "missing user_id in token")
	}

	err = h.service.DeleteProfile(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete profile: %v", err)
	}

	return &pb.DeleteProfileResponse{
		Message: "Profile deleted (soft delete)",
	}, nil
}

func (h *AuthHandler) RequestPasswordReset(ctx context.Context, req *pb.RequestPasswordResetRequest) (*pb.RequestPasswordResetResponse, error) {
	token, err := h.service.RequestPasswordReset(ctx, req.Email)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate reset token: %v", err)
	}
	return &pb.RequestPasswordResetResponse{
		ResetToken: token,
	}, nil
}

func (h *AuthHandler) ResetPassword(ctx context.Context, req *pb.ResetPasswordRequest) (*pb.ResetPasswordResponse, error) {
	err := h.service.ResetPassword(ctx, req.ResetToken, req.NewPassword)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "reset failed: %v", err)
	}
	return &pb.ResetPasswordResponse{
		Message: "Password reset successfully",
	}, nil
}
