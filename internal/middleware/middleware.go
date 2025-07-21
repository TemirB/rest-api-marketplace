package middleware

import (
	"context"
	"net/http"
	"strings"
)

type ctxKey string

const CtxUser ctxKey = "userLogin"

type authService interface {
	ValidateToken(tokenStr string) (string, error)
}

func JWTAuthMiddleware(s authService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			if !strings.HasPrefix(h, "Bearer ") {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			token := strings.TrimPrefix(h, "Bearer ")
			login, err := s.ValidateToken(token)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), CtxUser, login)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func OptionalAuthMiddleware(s authService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			if strings.HasPrefix(h, "Bearer ") {
				if login, err := s.ValidateToken(strings.TrimPrefix(h, "Bearer ")); err == nil {
					r = r.WithContext(context.WithValue(r.Context(), CtxUser, login))
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
