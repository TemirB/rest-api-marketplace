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

	// Initialize handlers
	authHandler := auth.NewHandler(authService, logger)
	postHandler := post.NewHandler(postService, logger)

	// Set up HTTP server and routes
	mux := http.NewServeMux()

	mux.HandleFunc("/register", authHandler.Register)
	mux.HandleFunc("/login", authHandler.Login)

	mux.Handle("/posts", middleware.JWTAuthMiddleware(authService)(
		http.HandlerFunc(postHandler.CreatePost),
	))
	mux.Handle("/posts/feed", middleware.OptionalAuthMiddleware(authService)(
		http.HandlerFunc(postHandler.GetPosts),
	))

	mux.Handle("/posts/", middleware.OptionalAuthMiddleware(authService)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				postHandler.GetPostByID(w, r)

			case http.MethodPut:
				// проверяем, что JWT был валидным
				if r.Context().Value(middleware.CtxUser) == nil {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
				postHandler.UpdatePost(w, r)

			case http.MethodDelete:
				if r.Context().Value(middleware.CtxUser) == nil {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
				postHandler.DeletePost(w, r)

			default:
				http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			}
		}),
	))

	ServerAddress := ":" + strconv.Itoa(cfg.AppPort)
	log.Printf("Server started at %s\n", ServerAddress)
	http.ListenAndServe(ServerAddress, mux)
}
