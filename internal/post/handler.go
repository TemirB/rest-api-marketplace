package post

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type Handler struct {
	Service *Service
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

func (h *Handler) GetPosts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Тут нужен нормальный парсинг в случае отсутствия параметров
	q := r.URL.Query()
	minPrice, errMinPrice := strconv.ParseFloat(q.Get("min_price"), 64)
	maxPrice, errMaxPrice := strconv.ParseFloat(q.Get("max_price"), 64)
	switch {
	case errMinPrice == nil && errMaxPrice == nil:
	case errMinPrice != nil:
		minPrice = 0
	case errMaxPrice != nil:
		maxPrice = 100000 // сюда макс цену
	default:
		minPrice = 0
		maxPrice = 100000 // сюда макс цену
	}

	sortBy := q.Get("sort_by")
	if sortBy == "" {
		sortBy = "date"
	}
	order := q.Get("order")
	if order == "" {
		order = "desc"
	}

	var currentUser string
	if userVal := r.Context().Value("userLogin"); userVal != nil {
		if login, ok := userVal.(string); ok {
			currentUser = login
		}
	}

	posts, err := h.Service.GetPosts(
		SortParams{Field: sortBy, Direction: order},
		FilterParams{MinPrice: minPrice, MaxPrice: maxPrice, Owner: currentUser},
	)
	if err != nil {
		http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}
