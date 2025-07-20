package auth

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Service *Service
	logger  *zap.Logger
}

func NewHandler(Service *Service, logger *zap.Logger) *Handler {
	return &Handler{
		Service: Service,
		logger:  logger,
	}
}

var registrationRequest struct {
	Login    string `json:"login" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=8"`
}

var LoginRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *Handler) Register(c *gin.Context) {
	if err := c.ShouldBindJSON(&registrationRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.Service.Register(registrationRequest.Login, registrationRequest.Password); err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": err.Error()},
		)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"login":    registrationRequest.Login,
		"password": registrationRequest.Password,
		// По условию нужно вернуть данные созданного пользователя,
		// но мне кажется, что возвращать пароль не ОК
	})
}

func (h *Handler) Login(c *gin.Context) {
	if err := c.ShouldBindJSON(&LoginRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.Service.Login(LoginRequest.Login, LoginRequest.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}
