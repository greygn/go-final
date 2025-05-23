package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"auth-service/internal/config"
	"auth-service/internal/handler"
	"auth-service/internal/logger"
	"auth-service/internal/repository"
	"auth-service/internal/service"
)

// @title           Auth Service API
// @version         1.0
// @description     Authentication service with JWT token support.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey  BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	logger := logger.GetLogger()

	db, err := repository.NewPostgresDB(cfg.DatabaseURL)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(db)
	authService := service.NewAuthService(userRepo, tokenRepo, cfg)
	authHandler := handler.NewAuthHandler(authService)

	r := gin.Default()

	// API routes
	v1 := r.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			// @Summary Register a new user
			// @Description Register a new user with username, email and password
			// @Tags auth
			// @Accept json
			// @Produce json
			// @Param input body handler.RegisterRequest true "Registration details"
			// @Success 200 {object} handler.Response
			// @Failure 400 {object} handler.ErrorResponse
			// @Router /auth/register [post]
			auth.POST("/register", authHandler.Register)

			// @Summary Login user
			// @Description Login with username and password to get JWT tokens
			// @Tags auth
			// @Accept json
			// @Produce json
			// @Param input body handler.LoginRequest true "Login credentials"
			// @Success 200 {object} handler.LoginResponse
			// @Failure 400 {object} handler.ErrorResponse
			// @Router /auth/login [post]
			auth.POST("/login", authHandler.Login)

			// @Summary Refresh token
			// @Description Get new access token using refresh token
			// @Tags auth
			// @Accept json
			// @Produce json
			// @Param input body handler.RefreshRequest true "Refresh token"
			// @Success 200 {object} handler.LoginResponse
			// @Failure 400 {object} handler.ErrorResponse
			// @Router /auth/refresh [post]
			auth.POST("/refresh", authHandler.RefreshToken)

			// @Summary Logout user
			// @Description Logout user by invalidating refresh token
			// @Tags auth
			// @Accept json
			// @Produce json
			// @Param input body handler.RefreshRequest true "Refresh token"
			// @Success 200 {object} handler.Response
			// @Failure 400 {object} handler.ErrorResponse
			// @Router /auth/logout [post]
			auth.POST("/logout", authHandler.Logout)

			// @Summary Validate token
			// @Description Validate JWT token and return user information
			// @Tags auth
			// @Accept json
			// @Produce json
			// @Security BearerAuth
			// @Success 200 {object} handler.ValidateResponse
			// @Failure 401 {object} handler.ErrorResponse
			// @Router /auth/validate [get]
			auth.GET("/validate", authHandler.ValidateToken)
		}
	}

	if err := r.Run(":8080"); err != nil {
		logger.Fatal().Err(err).Msg("Failed to start server")
	}
}
