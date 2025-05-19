package service

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/greygn/forum-service/internal/config"
	"github.com/greygn/forum-service/internal/repository"
	"go.uber.org/zap"
)

type Message struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type Client struct {
	Conn     *websocket.Conn
	Send     chan []byte
	UserID   string
	Username string
}

type ChatService struct {
	messageRepo repository.MessageRepository
	config      *config.Config
	logger      *zap.Logger
	clients     map[*Client]bool
	broadcast   chan []byte
	register    chan *Client
	unregister  chan *Client
	mu          sync.RWMutex
}

func NewChatService(messageRepo repository.MessageRepository, config *config.Config, logger *zap.Logger) *ChatService {
	return &ChatService{
		messageRepo: messageRepo,
		config:      config,
		logger:      logger,
		clients:     make(map[*Client]bool),
		broadcast:   make(chan []byte),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
	}
}

func (s *ChatService) Run() {
	for {
		select {
		case client := <-s.register:
			s.mu.Lock()
			s.clients[client] = true
			s.mu.Unlock()
		case client := <-s.unregister:
			s.mu.Lock()
			if _, ok := s.clients[client]; ok {
				delete(s.clients, client)
				close(client.Send)
			}
			s.mu.Unlock()
		case message := <-s.broadcast:
			s.mu.RLock()
			for client := range s.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(s.clients, client)
				}
			}
			s.mu.RUnlock()
		}
	}
}

func (s *ChatService) Register(client *Client) {
	s.register <- client
}

func (s *ChatService) Unregister(client *Client) {
	s.unregister <- client
}

func (s *ChatService) Broadcast(message []byte) {
	s.broadcast <- message
}

func (s *ChatService) SaveMessage(ctx context.Context, userID string, username string, content string) error {
	if content == "" {
		return errors.New("content is required")
	}

	message := &repository.Message{
		UserID:    userID,
		Username:  username,
		Content:   content,
		CreatedAt: time.Now(),
	}

	if err := s.messageRepo.Create(ctx, message); err != nil {
		return err
	}

	// Broadcast message to all connected clients
	jsonMessage, err := json.Marshal(message)
	if err != nil {
		s.logger.Error("failed to marshal message", zap.Error(err))
		return err
	}

	s.Broadcast(jsonMessage)
	return nil
}

func (s *ChatService) GetMessages(ctx context.Context) ([]repository.Message, error) {
	return s.messageRepo.GetAll(ctx)
}

func (s *ChatService) GetMessage(ctx context.Context, messageID string) (*repository.Message, error) {
	return s.messageRepo.GetByID(ctx, messageID)
}

func (s *ChatService) UpdateMessage(ctx context.Context, messageID string, userID string, content string) error {
	if content == "" {
		return errors.New("content is required")
	}

	message, err := s.messageRepo.GetByID(ctx, messageID)
	if err != nil {
		return err
	}

	if message.UserID != userID {
		return errors.New("unauthorized: only message owner can update")
	}

	message.Content = content
	return s.messageRepo.Update(ctx, message)
}

func (s *ChatService) DeleteMessage(ctx context.Context, messageID string, userID string) error {
	message, err := s.messageRepo.GetByID(ctx, messageID)
	if err != nil {
		return err
	}

	if message.UserID != userID {
		return errors.New("unauthorized: only message owner can delete")
	}

	return s.messageRepo.Delete(ctx, messageID)
}
