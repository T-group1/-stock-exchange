package worker

import (
	"context"
	"fmt"
	"log"
	"time"

	"T_Project/internal/db"

	"github.com/jackc/pgx/v5/pgtype"
)

// AlertsWorker воркер для проверки алертов
type AlertsWorker struct {
	queries  db.Querier
	interval time.Duration
}

// NewAlertsWorker создаёт новый воркер для проверки алертов
func NewAlertsWorker(queries db.Querier, interval time.Duration) *AlertsWorker {
	return &AlertsWorker{
		queries:  queries,
		interval: interval,
	}
}

// Start запускает воркер
func (w *AlertsWorker) Start(ctx context.Context) {
	log.Println("Alerts worker started")
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Alerts worker stopped")
			return
		case <-ticker.C:
			if err := w.checkAlerts(ctx); err != nil {
				log.Printf("Error checking alerts: %v", err)
			}
		}
	}
}

// checkAlerts проверяет все активные подписки
func (w *AlertsWorker) checkAlerts(ctx context.Context) error {
	// Получаем все активные подписки
	subscriptions, err := w.queries.GetActiveSubscriptions(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active subscriptions: %w", err)
	}

	if len(subscriptions) == 0 {
		return nil
	}

	log.Printf("Checking %d active subscriptions", len(subscriptions))

	// Группируем подписки по валюте для оптимизации
	currencySubs := make(map[string][]db.Subscription)
	for _, sub := range subscriptions {
		currencySubs[sub.CurrencyCode] = append(currencySubs[sub.CurrencyCode], sub)
	}

	// Для каждой валюты получаем текущий курс и проверяем подписки
	for currencyCode, subs := range currencySubs {
		rate, err := w.queries.GetLatestRateByCurrency(ctx, currencyCode)
		if err != nil {
			log.Printf("Failed to get rate for %s: %v", currencyCode, err)
			continue
		}

		// Получаем числовое значение курса
		rateValue, err := rate.Rate.Float64Value()
		if err != nil || !rateValue.Valid {
			continue
		}

		currentRate := rateValue.Float64

		// Проверяем каждую подписку для этой валюты
		for _, sub := range subs {
			subRateValue, err := sub.RateValue.Float64Value()
			if err != nil || !subRateValue.Valid {
				continue
			}

			targetRate := subRateValue.Float64
			triggered := false

			// Проверяем условие
			if sub.Condition == "above" && currentRate >= targetRate {
				triggered = true
			} else if sub.Condition == "below" && currentRate <= targetRate {
				triggered = true
			}

			if triggered {
				log.Printf("Alert triggered for user %s: %s %s %v (current: %v)",
					sub.UserID.String(), sub.CurrencyCode, sub.Condition, targetRate, currentRate)

				// Создаём уведомление
				err = w.createNotification(ctx, sub, currentRate)
				if err != nil {
					log.Printf("Failed to create notification: %v", err)
					continue
				}

				// Деактивируем подписку
				err = w.queries.DeactivateSubscription(ctx, db.DeactivateSubscriptionParams{
					ID: sub.ID,
					TriggeredAt: pgtype.Int8{
						Int64: time.Now().Unix(),
						Valid: true,
					},
				})
				if err != nil {
					log.Printf("Failed to deactivate subscription: %v", err)
				}
			}
		}
	}

	return nil
}

// createNotification создаёт уведомление для пользователя
func (w *AlertsWorker) createNotification(ctx context.Context, sub db.Subscription, currentRate float64) error {
	// Получаем числовое значение целевого курса
	subRateValue, err := sub.RateValue.Float64Value()
	if err != nil || !subRateValue.Valid {
		return fmt.Errorf("failed to get rate value: %w", err)
	}
	targetRate := subRateValue.Float64

	// Формируем сообщение
	title := fmt.Sprintf("Алерт сработал для %s", sub.CurrencyCode)
	message := fmt.Sprintf(
		"Курс %s достиг %s %.2f (текущий курс: %.2f)",
		sub.CurrencyCode,
		map[string]string{"above": "выше", "below": "ниже"}[sub.Condition],
		targetRate,
		currentRate,
	)

	// Создаём уведомление в БД (ID генерируется автоматически)
	_, err = w.queries.CreateNotification(ctx, db.CreateNotificationParams{
		UserID:         sub.UserID,
		SubscriptionID: sub.ID,
		Type:           "rate_alert",
		Title:          title,
		Message:        message,
	})

	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	return nil
}
