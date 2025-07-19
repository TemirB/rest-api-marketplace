package auth

import (
	"github.com/gin-gonic/gin"
)

func GenerateJWT(jwtSecret, login string) (string, error) gin.HandlerFunc {

}