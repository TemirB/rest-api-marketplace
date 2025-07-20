package post

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	postStorage "github.com/TemirB/rest-api-marketplace/internal/database/publication"
)

type Handler struct {
	Service *Service
}

func NewPostHandler(Service *Service) *Handler {
	return &Handler{Service: Service}
}

func (h *Handler) CreatePost(c *gin.Context) {
	owner := c.MustGet("userLogin").(string)

	var request struct {
		Title       string  `json:"title" binding:"required,max=200"`
		Description string  `json:"description" binding:"required,max=1000"`
		Price       float64 `json:"price" binding:"required,gt=0"`
		ImageURL    string  `json:"image_url" binding:"required,url"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	post, err := h.Service.CreatePost(&Post{
		Title:       request.Title,
		Description: request.Description,
		Price:       request.Price,
		ImageURL:    request.ImageURL,
		Owner:       owner,
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, post)
}

func (h *Handler) GetPosts(c *gin.Context, login string) {
	minPrice, _ := strconv.ParseFloat(c.Query("min_price"), 64)
	maxPrice, _ := strconv.ParseFloat(c.Query("max_price"), 64)

	sortBy := c.DefaultQuery("sort_by", "date")
	order := c.DefaultQuery("order", "desc")

	currentUser := ""
	if user, exists := c.Get("userLogin"); exists {
		currentUser = user.(string)
	}

	posts, err := h.Service.GetPosts(
		postStorage.SortParams{
			Field:     sortBy,
			Direction: order,
		},
		postStorage.FilterParams{
			MinPrice: minPrice,
			MaxPrice: maxPrice,
			Owner:    currentUser,
		},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, posts)
}
