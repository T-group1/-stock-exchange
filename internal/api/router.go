package api

import (
	"net/http"

	"T_Project/internal/api/handlers"
	customMiddleware "T_Project/internal/api/middleware"
	"T_Project/internal/db"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

// Router настраивает и возвращает HTTP роутер
func Router(queries db.Querier) http.Handler {
	r := chi.NewRouter()

	// Chi middleware
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.Compress(5))
	r.Use(chiMiddleware.Timeout(30))

	// Custom middleware
	r.Use(customMiddleware.CORS)

	// Создаём хендлеры
	currenciesHandler := handlers.NewCurrenciesHandler(queries)
	ratesHandler := handlers.NewRatesHandler(queries)
	convertHandler := handlers.NewConvertHandler(queries)

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

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	return r
}
