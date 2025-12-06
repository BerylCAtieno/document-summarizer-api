package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/BerylCAtieno/document-summarizer-api/internal/config"
	"github.com/BerylCAtieno/document-summarizer-api/internal/db"
	"github.com/BerylCAtieno/document-summarizer-api/internal/repository"
	"github.com/BerylCAtieno/document-summarizer-api/internal/router"
	"github.com/BerylCAtieno/document-summarizer-api/internal/services"
	"github.com/BerylCAtieno/document-summarizer-api/internal/utils"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	logger := utils.NewLogger(cfg.LogLevel)

	// Initialize database
	database, err := db.NewPostgresDB(cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("Failed to connect to database", "error", err)
	}
	defer database.Close()

	// Run migrations
	if err := db.RunMigrations(cfg.DatabaseURL); err != nil {
		logger.Fatal("Failed to run migrations", "error", err)
	}

	// Initialize document service
	docRepo := repository.NewRepository(database)
	docService := services.NewService(docRepo, cfg, logger)

	// Setup HTTP router
	handler := router.NewRouter(docService, logger)

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server
	go func() {
		logger.Info("Starting server", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed to start", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", "error", err)
	}

	logger.Info("Server exited")
}
