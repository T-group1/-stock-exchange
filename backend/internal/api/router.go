package api

import (
	"net/http"
	"time"

	"T_Project/internal/api/handlers"
	customMiddleware "T_Project/internal/api/middleware"
	"T_Project/internal/config"
	"T_Project/internal/db"
	"T_Project/internal/service/auth"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

// Router настраивает и возвращает HTTP роутер
func Router(queries db.Querier, cfg *config.Config) http.Handler {
	r := chi.NewRouter()

	// Chi middleware
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.Compress(5))
	r.Use(chiMiddleware.Timeout(30 * time.Second))

	// Custom middleware
	r.Use(customMiddleware.CORS)

	// Создаём хендлеры
	authHandler := handlers.NewAuthHandler(queries, cfg)
	currenciesHandler := handlers.NewCurrenciesHandler(queries)
	ratesHandler := handlers.NewRatesHandler(queries)
	convertHandler := handlers.NewConvertHandler(queries)

	// Создаём JWT сервис для auth middleware
	jwtService := auth.NewJWTService(
		cfg.JWT.Secret,
		cfg.JWT.AccessTokenExpiry,
		cfg.JWT.RefreshTokenExpiry,
	)

	// Создаём хендлеры для алертов
	notificationsHandler := handlers.NewNotificationsHandler(queries)
	subscriptionsHandler := handlers.NewSubscriptionsHandler(queries)
	settingsHandler := handlers.NewNotificationSettingsHandler(queries)

	// Auth endpoints (публичные)
	r.Post("/auth/register", authHandler.Register)
	r.Post("/auth/login", authHandler.Login)
	// r.Post("/auth/verify", authHandler.Verify)

	// Публичные эндпоинты
	r.Route("/currencies", func(r chi.Router) {
		r.Get("/", currenciesHandler.List)
		r.Get("/{code}", currenciesHandler.GetByCode)
	})

	r.Route("/rates", func(r chi.Router) {
		r.Get("/", ratesHandler.GetAll)
		r.Get("/{pair}", ratesHandler.GetByPair)
		r.Get("/{pair}/history", ratesHandler.GetHistory)
	})

	r.Post("/convert", convertHandler.Convert)

	// Защищённые эндпоинты (требуют авторизации)
	r.Group(func(r chi.Router) {
		r.Use(customMiddleware.Auth(jwtService))

		r.Route("/notifications", func(r chi.Router) {
			r.Get("/", notificationsHandler.List)
			r.Get("/unread-count", notificationsHandler.GetUnreadCount)
			r.Post("/{id}/read", notificationsHandler.MarkAsRead)
		})

		r.Route("/subscriptions", func(r chi.Router) {
			r.Get("/", subscriptionsHandler.List)
			r.Post("/", subscriptionsHandler.Create)
			r.Delete("/{id}", subscriptionsHandler.Delete)
		})

		r.Route("/notifications/settings", func(r chi.Router) {
			r.Get("/", settingsHandler.Get)
			r.Put("/", settingsHandler.Update)
		})
	})

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	return r
}
