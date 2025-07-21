package main

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"

	"github.com/TemirB/rest-api-marketplace/config"
	auth "github.com/TemirB/rest-api-marketplace/internal/auth"
	"github.com/TemirB/rest-api-marketplace/internal/database"
	"github.com/TemirB/rest-api-marketplace/internal/middleware"
	post "github.com/TemirB/rest-api-marketplace/internal/post"
	"github.com/TemirB/rest-api-marketplace/pkg/jwt"
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
	dbRepo, err := database.NewPostgresDB(cfg, logger)
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

	// Initialize storages
	userDB := auth.NewStorage(dbRepo, logger)
	postDB := post.NewStorage(dbRepo, logger)

	// Initialize services
	tokemManager := jwt.New(cfg.JWT.Secret, time.Duration(cfg.JWT.Expiration)*time.Minute)
	authService := auth.NewService(userDB, tokemManager, logger)
	postService := post.NewService(postDB, logger)

	// Initialize middleware
	jwtMiddleware := middleware.JWTAuthMiddleware(authService)

	// Initialize handlers
	authHandler := auth.NewHandler(authService, logger)
	postHandler := post.NewHandler(postService, logger)

	// Set up HTTP server and routes
	http.HandleFunc("/register", authHandler.Register)
	http.HandleFunc("/login", authHandler.Login)

	http.Handle("/posts", jwtMiddleware(http.HandlerFunc(postHandler.CreatePost)))
	http.HandleFunc("/posts/feed", postHandler.GetPosts)

	ServerAddress := "localhost:" + strconv.Itoa(cfg.AppPort)
	log.Printf("Server started at %s\n", ServerAddress)
	http.ListenAndServe(ServerAddress, nil)
}
