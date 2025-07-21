package post

// mockgen  -source=handler.go -destination=handler_mock_test.go -package=post

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/TemirB/rest-api-marketplace/internal/middleware"
	"github.com/TemirB/rest-api-marketplace/pkg/jwt"
	"go.uber.org/zap"
)

type service interface {
	CreatePost(post *Post) (*Post, error)
	GetPosts(sort *SortParams, filter *FilterParams) ([]*Post, error)
	UpdatePost(post *Post) error
	DeletePost(id uint64) error

	GetPostByID(id uint) (*Post, error)
}

type Handler struct {
	service service
	logger  *zap.Logger
}

func NewHandler(service service, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

func (h *Handler) CreatePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.logger.Info(
			"Method not allowed",
			zap.String("method", r.Method),
		)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	loginVal, err := jwt.GetLogin(r)
	if err != nil {
		h.logger.Info(
			"Internal Server Error: invalid user context",
			zap.String("author", "jwt.GetLogin(r) == nil"),
		)
		http.Error(w, "Internal Server Error: invalid user context", http.StatusUnauthorized)
		return
	}

	var req struct {
		Title       string  `json:"title"`
		Description string  `json:"description"`
		Price       float64 `json:"price"`
		ImageURL    string  `json:"image_url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error(
			"Error while decoding request body",
			zap.String("author", loginVal),
			zap.Any("body", r.Body),
			zap.Error(err),
		)
		http.Error(w, "Bad Request: "+err.Error(), http.StatusBadRequest)
		return
	}

	newPost, err := h.service.CreatePost(&Post{
		Title:       req.Title,
		Description: req.Description,
		Price:       req.Price,
		ImageURL:    req.ImageURL,
		Owner:       loginVal,
	})
	if err != nil {
		h.logger.Error(
			"Failed to create post",
			zap.String("author", loginVal),
			zap.Any("post", newPost),
			zap.Error(err),
		)
		http.Error(w, "Bad Request: "+err.Error(), http.StatusBadRequest)
		return
	}

	h.logger.Info(
		"Post created successfully",
		zap.Any("post", newPost),
	)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newPost)
}

func setFilter(q url.Values, owner string) *FilterParams {
	var (
		minPrice float64
		maxPrice float64
		err      error
	)

	if minPrice, err = strconv.ParseFloat(q.Get("min_price"), 64); err != nil || minPrice < 0 {
		minPrice = 0
	}
	if maxPrice, err = strconv.ParseFloat(q.Get("max_price"), 64); err != nil || maxPrice < 0 {
		maxPrice = -1
	}

	if maxPrice >= 0 && minPrice > maxPrice {
		minPrice, maxPrice = maxPrice, minPrice
	}

	return &FilterParams{
		MinPrice: minPrice,
		MaxPrice: maxPrice,
		Owner:    owner,
	}
}

func setSort(q url.Values) *SortParams {
	sortBy := q.Get("sort_by")
	if strings.ToLower(sortBy) == "price" {
		sortBy = "price"
	} else {
		sortBy = "created_at"
	}
	order := q.Get("order")
	if strings.ToLower(order) == "asc" {
		order = "ASC"
	} else {
		order = "DESC"
	}
	return &SortParams{
		Field:     sortBy,
		Direction: order,
	}
}

func (h *Handler) GetPosts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.logger.Info(
			"Method not allowed",
			zap.String("method", r.Method),
		)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query()
	currentUser, _ := jwt.GetLogin(r)
	filter := setFilter(q, currentUser)
	sort := setSort(q)

	posts, err := h.service.GetPosts(sort, filter)
	if err != nil {
		h.logger.Error(
			"Failed to get posts",
			zap.String("author", currentUser),
			zap.Any("sort", sort),
			zap.Any("filter", filter),
			zap.Error(err),
		)
		http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Info(
		"Posts fetched successfully",
		zap.String("author", currentUser),
		zap.Any("sort", sort),
		zap.Any("filter", filter),
	)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

func (h *Handler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 3 || parts[1] != "posts" {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	id64, err := strconv.ParseUint(parts[2], 10, 64)
	if err != nil {
		http.Error(w, "Bad Request: invalid id", http.StatusBadRequest)
		return
	}
	id := uint(id64)

	var updatePostRequest *UpdatePostRequest
	if err := json.NewDecoder(r.Body).Decode(updatePostRequest); err != nil {
		http.Error(w, "Bad Request: "+err.Error(), http.StatusBadRequest)
		return
	}

	post, err := h.service.GetPostByID(id)
	if err != nil {
		h.logger.Error(
			"Failed to get post",
			zap.Uint("id", id),
			zap.Error(err),
		)
		http.Error(w, "Internal Server Error", http.StatusBadRequest)
		return
	}
	login, err := jwt.GetLogin(r)
	if err != nil {
		h.logger.Info(
			"Internal Server Error: invalid user context",
			zap.String("author", "jwt.GetLogin(r) == nil"),
		)
		http.Error(w, "Internal Server Error: invalid user context", http.StatusUnauthorized)
		return
	}
	post.IsOwner = false
	if post.Owner == login {
		post.IsOwner = true
	}

	mergePostUpdates(post, updatePostRequest)

	err = validatePost(post)
	if err != nil {
		http.Error(w, "Bad Request: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.UpdatePost(post); err != nil {
		h.logger.Error("Failed to update post", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func mergePostUpdates(post *Post, updatePostRequest *UpdatePostRequest) {
	if updatePostRequest.Title != nil {
		post.Title = *updatePostRequest.Title
	}
	if updatePostRequest.Description != nil {
		post.Description = *updatePostRequest.Description
	}
	if updatePostRequest.Price != nil {
		post.Price = *updatePostRequest.Price
	}
	if updatePostRequest.ImageURL != nil {
		post.ImageURL = *updatePostRequest.ImageURL
	}
}

func (h *Handler) DeletePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 3 || parts[1] != "posts" {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	id64, err := strconv.ParseUint(parts[2], 10, 32)
	if err != nil {
		http.Error(w, "Bad Request: invalid id", http.StatusBadRequest)
		return
	}

	post, err := h.service.GetPostByID(uint(id64))
	if err != nil {
		h.logger.Error(
			"Failed to get post",
			zap.Uint64("id", id64),
			zap.Error(err),
		)
		http.Error(w, "Post doesnt exist", http.StatusBadRequest)
		return
	}
	if post.Owner != r.Context().Value(middleware.CtxUser) {
		http.Error(w, "Unauthorized: not owner", http.StatusUnauthorized)
		return
	}

	if err := h.service.DeletePost(id64); err != nil {
		h.logger.Error("Failed to delete post", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetPostByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 3 || parts[1] != "posts" {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	id64, err := strconv.ParseUint(parts[2], 10, 64)
	if err != nil {
		http.Error(w, "Bad Request: invalid id", http.StatusBadRequest)
		return
	}

	id := uint(id64)
	post, err := h.service.GetPostByID(id)
	if err != nil {
		if errors.Is(err, ErrPostNotFound) {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		h.logger.Error(
			"Failed to get post",
			zap.Uint("id", id),
			zap.Error(err),
		)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	login, err := jwt.GetLogin(r)
	if err != nil {
		h.logger.Info(
			"Internal Server Error: invalid user context",
			zap.String("author", "jwt.GetLogin(r) == nil"),
		)
		http.Error(w, "Internal Server Error: invalid user context", http.StatusUnauthorized)
		return
	}
	post.IsOwner = (login != "" && login == post.Owner)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}
