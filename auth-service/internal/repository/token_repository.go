package repository

import (
	"context"
	"database/sql"
	"errors"
)

type tokenRepository struct {
	db *sql.DB
}

func NewTokenRepository(db *sql.DB) TokenRepository {
	return &tokenRepository{db: db}
}

func (r *tokenRepository) Create(ctx context.Context, token *RefreshToken) error {
	query := `
		INSERT INTO tokens (user_id, refresh_token, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`

	return r.db.QueryRowContext(ctx, query,
		token.UserID,
		token.Token,
		token.ExpiresAt,
	).Scan(&token.ID, &token.CreatedAt)
}

func (r *tokenRepository) Get(ctx context.Context, token string) (*RefreshToken, error) {
	refreshToken := &RefreshToken{}
	query := `
		SELECT id, user_id, refresh_token, expires_at, created_at
		FROM tokens
		WHERE refresh_token = $1`

	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&refreshToken.ID,
		&refreshToken.UserID,
		&refreshToken.Token,
		&refreshToken.ExpiresAt,
		&refreshToken.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, errors.New("token not found")
	}
	return refreshToken, err
}

func (r *tokenRepository) Delete(ctx context.Context, token string) error {
	query := `DELETE FROM tokens WHERE refresh_token = $1`
	result, err := r.db.ExecContext(ctx, query, token)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New("token not found")
	}

	return nil
}

func (r *tokenRepository) DeleteAllForUser(ctx context.Context, userID string) error {
	query := `DELETE FROM tokens WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}
