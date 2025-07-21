package auth

// mockgen  -source=handler.go -destination=handler_mock_test.go -package=auth

import (
	"encoding/json"
	"errors"
	"net/http"

	"go.uber.org/zap"
)

type service interface {
	Register(login, password string) error
	Login(login, password string) (string, error)
}

type Handler struct {
	Service service
	logger  *zap.Logger
}

func NewHandler(service service, logger *zap.Logger) *Handler {
	return &Handler{
		Service: service,
		logger:  logger,
	}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.logger.Info(
			"Method not allowed",
			zap.String("method", r.Method),
		)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error(
			"Error while decoding request body",
			zap.Any("body", r.Body),
			zap.Error(err),
		)
		http.Error(w, "Bad Request: "+err.Error(), http.StatusBadRequest)
		return
	}

	err := h.Service.Register(req.Login, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserExists):
			http.Error(w, "User already exists", http.StatusConflict)
			return
		case errors.Is(err, ErrInvalidLogin),
			errors.Is(err, ErrInvalidPassword):
			http.Error(w, "Bad Request: "+err.Error(), http.StatusBadRequest)
			return
		default:
			h.logger.Error("Register failed", zap.Error(err))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	h.logger.Info(
		"User registered successfully",
		zap.String("login", req.Login),
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"login": req.Login,
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.logger.Info(
			"Method not allowed",
			zap.String("method", r.Method),
		)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error(
			"Error while decoding request body",
			zap.Any("body", r.Body),
			zap.Error(err),
		)
		http.Error(w, "Bad Request: "+err.Error(), http.StatusBadRequest)
		return
	}

	token, err := h.Service.Login(req.Login, req.Password)
	if err != nil {
		h.logger.Error(
			"Error while logging in user",
			zap.String("login", req.Login),
			zap.Error(err),
		)
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	h.logger.Info(
		"User logged in successfully",
		zap.String("login", req.Login),
	)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token": token,
	})
}
