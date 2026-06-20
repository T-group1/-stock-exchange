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
	"T_Project/internal/parser/cbr"
	"T_Project/internal/service/auth"
	"T_Project/internal/service/rates"
	"T_Project/internal/worker"

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

	// Инициализируем воркер для сбора данных
	cbrClient := cbr.NewClient()
	ratesService := rates.NewService(cbrClient, queries)
	ratesWorker := worker.NewRatesWorker(ratesService, 1*time.Hour)

	// Запускаем воркер в фоне
	workerCtx, workerCancel := context.WithCancel(ctx)
	defer workerCancel()
	ratesWorker.Start(workerCtx)
	log.Println("Rates worker started (interval: 1h)")

	// ИСПРАВЛЕНИЕ: Создаём emailService для воркера алертов
	emailService := auth.NewEmailService(
		cfg.SMTP.Host,
		cfg.SMTP.Port,
		cfg.SMTP.Username,
		cfg.SMTP.Password,
		cfg.SMTP.From,
		cfg.FrontendURL,
	)

	// Запускаем воркер алертов (проверяет каждые 30 секунд)
	alertsWorker := worker.NewAlertsWorker(queries, emailService, 30*time.Second)
	go alertsWorker.Start(workerCtx)
	log.Println("Alerts worker started (interval: 30s)")

	// Выполняем первичную синхронизацию
	log.Println("Running initial sync...")
	if err := ratesService.SyncToday(ctx); err != nil {
		log.Printf("Initial sync failed (will retry in background): %v", err)
	} else {
		log.Println("Initial sync completed successfully")
	}

	// Создаём роутер
	router := api.Router(queries, cfg)

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

	// Останавливаем воркер
	workerCancel()

	// Graceful shutdown с таймаутом
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}
