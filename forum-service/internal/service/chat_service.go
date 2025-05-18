package service

import (
	"context"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/greygn/forum-service/internal/config"
	"github.com/greygn/forum-service/internal/repository"
	"go.uber.org/zap"
)

type Client struct {
	ID       string
	UserID   int64
	Username string
	Conn     *websocket.Conn
	Send     chan []byte
}

type ChatService struct {
	messageRepo repository.MessageRepository
	config      *config.Config
	logger      *zap.Logger
	clients     map[*Client]bool
	broadcast   chan []byte
	register    chan *Client
	unregister  chan *Client
	mutex       sync.RWMutex
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
	ticker := time.NewTicker(s.config.MessageTTL)
	defer ticker.Stop()

	for {
		select {
		case client := <-s.register:
			s.mutex.Lock()
			s.clients[client] = true
			s.mutex.Unlock()

		case client := <-s.unregister:
			s.mutex.Lock()
			if _, ok := s.clients[client]; ok {
				delete(s.clients, client)
				close(client.Send)
			}
			s.mutex.Unlock()

		case message := <-s.broadcast:
			s.mutex.RLock()
			for client := range s.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(s.clients, client)
				}
			}
			s.mutex.RUnlock()

		case <-ticker.C:
			if err := s.messageRepo.DeleteOld(context.Background(), s.config.MessageTTL); err != nil {
				s.logger.Error("failed to delete old messages", zap.Error(err))
			}
		}
	}
}

func (s *ChatService) GetMessages(ctx context.Context) ([]*repository.Message, error) {
	return s.messageRepo.List(ctx)
}

func (s *ChatService) SaveMessage(ctx context.Context, userID int64, username, content string) error {
	message := &repository.Message{
		UserID:    userID,
		Username:  username,
		Content:   content,
		CreatedAt: time.Now(),
	}

	return s.messageRepo.Create(ctx, message)
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
