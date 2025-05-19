package postservice

import (
	"context"
	"errors"
	"time"

	"github.com/greygn/forum-service/internal/repository"
)

type Service interface {
	// Post operations
	CreatePost(ctx context.Context, userID, username, title, content string) (*repository.Post, error)
	GetAllPosts(ctx context.Context) ([]repository.Post, error)
	GetPostByID(ctx context.Context, id string) (*repository.Post, error)
	UpdatePost(ctx context.Context, id, userID, title, content string) error
	DeletePost(ctx context.Context, id string) error
	DeleteOldPosts(ctx context.Context, olderThan time.Duration) error

	// Comment operations
	GetComments(ctx context.Context, postID string) ([]repository.Comment, error)
	CreateComment(ctx context.Context, postID, userID, username, content string) (*repository.Comment, error)
	GetCommentByID(ctx context.Context, id string) (*repository.Comment, error)
	UpdateComment(ctx context.Context, id, userID, content string) error
	DeleteComment(ctx context.Context, id string) error
}

type service struct {
	repo repository.PostRepository
}

func New(repo repository.PostRepository) Service {
	return &service{repo: repo}
}

func (s *service) CreatePost(ctx context.Context, userID, username, title, content string) (*repository.Post, error) {
	if title == "" {
		return nil, errors.New("title is required")
	}
	if content == "" {
		return nil, errors.New("content is required")
	}

	post := &repository.Post{
		UserID:   userID,
		Username: username,
		Title:    title,
		Content:  content,
	}

	if err := s.repo.CreatePost(ctx, post); err != nil {
		return nil, err
	}

	return post, nil
}

func (s *service) GetAllPosts(ctx context.Context) ([]repository.Post, error) {
	return s.repo.GetAllPosts(ctx)
}

func (s *service) GetPostByID(ctx context.Context, id string) (*repository.Post, error) {
	return s.repo.GetPostByID(ctx, id)
}

func (s *service) UpdatePost(ctx context.Context, id, userID, title, content string) error {
	if title == "" {
		return errors.New("title is required")
	}
	if content == "" {
		return errors.New("content is required")
	}

	post := &repository.Post{
		ID:      id,
		UserID:  userID,
		Title:   title,
		Content: content,
	}

	return s.repo.UpdatePost(ctx, post)
}

func (s *service) DeletePost(ctx context.Context, id string) error {
	return s.repo.DeletePost(ctx, id)
}

func (s *service) DeleteOldPosts(ctx context.Context, olderThan time.Duration) error {
	return s.repo.DeleteOldPosts(ctx, olderThan)
}

func (s *service) GetComments(ctx context.Context, postID string) ([]repository.Comment, error) {
	return s.repo.GetComments(ctx, postID)
}

func (s *service) CreateComment(ctx context.Context, postID, userID, username, content string) (*repository.Comment, error) {
	if content == "" {
		return nil, errors.New("content is required")
	}

	comment := &repository.Comment{
		PostID:   postID,
		UserID:   userID,
		Username: username,
		Content:  content,
	}

	if err := s.repo.CreateComment(ctx, comment); err != nil {
		return nil, err
	}

	return comment, nil
}

func (s *service) GetCommentByID(ctx context.Context, id string) (*repository.Comment, error) {
	return s.repo.GetCommentByID(ctx, id)
}

func (s *service) UpdateComment(ctx context.Context, id, userID, content string) error {
	if content == "" {
		return errors.New("content is required")
	}

	comment := &repository.Comment{
		ID:      id,
		UserID:  userID,
		Content: content,
	}

	return s.repo.UpdateComment(ctx, comment)
}

func (s *service) DeleteComment(ctx context.Context, id string) error {
	return s.repo.DeleteComment(ctx, id)
}
