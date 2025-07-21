package post

// mockgen  -source=handler.go -destination=handler_mock_test.go -package=post

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

type service interface {
	CreatePost(post *Post) (*Post, error)
	GetPosts(sort *SortParams, filter *FilterParams) ([]*Post, error)
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

	loginVal := r.Context().Value("userLogin")
	if loginVal == nil {
		h.logger.Info(
			"Unauthorized: no author",
			zap.String("author", "loginVal==nil"),
		)
		http.Error(w, "Unauthorized: no user", http.StatusUnauthorized)
		return
	}
	ownerLogin, ok := loginVal.(string)
	if !ok {
		h.logger.Info(
			"Internal Server Error: invalid user context",
			zap.String("author", "loginVal.(string) == false"),
		)
		http.Error(w, "Internal Server Error: invalid user context", http.StatusInternalServerError)
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
			zap.String("author", ownerLogin),
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
		Owner:       ownerLogin,
	})
	if err != nil {
		h.logger.Error(
			"Failed to create post",
			zap.String("author", ownerLogin),
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
	var currentUser string
	if userVal := r.Context().Value("userLogin"); userVal != nil {
		if login, ok := userVal.(string); ok {
			currentUser = login
		}
	}
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
