package service

import (
	"context"
	"errors"
	"time"

	"github.com/greygn/forum-service/internal/repository"
)

type PostService interface {
	// Post operations
	CreatePost(ctx context.Context, post *repository.Post) error
	GetAllPosts(ctx context.Context) ([]repository.Post, error)
	GetPostByID(ctx context.Context, id string) (*repository.Post, error)
	UpdatePost(ctx context.Context, post *repository.Post) error
	DeletePost(ctx context.Context, id string, userID string) error
	DeleteOldPosts(ctx context.Context, olderThan time.Duration) error

	// Comment operations
	GetComments(ctx context.Context, postID string) ([]repository.Comment, error)
	CreateComment(ctx context.Context, comment *repository.Comment) error
	GetCommentByID(ctx context.Context, id string) (*repository.Comment, error)
	UpdateComment(ctx context.Context, comment *repository.Comment) error
	DeleteComment(ctx context.Context, id string, userID string) error
}

type postService struct {
	repo repository.PostRepository
}

func NewPostService(repo repository.PostRepository) PostService {
	return &postService{repo: repo}
}

func (s *postService) CreatePost(ctx context.Context, post *repository.Post) error {
	if post.Title == "" {
		return errors.New("title is required")
	}
	if post.Content == "" {
		return errors.New("content is required")
	}

	return s.repo.CreatePost(ctx, post)
}

func (s *postService) GetAllPosts(ctx context.Context) ([]repository.Post, error) {
	return s.repo.GetAllPosts(ctx)
}

func (s *postService) GetPostByID(ctx context.Context, id string) (*repository.Post, error) {
	return s.repo.GetPostByID(ctx, id)
}

func (s *postService) UpdatePost(ctx context.Context, post *repository.Post) error {
	if post.Title == "" {
		return errors.New("title is required")
	}
	if post.Content == "" {
		return errors.New("content is required")
	}

	return s.repo.UpdatePost(ctx, post)
}

func (s *postService) DeletePost(ctx context.Context, id string, userID string) error {
	return s.repo.DeletePost(ctx, id)
}

func (s *postService) DeleteOldPosts(ctx context.Context, olderThan time.Duration) error {
	return s.repo.DeleteOldPosts(ctx, olderThan)
}

func (s *postService) GetComments(ctx context.Context, postID string) ([]repository.Comment, error) {
	return s.repo.GetComments(ctx, postID)
}

func (s *postService) CreateComment(ctx context.Context, comment *repository.Comment) error {
	if comment.Content == "" {
		return errors.New("content is required")
	}

	return s.repo.CreateComment(ctx, comment)
}

func (s *postService) GetCommentByID(ctx context.Context, id string) (*repository.Comment, error) {
	return s.repo.GetCommentByID(ctx, id)
}

func (s *postService) UpdateComment(ctx context.Context, comment *repository.Comment) error {
	if comment.Content == "" {
		return errors.New("content is required")
	}

	return s.repo.UpdateComment(ctx, comment)
}

func (s *postService) DeleteComment(ctx context.Context, id string, userID string) error {
	return s.repo.DeleteComment(ctx, id)
}
