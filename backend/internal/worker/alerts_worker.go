package worker

import (
	"context"
	"fmt"
	"log"
	"time"

	"T_Project/internal/db"
	"T_Project/internal/service/auth"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// AlertsWorker воркер для проверки алертов
type AlertsWorker struct {
	queries      db.Querier
	emailService *auth.EmailService
	interval     time.Duration
}

// NewAlertsWorker создаёт новый воркер для проверки алертов
func NewAlertsWorker(queries db.Querier, emailService *auth.EmailService, interval time.Duration) *AlertsWorker {
	return &AlertsWorker{
		queries:      queries,
		emailService: emailService,
		interval:     interval,
	}
}

// Start запускает воркер
func (w *AlertsWorker) Start(ctx context.Context) {
	log.Println("Alerts worker started")

	// ИСПРАВЛЕНИЕ ОШИБКИ #6: Первичная проверка сразу при старте
	if err := w.checkAlerts(ctx); err != nil {
		log.Printf("Error in initial alerts check: %v", err)
	}

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
	// ИСПРАВЛЕНИЕ ОШИБКИ #5: Проверяем, существует ли пользователь и верифицирован ли email
	user, err := w.queries.GetUserByID(ctx, sub.UserID)
	if err != nil {
		if err == pgx.ErrNoRows {
			log.Printf("User %s not found, skipping notification", sub.UserID.String())
			return nil
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Проверяем, верифицирован ли email
	if !user.IsVerified.Bool {
		log.Printf("User %s email not verified, skipping email notification", user.ID.String())
		// Продолжаем создавать уведомление в БД, но не отправляем email
	}

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

	// ИСПРАВЛЕНИЕ ОШИБКИ #3 и #4: Отправляем email, если настройки позволяют
	if user.IsVerified.Bool {
		err = w.sendEmailIfAllowed(ctx, user, title, message)
		if err != nil {
			log.Printf("Failed to send email notification: %v", err)
			// Не прерываем выполнение, уведомление уже создано в БД
		}
	}

	return nil
}

// sendEmailIfAllowed проверяет настройки и отправляет email, если разрешено
func (w *AlertsWorker) sendEmailIfAllowed(ctx context.Context, user db.User, title, message string) error {
	// Получаем настройки уведомлений пользователя
	settings, err := w.queries.GetNotificationSettings(ctx, user.ID)
	if err != nil {
		if err == pgx.ErrNoRows {
			// Если настроек нет, используем значения по умолчанию (email включён)
			log.Printf("No notification settings for user %s, using defaults", user.ID.String())
			return w.emailService.SendAlertEmail(user.Email, title, message)
		}
		return fmt.Errorf("failed to get notification settings: %w", err)
	}

	// Проверяем, включены ли email уведомления
	if !settings.EmailEnabled.Bool {
		log.Printf("Email notifications disabled for user %s", user.ID.String())
		return nil
	}

	// ИСПРАВЛЕНИЕ ОШИБКИ #4: Проверяем тихие часы
	if w.isQuietHours(settings) {
		log.Printf("Quiet hours active for user %s, skipping email notification", user.ID.String())
		return nil
	}

	// Отправляем email
	return w.emailService.SendAlertEmail(user.Email, title, message)
}

// isQuietHours проверяет, находится ли текущее время в тихих часах
func (w *AlertsWorker) isQuietHours(settings db.NotificationSetting) bool {
	// Если тихие часы не настроены, возвращаем false
	if !settings.QuietHoursStart.Valid || !settings.QuietHoursEnd.Valid {
		return false
	}

	if settings.QuietHoursStart.String == "" || settings.QuietHoursEnd.String == "" {
		return false
	}

	// Получаем текущее время
	now := time.Now()
	currentTime := now.Format("15:04")

	// Парсим время начала и окончания тихих часов
	startTime := settings.QuietHoursStart.String
	endTime := settings.QuietHoursEnd.String

	// Проверяем, попадает ли текущее время в диапазон
	// Учитываем случай, когда тихие часы переходят через полночь (например, 22:00 - 08:00)
	if startTime <= endTime {
		// Обычный случай: например, 09:00 - 18:00
		return currentTime >= startTime && currentTime <= endTime
	} else {
		// Случай через полночь: например, 22:00 - 08:00
		return currentTime >= startTime || currentTime <= endTime
	}
}
