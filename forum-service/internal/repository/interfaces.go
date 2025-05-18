package repository

import (
	"context"
	"time"
)

type Message struct {
	ID        int64
	UserID    int64
	Username  string
	Content   string
	CreatedAt time.Time
}

type MessageRepository interface {
	Create(ctx context.Context, message *Message) error
	List(ctx context.Context) ([]*Message, error)
	DeleteOld(ctx context.Context, olderThan time.Duration) error
}
