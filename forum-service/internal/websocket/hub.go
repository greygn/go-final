package websocket

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan *Message
	userID   string
	username string
}

type Message struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan *Message
	register   chan *Client
	unregister chan *Client
	messages   []*Message
	mu         sync.RWMutex
	ttl        time.Duration
}

func NewHub(messageTTL time.Duration) *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan *Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		messages:   make([]*Message, 0),
		ttl:        messageTTL,
	}
}

func (h *Hub) Run() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			// Send recent messages to newly connected client
			h.mu.RLock()
			for _, msg := range h.messages {
				client.send <- msg
			}
			h.mu.RUnlock()

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}

		case message := <-h.broadcast:
			message.ID = uuid.New().String()
			message.CreatedAt = time.Now()

			h.mu.Lock()
			h.messages = append(h.messages, message)
			h.mu.Unlock()

			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}

		case <-ticker.C:
			h.cleanupOldMessages()
		}
	}
}

func (h *Hub) cleanupOldMessages() {
	threshold := time.Now().Add(-h.ttl)

	h.mu.Lock()
	defer h.mu.Unlock()

	var validMessages []*Message
	for _, msg := range h.messages {
		if msg.CreatedAt.After(threshold) {
			validMessages = append(validMessages, msg)
		}
	}
	h.messages = validMessages
}
