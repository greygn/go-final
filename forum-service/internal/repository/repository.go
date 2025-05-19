package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Post struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type Comment struct {
	ID        string    `json:"id"`
	PostID    string    `json:"post_id"`
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type PostRepository interface {
	// Post operations
	CreatePost(ctx context.Context, post *Post) error
	GetAllPosts(ctx context.Context) ([]Post, error)
	GetPostByID(ctx context.Context, id string) (*Post, error)
	UpdatePost(ctx context.Context, post *Post) error
	DeletePost(ctx context.Context, id string) error
	DeleteOldPosts(ctx context.Context, olderThan time.Duration) error

	// Comment operations
	GetComments(ctx context.Context, postID string) ([]Comment, error)
	CreateComment(ctx context.Context, comment *Comment) error
	GetCommentByID(ctx context.Context, id string) (*Comment, error)
	UpdateComment(ctx context.Context, comment *Comment) error
	DeleteComment(ctx context.Context, id string) error
}

type postRepository struct {
	db *sql.DB
}

func NewPostRepository(db *sql.DB) PostRepository {
	return &postRepository{db: db}
}

func (r *postRepository) CreatePost(ctx context.Context, post *Post) error {
	query := `
		INSERT INTO posts (id, user_id, username, title, content, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	post.ID = uuid.New().String()
	post.CreatedAt = time.Now()
	_, err := r.db.ExecContext(ctx, query, post.ID, post.UserID, post.Username, post.Title, post.Content, post.CreatedAt)
	return err
}

func (r *postRepository) GetAllPosts(ctx context.Context) ([]Post, error) {
	query := `
		SELECT id, user_id, username, title, content, created_at
		FROM posts
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.UserID, &post.Username, &post.Title, &post.Content, &post.CreatedAt); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	return posts, nil
}

func (r *postRepository) GetPostByID(ctx context.Context, id string) (*Post, error) {
	query := `
		SELECT id, user_id, username, title, content, created_at
		FROM posts
		WHERE id = $1
	`
	var post Post
	err := r.db.QueryRowContext(ctx, query, id).Scan(&post.ID, &post.UserID, &post.Username, &post.Title, &post.Content, &post.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &post, nil
}

func (r *postRepository) UpdatePost(ctx context.Context, post *Post) error {
	query := `
		UPDATE posts
		SET title = $1, content = $2
		WHERE id = $3 AND user_id = $4
	`
	result, err := r.db.ExecContext(ctx, query, post.Title, post.Content, post.ID, post.UserID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("post not found or unauthorized")
	}
	return nil
}

func (r *postRepository) DeletePost(ctx context.Context, id string) error {
	query := `
		DELETE FROM posts
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *postRepository) DeleteOldPosts(ctx context.Context, olderThan time.Duration) error {
	query := `
		DELETE FROM posts
		WHERE created_at < $1
	`
	_, err := r.db.ExecContext(ctx, query, time.Now().Add(-olderThan))
	return err
}

func (r *postRepository) GetComments(ctx context.Context, postID string) ([]Comment, error) {
	query := `
		SELECT id, post_id, user_id, username, content, created_at
		FROM comments
		WHERE post_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var comment Comment
		if err := rows.Scan(&comment.ID, &comment.PostID, &comment.UserID, &comment.Username, &comment.Content, &comment.CreatedAt); err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}
	return comments, nil
}

func (r *postRepository) CreateComment(ctx context.Context, comment *Comment) error {
	query := `
		INSERT INTO comments (id, post_id, user_id, username, content, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	comment.ID = uuid.New().String()
	comment.CreatedAt = time.Now()
	_, err := r.db.ExecContext(ctx, query, comment.ID, comment.PostID, comment.UserID, comment.Username, comment.Content, comment.CreatedAt)
	return err
}

func (r *postRepository) GetCommentByID(ctx context.Context, id string) (*Comment, error) {
	query := `
		SELECT id, post_id, user_id, username, content, created_at
		FROM comments
		WHERE id = $1
	`
	var comment Comment
	err := r.db.QueryRowContext(ctx, query, id).Scan(&comment.ID, &comment.PostID, &comment.UserID, &comment.Username, &comment.Content, &comment.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &comment, nil
}

func (r *postRepository) UpdateComment(ctx context.Context, comment *Comment) error {
	query := `
		UPDATE comments
		SET content = $1
		WHERE id = $2 AND user_id = $3
	`
	result, err := r.db.ExecContext(ctx, query, comment.Content, comment.ID, comment.UserID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("comment not found or unauthorized")
	}
	return nil
}

func (r *postRepository) DeleteComment(ctx context.Context, id string) error {
	query := `
		DELETE FROM comments
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
