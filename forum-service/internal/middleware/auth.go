package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/greygn/forum-service/internal/config"
	"go.uber.org/zap"
)

type AuthMiddleware struct {
	config *config.Config
	logger *zap.Logger
}

func NewAuthMiddleware(config *config.Config, logger *zap.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		config: config,
		logger: logger,
	}
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header is required", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		// Call auth service to validate token
		client := &http.Client{}
		req, err := http.NewRequest("GET", m.config.AuthServiceURL+"/api/v1/auth/validate", nil)
		if err != nil {
			m.logger.Error("failed to create request", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		req.Header.Set("Authorization", authHeader)

		resp, err := client.Do(req)
		if err != nil {
			m.logger.Error("failed to validate token", zap.Error(err))
			http.Error(w, "Failed to validate token", http.StatusUnauthorized)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Parse response to get user info
		var validateResp struct {
			UserID   string `json:"user_id"`
			Username string `json:"username"`
			IsValid  bool   `json:"is_valid"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&validateResp); err != nil {
			m.logger.Error("failed to decode response", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if !validateResp.IsValid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Add user info to context
		ctx := context.WithValue(r.Context(), "user_id", validateResp.UserID)
		ctx = context.WithValue(ctx, "username", validateResp.Username)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
