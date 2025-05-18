package repository

import (
	"context"
	"database/sql"
	"time"
)

type messageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) Create(ctx context.Context, message *Message) error {
	query := `
		INSERT INTO messages (user_id, username, content, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id`

	return r.db.QueryRowContext(ctx, query,
		message.UserID,
		message.Username,
		message.Content,
		time.Now(),
	).Scan(&message.ID)
}

func (r *messageRepository) List(ctx context.Context) ([]*Message, error) {
	query := `
		SELECT id, user_id, username, content, created_at
		FROM messages
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*Message
	for rows.Next() {
		message := &Message{}
		if err := rows.Scan(
			&message.ID,
			&message.UserID,
			&message.Username,
			&message.Content,
			&message.CreatedAt,
		); err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

func (r *messageRepository) DeleteOld(ctx context.Context, olderThan time.Duration) error {
	query := `DELETE FROM messages WHERE created_at < $1`
	_, err := r.db.ExecContext(ctx, query, time.Now().Add(-olderThan))
	return err
}
