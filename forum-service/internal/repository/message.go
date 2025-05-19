package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type MessageRepository interface {
	Create(ctx context.Context, message *Message) error
	GetAll(ctx context.Context) ([]Message, error)
	GetByID(ctx context.Context, id string) (*Message, error)
	Update(ctx context.Context, message *Message) error
	Delete(ctx context.Context, id string) error
	DeleteOld(ctx context.Context, olderThan time.Duration) error
}

type messageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) Create(ctx context.Context, message *Message) error {
	query := `
		INSERT INTO messages (id, user_id, username, content, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	message.ID = uuid.New().String()
	message.CreatedAt = time.Now()
	_, err := r.db.ExecContext(ctx, query, message.ID, message.UserID, message.Username, message.Content, message.CreatedAt)
	return err
}

func (r *messageRepository) GetAll(ctx context.Context) ([]Message, error) {
	query := `
		SELECT id, user_id, username, content, created_at
		FROM messages
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.ID, &msg.UserID, &msg.Username, &msg.Content, &msg.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	return messages, nil
}

func (r *messageRepository) GetByID(ctx context.Context, id string) (*Message, error) {
	query := `
		SELECT id, user_id, username, content, created_at
		FROM messages
		WHERE id = $1
	`
	var msg Message
	err := r.db.QueryRowContext(ctx, query, id).Scan(&msg.ID, &msg.UserID, &msg.Username, &msg.Content, &msg.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &msg, nil
}

func (r *messageRepository) Update(ctx context.Context, message *Message) error {
	query := `
		UPDATE messages
		SET content = $1
		WHERE id = $2 AND user_id = $3
	`
	result, err := r.db.ExecContext(ctx, query, message.Content, message.ID, message.UserID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("message not found or unauthorized")
	}
	return nil
}

func (r *messageRepository) Delete(ctx context.Context, id string) error {
	query := `
		DELETE FROM messages
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *messageRepository) DeleteOld(ctx context.Context, olderThan time.Duration) error {
	query := `
		DELETE FROM messages
		WHERE created_at < $1
	`
	_, err := r.db.ExecContext(ctx, query, time.Now().Add(-olderThan))
	return err
}
