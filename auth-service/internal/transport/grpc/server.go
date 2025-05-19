// Package grpc provides the gRPC server implementation for the auth service.
//
// @title Auth Service API
// @version 1.0
// @description Authentication service for the forum application.
// @host localhost:50051
// @BasePath /auth
package grpc

import (
	"context"

	"auth-service/internal/service"

	authv1 "protos/auth/v1"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	*grpc.Server
}

func NewServer() *Server {
	return &Server{
		Server: grpc.NewServer(),
	}
}

// AuthServer implements the gRPC auth service.
type AuthServer struct {
	authv1.UnimplementedAuthServiceServer
	authService *service.AuthService
	logger      *zap.Logger
}

// NewAuthServer creates a new instance of AuthServer.
func NewAuthServer(authService *service.AuthService, logger *zap.Logger) *AuthServer {
	return &AuthServer{
		authService: authService,
		logger:      logger,
	}
}

// Register handles user registration.
//
// @Summary Register a new user
// @Description Register a new user with the provided credentials
// @Tags auth
// @Accept json
// @Produce json
// @Param request body authv1.RegisterRequest true "Registration request"
// @Success 200 {object} authv1.RegisterResponse
// @Failure 400 {object} status.Status "Invalid request"
// @Failure 500 {object} status.Status "Internal server error"
// @Router /register [post]
func (s *AuthServer) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	if err := s.authService.Register(ctx, req.GetUsername(), req.GetEmail(), req.GetPassword()); err != nil {
		s.logger.Error("failed to register user", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to register user")
	}

	user, err := s.authService.GetUserByUsername(ctx, req.GetUsername())
	if err != nil {
		s.logger.Error("failed to get user after registration", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get user")
	}

	tokens, err := s.authService.Login(ctx, req.GetUsername(), req.GetPassword())
	if err != nil {
		s.logger.Error("failed to login after registration", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to login")
	}

	return &authv1.RegisterResponse{
		UserId:       user.ID,
		Username:     user.Username,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

// Login handles user authentication.
//
// @Summary Authenticate a user
// @Description Authenticate a user and return access and refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body authv1.LoginRequest true "Login request"
// @Success 200 {object} authv1.LoginResponse
// @Failure 401 {object} status.Status "Invalid credentials"
// @Failure 500 {object} status.Status "Internal server error"
// @Router /login [post]
func (s *AuthServer) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	tokens, err := s.authService.Login(ctx, req.GetUsername(), req.GetPassword())
	if err != nil {
		s.logger.Error("failed to login", zap.Error(err))
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}

	user, err := s.authService.GetUserByUsername(ctx, req.GetUsername())
	if err != nil {
		s.logger.Error("failed to get user after login", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get user")
	}

	return &authv1.LoginResponse{
		UserId:       user.ID,
		Username:     user.Username,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

// IsAdmin checks if a user is an admin.
//
// @Summary Check if a user is an admin
// @Description Check if the user with the provided ID has admin privileges
// @Tags auth
// @Accept json
// @Produce json
// @Param request body authv1.IsAdminRequest true "Admin check request"
// @Success 200 {object} authv1.IsAdminResponse
// @Router /is-admin [post]
func (s *AuthServer) IsAdmin(ctx context.Context, req *authv1.IsAdminRequest) (*authv1.IsAdminResponse, error) {
	// For simplicity, let's consider user with ID 1 as admin
	return &authv1.IsAdminResponse{
		IsAdmin: req.GetUserId() == "1",
	}, nil
}

// RegisterGRPC registers the auth server with the gRPC server.
func (s *AuthServer) RegisterGRPC(grpcServer *grpc.Server) {
	authv1.RegisterAuthServiceServer(grpcServer, s)
}

func (s *AuthServer) RefreshToken(ctx context.Context, req *authv1.RefreshTokenRequest) (*authv1.RefreshTokenResponse, error) {
	tokens, err := s.authService.RefreshTokens(ctx, req.GetRefreshToken())
	if err != nil {
		s.logger.Error("failed to refresh token", zap.Error(err))
		return nil, status.Error(codes.Unauthenticated, "invalid refresh token")
	}

	return &authv1.RefreshTokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func (s *AuthServer) ValidateToken(ctx context.Context, req *authv1.ValidateTokenRequest) (*authv1.ValidateTokenResponse, error) {
	userID, err := s.authService.ValidateToken(ctx, req.GetAccessToken())
	if err != nil {
		s.logger.Error("failed to validate token", zap.Error(err))
		return &authv1.ValidateTokenResponse{IsValid: false}, nil
	}

	user, err := s.authService.GetUserByID(ctx, userID)
	if err != nil {
		s.logger.Error("failed to get user from token", zap.Error(err))
		return &authv1.ValidateTokenResponse{IsValid: false}, nil
	}

	return &authv1.ValidateTokenResponse{
		UserId:   user.ID,
		Username: user.Username,
		IsValid:  true,
	}, nil
}
