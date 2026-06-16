package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"T_Project/internal/api"
	"T_Project/internal/config"
	"T_Project/internal/db"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Загружаем конфигурацию
	cfg := config.Load()

	// Подключаемся к БД
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.Database.DSN())
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer pool.Close()

	// Проверяем подключение
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Unable to ping database: %v", err)
	}
	log.Println("Successfully connected to database")

	// Создаём sqlc queries
	queries := db.New(pool)

	// Создаём роутер
	router := api.Router(queries)

	// Создаём HTTP сервер
	srv := &http.Server{
		Addr:         cfg.Server.Addr(),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Канал для сигналов graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Запускаем сервер в горутине
	go func() {
		log.Printf("Server starting on %s", cfg.Server.Addr())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Ждём сигнал завершения
	<-quit
	log.Println("Server is shutting down...")

	// Graceful shutdown с таймаутом
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}
