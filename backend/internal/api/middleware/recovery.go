package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"T_Project/internal/api/response"
)

// Recovery middleware для восстановления от паник
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("PANIC: %v\n%s", err, debug.Stack())
				response.InternalError(w, "Internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
