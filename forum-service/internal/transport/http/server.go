package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/greygn/forum-service/internal/service"
	"go.uber.org/zap"
)

type Server struct {
	chatService *service.ChatService
	logger      *zap.Logger
	upgrader    websocket.Upgrader
}

type CreateMessageRequest struct {
	Content string `json:"content"`
}

func NewServer(chatService *service.ChatService, logger *zap.Logger) *Server {
	return &Server{
		chatService: chatService,
		logger:      logger,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for now
			},
		},
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Handle API v1 routes
	if strings.HasPrefix(r.URL.Path, "/api/v1") {
		switch {
		case strings.HasPrefix(r.URL.Path, "/api/v1/messages"):
			s.handleMessages(w, r)
		case strings.HasPrefix(r.URL.Path, "/api/v1/ws"):
			s.handleWebSocket(w, r)
		default:
			http.NotFound(w, r)
		}
		return
	}

	// Handle legacy routes
	switch r.URL.Path {
	case "/ws":
		s.handleWebSocket(w, r)
	case "/messages":
		s.handleMessages(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Error("failed to upgrade connection", zap.Error(err))
		return
	}

	// Get user info from context or token
	userID := r.Context().Value("user_id").(int64)
	username := r.Context().Value("username").(string)

	client := &service.Client{
		Conn:     conn,
		Send:     make(chan []byte, 256),
		UserID:   userID,
		Username: username,
	}

	s.chatService.Register(client)
	defer s.chatService.Unregister(client)

	go s.writePump(client)
	s.readPump(r.Context(), client)
}

func (s *Server) writePump(client *service.Client) {
	ticker := time.NewTicker(time.Second * 54)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := client.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (s *Server) readPump(ctx context.Context, client *service.Client) {
	defer client.Conn.Close()
	client.Conn.SetReadLimit(512)
	client.Conn.SetReadDeadline(time.Now().Add(time.Second * 60))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(time.Second * 60))
		return nil
	})

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				s.logger.Error("unexpected close error", zap.Error(err))
			}
			break
		}

		// Process message
		if err := s.chatService.SaveMessage(ctx, client.UserID, client.Username, string(message)); err != nil {
			s.logger.Error("failed to save message", zap.Error(err))
			continue
		}

		s.chatService.Broadcast(message)
	}
}

func (s *Server) handleMessages(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		messages, err := s.chatService.GetMessages(r.Context())
		if err != nil {
			s.logger.Error("failed to get messages", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(messages)

	case http.MethodPost:
		var req CreateMessageRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Get user info from context or token
		userID := r.Context().Value("user_id").(int64)
		username := r.Context().Value("username").(string)

		if err := s.chatService.SaveMessage(r.Context(), userID, username, req.Content); err != nil {
			s.logger.Error("failed to save message", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
