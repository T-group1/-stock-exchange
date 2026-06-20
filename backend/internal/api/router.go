package api

import (
	"net/http"

	"T_Project/internal/api/handlers"
	"T_Project/internal/api/middleware"
	"T_Project/internal/config"
	"T_Project/internal/db"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Router создаёт и настраивает маршрутизатор
func Router(queries db.Querier, cfg *config.Config) http.Handler {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	// CORS middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	// Публичные маршруты (без авторизации)
	r.Post("/auth/register", handlers.NewAuthHandler(queries, cfg).Register)
	r.Post("/auth/login", handlers.NewAuthHandler(queries, cfg).Login)
	r.Post("/auth/verify", handlers.NewAuthHandler(queries, cfg).Verify)
	r.Post("/auth/refresh", handlers.NewAuthHandler(queries, cfg).Refresh)

	// Защищённые маршруты (требуют авторизации)
	r.Group(func(r chi.Router) {
		r.Use(middleware.JWTAuthMiddleware(cfg.JWT.Secret))

		// Auth
		r.Get("/auth/me", handlers.NewAuthHandler(queries, cfg).Me)

		// Currencies
		r.Get("/currencies", handlers.NewCurrenciesHandler(queries).List)

		// Rates
		r.Get("/rates", handlers.NewRatesHandler(queries).Latest)
		r.Get("/rates/{pair}", handlers.NewRatesHandler(queries).History)
		r.Post("/convert", handlers.NewConversionHandler(queries).Convert)

		// Subscriptions
		r.Route("/subscriptions", func(r chi.Router) {
			r.Get("/", handlers.NewSubscriptionsHandler(queries).List)
			r.Post("/", handlers.NewSubscriptionsHandler(queries).Create)
			r.Delete("/{id}", handlers.NewSubscriptionsHandler(queries).Delete)
		})

		// Notifications
		// ИСПРАВЛЕНО: literal routes идут ПЕРЕД parameterized
		r.Route("/notifications", func(r chi.Router) {
			r.Get("/", handlers.NewNotificationsHandler(queries).List)
			r.Get("/unread-count", handlers.NewNotificationsHandler(queries).GetUnreadCount)
			r.Route("/settings", func(r chi.Router) {
				r.Get("/", handlers.NewNotificationSettingsHandler(queries).Get)
				r.Put("/", handlers.NewNotificationSettingsHandler(queries).Update)
			})
			// Parameterized route в конце
			r.Post("/{id}/read", handlers.NewNotificationsHandler(queries).MarkAsRead)
		})

		// User pairs
		r.Get("/user/pairs", handlers.NewUserPairsHandler(queries).Get)
		r.Put("/user/pairs", handlers.NewUserPairsHandler(queries).Update)

		// Favorites
		r.Route("/favorites", func(r chi.Router) {
			r.Get("/", handlers.NewFavoritesHandler(queries).List)
			r.Post("/{pair}", handlers.NewFavoritesHandler(queries).Add)
			r.Delete("/{pair}", handlers.NewFavoritesHandler(queries).Remove)
		})
	})

	return r
}
