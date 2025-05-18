package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/greygn/auth-service/internal/config"
	"github.com/greygn/auth-service/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *repository.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*repository.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*repository.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*repository.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.User), args.Error(1)
}

type MockTokenRepository struct {
	mock.Mock
}

func (m *MockTokenRepository) Create(ctx context.Context, token *repository.RefreshToken) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockTokenRepository) Get(ctx context.Context, token string) (*repository.RefreshToken, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.RefreshToken), args.Error(1)
}

func (m *MockTokenRepository) Delete(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockTokenRepository) DeleteAllForUser(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func TestAuthService_Register(t *testing.T) {
	userRepo := new(MockUserRepository)
	tokenRepo := new(MockTokenRepository)
	cfg := &config.Config{
		AccessTokenTTL:  time.Minute * 15,
		RefreshTokenTTL: time.Hour * 24 * 7,
		JWTSecretKey:    "test-secret",
	}

	service := NewAuthService(userRepo, tokenRepo, cfg)

	t.Run("successful registration", func(t *testing.T) {
		userRepo.On("GetByUsername", mock.Anything, "testuser").Return(nil, repository.ErrNotFound)
		userRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, repository.ErrNotFound)
		userRepo.On("Create", mock.Anything, mock.AnythingOfType("*repository.User")).Return(nil)

		err := service.Register(context.Background(), "testuser", "test@example.com", "password123")
		assert.NoError(t, err)

		userRepo.AssertExpectations(t)
	})

	t.Run("username already exists", func(t *testing.T) {
		userID := uuid.New().String()
		existingUser := &repository.User{ID: userID, Username: "testuser"}
		userRepo.On("GetByUsername", mock.Anything, "testuser").Return(existingUser, nil)

		err := service.Register(context.Background(), "testuser", "test@example.com", "password123")
		assert.Error(t, err)
		assert.Equal(t, "username already taken", err.Error())

		userRepo.AssertExpectations(t)
	})
}

func TestAuthService_Login(t *testing.T) {
	userRepo := new(MockUserRepository)
	tokenRepo := new(MockTokenRepository)
	cfg := &config.Config{
		AccessTokenTTL:  time.Minute * 15,
		RefreshTokenTTL: time.Hour * 24 * 7,
		JWTSecretKey:    "test-secret",
	}

	service := NewAuthService(userRepo, tokenRepo, cfg)

	t.Run("successful login", func(t *testing.T) {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		userID := uuid.New().String()
		user := &repository.User{
			ID:           userID,
			Username:     "testuser",
			PasswordHash: string(hashedPassword),
		}

		userRepo.On("GetByUsername", mock.Anything, "testuser").Return(user, nil)
		tokenRepo.On("Create", mock.Anything, mock.AnythingOfType("*repository.RefreshToken")).Return(nil)

		tokens, err := service.Login(context.Background(), "testuser", "password123")
		assert.NoError(t, err)
		assert.NotEmpty(t, tokens.AccessToken)
		assert.NotEmpty(t, tokens.RefreshToken)

		userRepo.AssertExpectations(t)
		tokenRepo.AssertExpectations(t)
	})

	t.Run("invalid credentials", func(t *testing.T) {
		userRepo.On("GetByUsername", mock.Anything, "testuser").Return(nil, repository.ErrNotFound)

		tokens, err := service.Login(context.Background(), "testuser", "password123")
		assert.Error(t, err)
		assert.Nil(t, tokens)
		assert.Equal(t, "invalid credentials", err.Error())

		userRepo.AssertExpectations(t)
	})
}
