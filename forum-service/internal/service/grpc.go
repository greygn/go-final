package service

import (
	"context"
	"time"

	"github.com/greygn/forum-service/internal/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/greygn/protos/proto/forum"

	"github.com/google/uuid"
	"github.com/greygn/forum-service/internal/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCService struct {
	forum.UnimplementedForumServiceServer
	postService PostService
}

func NewGRPCService(postService PostService) *GRPCService {
	return &GRPCService{
		postService: postService,
	}
}

// Post operations
func (s *GRPCService) CreatePost(ctx context.Context, req *forum.CreatePostRequest) (*forum.CreatePostResponse, error) {
	if req.Title == "" || req.Content == "" {
		return nil, status.Error(codes.InvalidArgument, "title and content are required")
	}

	post := &repository.Post{
		ID:        uuid.New().String(),
		UserID:    req.UserId,
		Username:  req.Username,
		Title:     req.Title,
		Content:   req.Content,
		CreatedAt: time.Now(),
	}

	if err := s.postService.CreatePost(ctx, post); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &forum.CreatePostResponse{
		Success: true,
		Post: &forum.Post{
			Id:        post.ID,
			UserId:    post.UserID,
			Username:  post.Username,
			Title:     post.Title,
			Content:   post.Content,
			CreatedAt: post.CreatedAt.Unix(),
		},
	}, nil
}

func (s *GRPCService) GetPosts(ctx context.Context, req *forum.GetPostsRequest) (*forum.GetPostsResponse, error) {
	posts, err := s.postService.GetAllPosts(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	protoPosts := make([]*forum.Post, len(posts))
	for i, post := range posts {
		protoPosts[i] = &forum.Post{
			Id:        post.ID,
			UserId:    post.UserID,
			Username:  post.Username,
			Title:     post.Title,
			Content:   post.Content,
			CreatedAt: post.CreatedAt.Unix(),
		}
	}

	return &forum.GetPostsResponse{
		Posts: protoPosts,
	}, nil
}

func (s *GRPCService) GetPost(ctx context.Context, req *forum.GetPostRequest) (*forum.GetPostResponse, error) {
	post, err := s.postService.GetPostByID(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &forum.GetPostResponse{
		Success: true,
		Post: &forum.Post{
			Id:        post.ID,
			UserId:    post.UserID,
			Username:  post.Username,
			Title:     post.Title,
			Content:   post.Content,
			CreatedAt: post.CreatedAt.Unix(),
		},
	}, nil
}

func (s *GRPCService) UpdatePost(ctx context.Context, req *forum.UpdatePostRequest) (*forum.UpdatePostResponse, error) {
	if req.Title == "" || req.Content == "" {
		return nil, status.Error(codes.InvalidArgument, "title and content are required")
	}

	post := &repository.Post{
		ID:      req.Id,
		UserID:  req.UserId,
		Title:   req.Title,
		Content: req.Content,
	}

	if err := s.postService.UpdatePost(ctx, post); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &forum.UpdatePostResponse{
		Success: true,
	}, nil
}

func (s *GRPCService) DeletePost(ctx context.Context, req *forum.DeletePostRequest) (*forum.DeletePostResponse, error) {
	if err := s.postService.DeletePost(ctx, req.Id, req.UserId); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &forum.DeletePostResponse{
		Success: true,
	}, nil
}

// Comment operations
func (s *GRPCService) CreateComment(ctx context.Context, req *forum.CreateCommentRequest) (*forum.CreateCommentResponse, error) {
	if req.Content == "" {
		return nil, status.Error(codes.InvalidArgument, "content is required")
	}

	comment := &repository.Comment{
		ID:        uuid.New().String(),
		PostID:    req.PostId,
		UserID:    req.UserId,
		Username:  req.Username,
		Content:   req.Content,
		CreatedAt: time.Now(),
	}

	if err := s.postService.CreateComment(ctx, comment); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &forum.CreateCommentResponse{
		Success: true,
		Comment: &forum.Comment{
			Id:        comment.ID,
			PostId:    comment.PostID,
			UserId:    comment.UserID,
			Username:  comment.Username,
			Content:   comment.Content,
			CreatedAt: comment.CreatedAt.Unix(),
		},
	}, nil
}

func (s *GRPCService) GetComments(ctx context.Context, req *forum.GetCommentsRequest) (*forum.GetCommentsResponse, error) {
	comments, err := s.postService.GetComments(ctx, req.PostId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	protoComments := make([]*forum.Comment, len(comments))
	for i, comment := range comments {
		protoComments[i] = &forum.Comment{
			Id:        comment.ID,
			PostId:    comment.PostID,
			UserId:    comment.UserID,
			Username:  comment.Username,
			Content:   comment.Content,
			CreatedAt: comment.CreatedAt.Unix(),
		}
	}

	return &forum.GetCommentsResponse{
		Comments: protoComments,
	}, nil
}

func (s *GRPCService) GetComment(ctx context.Context, req *forum.GetCommentRequest) (*forum.GetCommentResponse, error) {
	comment, err := s.postService.GetCommentByID(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &forum.GetCommentResponse{
		Success: true,
		Comment: &forum.Comment{
			Id:        comment.ID,
			PostId:    comment.PostID,
			UserId:    comment.UserID,
			Username:  comment.Username,
			Content:   comment.Content,
			CreatedAt: comment.CreatedAt.Unix(),
		},
	}, nil
}

func (s *GRPCService) UpdateComment(ctx context.Context, req *forum.UpdateCommentRequest) (*forum.UpdateCommentResponse, error) {
	if req.Content == "" {
		return nil, status.Error(codes.InvalidArgument, "content is required")
	}

	comment := &repository.Comment{
		ID:      req.Id,
		UserID:  req.UserId,
		Content: req.Content,
	}

	if err := s.postService.UpdateComment(ctx, comment); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &forum.UpdateCommentResponse{
		Success: true,
	}, nil
}

func (s *GRPCService) DeleteComment(ctx context.Context, req *forum.DeleteCommentRequest) (*forum.DeleteCommentResponse, error) {
	if err := s.postService.DeleteComment(ctx, req.Id, req.UserId); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &forum.DeleteCommentResponse{
		Success: true,
	}, nil
}
