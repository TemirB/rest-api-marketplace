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

func JWTAuthMiddleware(authService authService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			login, err := authService.ValidateToken(tokenStr)
			if err != nil {
				http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), CtxUser, login)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
func OptionalAuthMiddleware(authService authService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
				if login, err := authService.ValidateToken(tokenStr); err == nil {
					r = r.WithContext(context.WithValue(r.Context(), CtxUser, login))
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
