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

type UpdateMessageRequest struct {
	Content string `json:"content"`
}

type CreateCommentRequest struct {
	Content string `json:"content"`
}

type UpdateCommentRequest struct {
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
		path := strings.TrimPrefix(r.URL.Path, "/api/v1")
		switch {
		case strings.HasPrefix(path, "/messages/"):
			// Handle message-specific operations
			parts := strings.Split(path, "/")
			if len(parts) >= 3 {
				messageID := parts[2]
				if len(parts) >= 4 && parts[3] == "comments" {
					s.handleComments(w, r, messageID)
				} else {
					s.handleMessage(w, r, messageID)
				}
			} else {
				s.handleMessages(w, r)
			}
		case strings.HasPrefix(path, "/comments/"):
			// Handle comment-specific operations
			parts := strings.Split(path, "/")
			if len(parts) >= 3 {
				commentID := parts[2]
				s.handleComment(w, r, commentID)
			} else {
				http.NotFound(w, r)
			}
		case strings.HasPrefix(path, "/messages"):
			s.handleMessages(w, r)
		case strings.HasPrefix(path, "/ws"):
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

	// Get user info from context
	userID := r.Context().Value("user_id").(string)
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

		// Get user info from context
		userID := r.Context().Value("user_id").(string)
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

func (s *Server) handleMessage(w http.ResponseWriter, r *http.Request, messageID string) {
	switch r.Method {
	case http.MethodGet:
		message, err := s.chatService.GetMessage(r.Context(), messageID)
		if err != nil {
			s.logger.Error("failed to get message", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(message)

	case http.MethodPut:
		var req UpdateMessageRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		userID := r.Context().Value("user_id").(string)
		if err := s.chatService.UpdateMessage(r.Context(), messageID, userID, req.Content); err != nil {
			s.logger.Error("failed to update message", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)

	case http.MethodDelete:
		userID := r.Context().Value("user_id").(string)
		if err := s.chatService.DeleteMessage(r.Context(), messageID, userID); err != nil {
			s.logger.Error("failed to delete message", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleComments(w http.ResponseWriter, r *http.Request, messageID string) {
	switch r.Method {
	case http.MethodGet:
		comments, err := s.chatService.GetComments(r.Context(), messageID)
		if err != nil {
			s.logger.Error("failed to get comments", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(comments)

	case http.MethodPost:
		var req CreateCommentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		userID := r.Context().Value("user_id").(string)
		username := r.Context().Value("username").(string)

		if err := s.chatService.CreateComment(r.Context(), messageID, userID, username, req.Content); err != nil {
			s.logger.Error("failed to create comment", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleComment(w http.ResponseWriter, r *http.Request, commentID string) {
	switch r.Method {
	case http.MethodPut:
		var req UpdateCommentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		userID := r.Context().Value("user_id").(string)
		if err := s.chatService.UpdateComment(r.Context(), commentID, userID, req.Content); err != nil {
			s.logger.Error("failed to update comment", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)

	case http.MethodDelete:
		userID := r.Context().Value("user_id").(string)
		if err := s.chatService.DeleteComment(r.Context(), commentID, userID); err != nil {
			s.logger.Error("failed to delete comment", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
