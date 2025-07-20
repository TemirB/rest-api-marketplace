package post

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

type Handler struct {
	Service *Service
	Logger  *zap.Logger
}

func NewHandler(service *Service, logger *zap.Logger) *Handler {
	return &Handler{
		Service: service,
		Logger:  logger,
	}
}

func (h *Handler) CreatePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	loginVal := r.Context().Value("userLogin")
	if loginVal == nil {
		http.Error(w, "Unauthorized: no user", http.StatusUnauthorized)
		return
	}
	ownerLogin, ok := loginVal.(string)
	if !ok {
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
		http.Error(w, "Bad Request: "+err.Error(), http.StatusBadRequest)
		return
	}

	newPost, err := h.Service.CreatePost(&Post{
		Title:       req.Title,
		Description: req.Description,
		Price:       req.Price,
		ImageURL:    req.ImageURL,
		Owner:       ownerLogin,
	})
	if err != nil {
		http.Error(w, "Bad Request: "+err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newPost)
}

func setFilter(q url.Values, owner string) *FilterParams {
	minPrice, errMinPrice := strconv.ParseFloat(q.Get("min_price"), 64)
	maxPrice, errMaxPrice := strconv.ParseFloat(q.Get("max_price"), 64)
	switch {
	case minPrice > maxPrice || maxPrice < 0 || minPrice < 0:
		minPrice = 0
		maxPrice = -1
	case errMinPrice == nil && errMaxPrice == nil:
	case errMinPrice != nil:
		minPrice = 0
	case errMaxPrice != nil:
		maxPrice = -1
	default:
		minPrice = 0
		maxPrice = -1
	}
	return &FilterParams{
		MinPrice: minPrice,
		MaxPrice: maxPrice,
		Owner:    owner,
	}
}

func setSort(q url.Values) *SortParams {
	sortBy := q.Get("sort_by")
	if strings.ToLower(sortBy) != "created_at" {
		sortBy = "price"
	} else {
		sortBy = "created_at"
	}
	order := q.Get("order")
	if strings.ToLower(order) != "desc" {
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
	if filter == nil {
		return
	}
	sort := setSort(q)

	posts, err := h.Service.GetPosts(sort, filter)
	if err != nil {
		http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}
