package worker

import (
	"context"
	"log"
	"time"

	"T_Project/internal/service/rates"
)

// RatesWorker отвечает за периодическое обновление курсов
type RatesWorker struct {
	service  *rates.Service
	interval time.Duration
}

// NewRatesWorker создаёт новый воркер
func NewRatesWorker(service *rates.Service, interval time.Duration) *RatesWorker {
	return &RatesWorker{
		service:  service,
		interval: interval,
	}
}

// Start запускает воркер в фоне
func (w *RatesWorker) Start(ctx context.Context) {
	go w.run(ctx)
}

// run — основной цикл воркера
func (w *RatesWorker) run(ctx context.Context) {
	// Выполняем сразу при старте
	if err := w.sync(); err != nil {
		log.Printf("Initial sync failed: %v", err)
	}

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Rates worker stopped")
			return
		case <-ticker.C:
			if err := w.sync(); err != nil {
				log.Printf("Sync failed: %v", err)
			}
		}
	}
}

// sync выполняет одну итерацию синхронизации
func (w *RatesWorker) sync() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Println("Starting rates sync...")

	if err := w.service.SyncToday(ctx); err != nil {
		return err
	}

	log.Println("Rates sync completed successfully")
	return nil
}

// SyncNow позволяет вручную триггернуть синхронизацию
func (w *RatesWorker) SyncNow() error {
	return w.sync()
}
