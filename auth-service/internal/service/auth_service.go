package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/greygn/auth-service/internal/config"
	"github.com/greygn/auth-service/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo  repository.UserRepository
	tokenRepo repository.TokenRepository
	config    *config.Config
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

func NewAuthService(userRepo repository.UserRepository, tokenRepo repository.TokenRepository, config *config.Config) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		config:    config,
	}
}

func (s *AuthService) Register(ctx context.Context, username, email, password string) error {
	// Check if user already exists
	if _, err := s.userRepo.GetByUsername(ctx, username); err == nil {
		return errors.New("username already taken")
	}
	if _, err := s.userRepo.GetByEmail(ctx, email); err == nil {
		return errors.New("email already registered")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &repository.User{
		Username:     username,
		Email:        email,
		PasswordHash: string(hashedPassword),
	}

	return s.userRepo.Create(ctx, user)
}

func (s *AuthService) Login(ctx context.Context, username, password string) (*TokenPair, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Generate token pair
	return s.generateTokenPair(ctx, user.ID)
}

func (s *AuthService) RefreshTokens(ctx context.Context, refreshToken string) (*TokenPair, error) {
	// Validate refresh token
	token, err := s.tokenRepo.Get(ctx, refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	if time.Now().After(token.ExpiresAt) {
		s.tokenRepo.Delete(ctx, refreshToken)
		return nil, errors.New("refresh token expired")
	}

	// Delete old refresh token
	if err := s.tokenRepo.Delete(ctx, refreshToken); err != nil {
		return nil, err
	}

	// Generate new token pair
	return s.generateTokenPair(ctx, token.UserID)
}

func (s *AuthService) generateTokenPair(ctx context.Context, userID string) (*TokenPair, error) {
	// Generate access token
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(s.config.AccessTokenTTL).Unix(),
	})

	accessTokenString, err := accessToken.SignedString([]byte(s.config.JWTSecretKey))
	if err != nil {
		return nil, err
	}

	// Generate refresh token
	refreshTokenBytes := make([]byte, 32)
	if _, err := rand.Read(refreshTokenBytes); err != nil {
		return nil, err
	}
	refreshTokenString := base64.URLEncoding.EncodeToString(refreshTokenBytes)

	// Store refresh token
	refreshTokenEntity := &repository.RefreshToken{
		UserID:    userID,
		Token:     refreshTokenString,
		ExpiresAt: time.Now().Add(s.config.RefreshTokenTTL),
	}

	if err := s.tokenRepo.Create(ctx, refreshTokenEntity); err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
	}, nil
}

func (s *AuthService) ValidateToken(ctx context.Context, tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.config.JWTSecretKey), nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, ok := claims["user_id"].(string)
		if !ok {
			return "", errors.New("invalid user id in token")
		}
		return userID, nil
	}

	return "", errors.New("invalid token")
}

func (s *AuthService) GetUserByUsername(ctx context.Context, username string) (*repository.User, error) {
	return s.userRepo.GetByUsername(ctx, username)
}

func (s *AuthService) GetUserByID(ctx context.Context, userID string) (*repository.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	return s.tokenRepo.Delete(ctx, refreshToken)
}

func (s *AuthService) LogoutAll(ctx context.Context, userID string) error {
	return s.tokenRepo.DeleteAllForUser(ctx, userID)
}
