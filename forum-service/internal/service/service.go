package service

import (
	"context"
	"errors"
	"time"

	"github.com/greygn/forum-service/internal/repository"
)

type PostService interface {
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

type postService struct {
	repo repository.PostRepository
}

func NewPostService(repo repository.PostRepository) PostService {
	return &postService{repo: repo}
}

func (s *postService) CreatePost(ctx context.Context, userID, username, title, content string) (*repository.Post, error) {
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

func (s *postService) GetAllPosts(ctx context.Context) ([]repository.Post, error) {
	return s.repo.GetAllPosts(ctx)
}

func (s *postService) GetPostByID(ctx context.Context, id string) (*repository.Post, error) {
	return s.repo.GetPostByID(ctx, id)
}

func (s *postService) UpdatePost(ctx context.Context, id, userID, title, content string) error {
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

func (s *postService) DeletePost(ctx context.Context, id string) error {
	return s.repo.DeletePost(ctx, id)
}

func (s *postService) DeleteOldPosts(ctx context.Context, olderThan time.Duration) error {
	return s.repo.DeleteOldPosts(ctx, olderThan)
}

func (s *postService) GetComments(ctx context.Context, postID string) ([]repository.Comment, error) {
	return s.repo.GetComments(ctx, postID)
}

func (s *postService) CreateComment(ctx context.Context, postID, userID, username, content string) (*repository.Comment, error) {
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

func (s *postService) GetCommentByID(ctx context.Context, id string) (*repository.Comment, error) {
	return s.repo.GetCommentByID(ctx, id)
}

func (s *postService) UpdateComment(ctx context.Context, id, userID, content string) error {
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

func (s *postService) DeleteComment(ctx context.Context, id string) error {
	return s.repo.DeleteComment(ctx, id)
}
