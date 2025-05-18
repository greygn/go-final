package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/greygn/go-final/pkg/common"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	authHeader   = "Authorization"
	bearerSchema = "Bearer "
	userIDKey    = "user_id"
	usernameKey  = "username"
)

// AuthMiddleware is a middleware that validates JWT tokens
func AuthMiddleware(validateToken func(string) (string, string, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(authHeader)
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": common.ErrUnauthorized.Error()})
			return
		}

		if !strings.HasPrefix(authHeader, bearerSchema) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": common.ErrInvalidToken.Error()})
			return
		}

		token := strings.TrimPrefix(authHeader, bearerSchema)
		userID, username, err := validateToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		c.Set(userIDKey, userID)
		c.Set(usernameKey, username)
		c.Next()
	}
}

// GRPCAuthInterceptor is a gRPC interceptor that validates JWT tokens
func GRPCAuthInterceptor(validateToken func(string) (string, string, error)) func(context.Context, interface{}, *grpc.UnaryServerInfo, grpc.UnaryHandler) (interface{}, error) {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, common.ErrUnauthorized
		}

		values := md.Get(authHeader)
		if len(values) == 0 {
			return nil, common.ErrUnauthorized
		}

		authHeader := values[0]
		if !strings.HasPrefix(authHeader, bearerSchema) {
			return nil, common.ErrInvalidToken
		}

		token := strings.TrimPrefix(authHeader, bearerSchema)
		userID, username, err := validateToken(token)
		if err != nil {
			return nil, err
		}

		newCtx := context.WithValue(ctx, userIDKey, userID)
		newCtx = context.WithValue(newCtx, usernameKey, username)

		return handler(newCtx, req)
	}
}

// GetUserID gets the user ID from the context
func GetUserID(c *gin.Context) string {
	return c.GetString(userIDKey)
}

// GetUsername gets the username from the context
func GetUsername(c *gin.Context) string {
	return c.GetString(usernameKey)
}
