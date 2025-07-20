package main

import (
	"time"

	"go.uber.org/zap"

	auth "github.com/TemirB/rest-api-marketplace/internal/auth"
	"github.com/TemirB/rest-api-marketplace/internal/pkg/config"
	"github.com/TemirB/rest-api-marketplace/internal/pkg/db"
	"github.com/TemirB/rest-api-marketplace/internal/pkg/token"
	"github.com/TemirB/rest-api-marketplace/internal/storage"
)

func main() {
	// logger
	logger := zap.NewExample()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		zap.L().Fatal(
			"Failed to load configuration",
			zap.Error(err),
		)
	}

	// Initialize database connection
	dbRepo, err := db.NewPostgresDB(cfg, logger)
	if err != nil {
		logger.Fatal(
			"Failed to connect to database",
			zap.Error(err),
		)
	}
	if err := dbRepo.Connect(); err != nil {
		logger.Fatal(
			"Failed to connect to database",
			zap.Error(err),
		)
	}
	defer dbRepo.Close()

	// Initialize repositories
	userDB := storage.NewUserDatabase(dbRepo)
	postDB := storage.NewStorageDatabase(dbRepo)

	// Initialize services
	tokenGererator := token.NewGenerator(cfg.JWT.Secret, time.Duration(cfg.JWT.Expiration)*time.Minute)
	authService := auth.NewService(userDB, tokenGererator)

}
