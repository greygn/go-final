package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/greygn/forum-service/internal/config"
	"github.com/greygn/forum-service/internal/middleware"
	"github.com/greygn/forum-service/internal/repository"
	"github.com/greygn/forum-service/internal/service"
	httpTransport "github.com/greygn/forum-service/internal/transport/http"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	cfg := config.Load()

	// Initialize PostgreSQL connection
	db, err := repository.NewPostgresDB(cfg.PostgresURL)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Run migrations
	if err := repository.RunMigrations(cfg.PostgresURL, "file://migrations"); err != nil {
		logger.Fatal("failed to run migrations", zap.Error(err))
	}

	// Initialize repositories
	messageRepo := repository.NewMessageRepository(db)

	// Initialize services
	chatService := service.NewChatService(messageRepo, cfg, logger)

	// Start chat service
	go chatService.Run()

	// Initialize HTTP server
	httpServer := httpTransport.NewServer(chatService, logger)

	// Initialize auth middleware
	authMiddleware := middleware.NewAuthMiddleware(cfg, logger)

	// Create mux with auth middleware
	mux := http.NewServeMux()
	mux.Handle("/", authMiddleware.Authenticate(httpServer))

	// Start HTTP server
	go func() {
		logger.Info("starting HTTP server", zap.String("addr", cfg.HTTPAddr))
		if err := http.ListenAndServe(cfg.HTTPAddr, mux); err != nil {
			logger.Fatal("failed to serve HTTP", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	logger.Info("shutting down server")
}
