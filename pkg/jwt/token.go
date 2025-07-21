package jwt

import (
	"errors"
	"net/http"
	"time"

	"github.com/TemirB/rest-api-marketplace/internal/middleware"
	jwt "github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
)

type Manager struct {
	secret     string
	expiration time.Duration
}

func New(secret string, expiration time.Duration) *Manager {
	return &Manager{
		secret:     secret,
		expiration: expiration,
	}
}

func (m *Manager) GenerateToken(login string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"login": login,
		"exp":   m.expiration,
	})
	return token.SignedString([]byte(m.secret))
}

func (m *Manager) ValidateToken(tokenStr string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return []byte(m.secret), nil
	})
	if err != nil || !token.Valid {
		return "", errors.New("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid claims")
	}
	login, ok := claims["login"].(string)
	if !ok {
		return "", errors.New("login claim missing")
	}
	return login, nil
}

func GetLogin(r *http.Request) (string, error) {
	if v := r.Context().Value(middleware.CtxUser); v != nil {
		if login, ok := v.(string); ok {
			return login, nil
		}
	}
	return "", ErrInvalidToken
}
