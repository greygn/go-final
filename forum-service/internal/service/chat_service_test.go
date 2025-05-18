package service

import (
	"context"
	"testing"
	"time"

	"github.com/greygn/forum-service/internal/config"
	"github.com/greygn/forum-service/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockMessageRepository struct {
	mock.Mock
}

func (m *MockMessageRepository) Create(ctx context.Context, message *repository.Message) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *MockMessageRepository) List(ctx context.Context) ([]*repository.Message, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*repository.Message), args.Error(1)
}

func (m *MockMessageRepository) DeleteOld(ctx context.Context, olderThan time.Duration) error {
	args := m.Called(ctx, olderThan)
	return args.Error(0)
}

func TestChatService_SaveMessage(t *testing.T) {
	messageRepo := new(MockMessageRepository)
	logger, _ := zap.NewProduction()
	cfg := &config.Config{
		MessageTTL: time.Second * 20,
	}

	service := NewChatService(messageRepo, cfg, logger)

	t.Run("successful save", func(t *testing.T) {
		messageRepo.On("Create", mock.Anything, mock.AnythingOfType("*repository.Message")).Return(nil)

		err := service.SaveMessage(context.Background(), 1, "testuser", "Hello, World!")
		assert.NoError(t, err)

		messageRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		messageRepo.On("Create", mock.Anything, mock.AnythingOfType("*repository.Message")).Return(assert.AnError)

		err := service.SaveMessage(context.Background(), 1, "testuser", "Hello, World!")
		assert.Error(t, err)

		messageRepo.AssertExpectations(t)
	})
}

func TestChatService_GetMessages(t *testing.T) {
	messageRepo := new(MockMessageRepository)
	logger, _ := zap.NewProduction()
	cfg := &config.Config{
		MessageTTL: time.Second * 20,
	}

	service := NewChatService(messageRepo, cfg, logger)

	t.Run("successful get", func(t *testing.T) {
		messages := []*repository.Message{
			{
				ID:        1,
				UserID:    1,
				Username:  "testuser",
				Content:   "Hello, World!",
				CreatedAt: time.Now(),
			},
		}

		messageRepo.On("List", mock.Anything).Return(messages, nil)

		result, err := service.GetMessages(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, messages, result)

		messageRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		messageRepo.On("List", mock.Anything).Return(nil, assert.AnError)

		result, err := service.GetMessages(context.Background())
		assert.Error(t, err)
		assert.Nil(t, result)

		messageRepo.AssertExpectations(t)
	})
}
