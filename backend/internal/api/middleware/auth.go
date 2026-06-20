package middleware

import (
	"context"
	"net/http"
	"strings"

	"T_Project/internal/api/response"
	"T_Project/internal/service/auth"
)

type contextKey string

const UserIDContextKey contextKey = "user_id"

// Auth middleware проверяет JWT токен и кладёт user_id в context
func Auth(jwtService *auth.JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", http.StatusText(http.StatusUnauthorized), "Authorization header is required")
				return
			}

			// Ожидаем формат: "Bearer <token>"
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				response.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", http.StatusText(http.StatusUnauthorized), "Invalid authorization header format")
				return
			}

			token := parts[1]
			claims, err := jwtService.ValidateToken(token)
			if err != nil {
				response.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", http.StatusText(http.StatusUnauthorized), "Invalid or expired token")
				return
			}

			// Кладём user_id в context
			userID, ok := claims["sub"].(string)
			if !ok {
				response.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", http.StatusText(http.StatusUnauthorized), "Invalid token claims")
				return
			}

			ctx := context.WithValue(r.Context(), UserIDContextKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserIDFromContext извлекает user_id из context
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDContextKey).(string)
	return userID, ok
}
