package main

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"

	auth "github.com/TemirB/rest-api-marketplace/internal/auth"
	"github.com/TemirB/rest-api-marketplace/internal/config"
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
	http.Handle("/posts/", jwtMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// GET /posts/{id}
			postHandler.GetPostByID(w, r)
		case http.MethodPut:
			// PUT /posts/{id}
			postHandler.UpdatePost(w, r)
		case http.MethodDelete:
			// DELETE /posts/{id}
			postHandler.DeletePost(w, r)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})))
	http.HandleFunc("/posts/feed", postHandler.GetPosts)

	ServerAddress := ":" + strconv.Itoa(cfg.AppPort)
	log.Printf("Server started at %s\n", ServerAddress)
	http.ListenAndServe(ServerAddress, nil)
}
