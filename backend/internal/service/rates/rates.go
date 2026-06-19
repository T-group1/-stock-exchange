package rates

import (
	"context"
	"fmt"
	"time"

	"T_Project/internal/db"
	"T_Project/internal/parser/cbr"

	"github.com/jackc/pgx/v5/pgtype"
)

// Service отвечает за синхронизацию курсов с ЦБ РФ и сохранение в БД
type Service struct {
	cbrClient cbr.CBRClient // ✅ Используем интерфейс вместо конкретной реализации
	queries   db.Querier
}

// NewService создаёт новый сервис
func NewService(cbrClient cbr.CBRClient, queries db.Querier) *Service {
	return &Service{
		cbrClient: cbrClient,
		queries:   queries,
	}
}

// SyncCurrencies синхронизирует список валют с ЦБ РФ
func (s *Service) SyncCurrencies(ctx context.Context) error {
	currencies, err := s.cbrClient.GetAllCurrencies()
	if err != nil {
		return fmt.Errorf("failed to get currencies from CBR: %w", err)
	}

	for _, c := range currencies {
		_, err := s.queries.CreateCurrency(ctx, db.CreateCurrencyParams{
			Code:    c.CharCode,
			Name:    c.Name,
			Nominal: int32(c.Nominal),
		})
		if err != nil {
			return fmt.Errorf("failed to upsert currency %s: %w", c.CharCode, err)
		}
	}

	return nil
}

// SyncRates синхронизирует курсы на указанную дату
func (s *Service) SyncRates(ctx context.Context, date time.Time) error {
	rates, err := s.cbrClient.GetDailyRates(date)
	if err != nil {
		return fmt.Errorf("failed to get rates from CBR: %w", err)
	}

	for _, r := range rates {
		if r.CharCode == "RUB" {
			continue // RUB не сохраняем в таблицу currency_rates
		}

		// Гарантируем, что валюта существует в базе перед записью курса
		_, _ = s.queries.CreateCurrency(ctx, db.CreateCurrencyParams{
			Code:    r.CharCode,
			Name:    r.Name,
			Nominal: int32(r.Nominal),
		})

		// Конвертируем дату в pgtype.Date
		parsedDate, err := time.Parse("2006-01-02", r.Date)
		if err != nil {
			return fmt.Errorf("failed to parse date %s: %w", r.Date, err)
		}

		pgDate := pgtype.Date{}
		if err := pgDate.Scan(parsedDate); err != nil {
			return fmt.Errorf("failed to scan date: %w", err)
		}

		// Конвертируем rate в pgtype.Numeric через строку
		pgRate := pgtype.Numeric{}
		strRate := fmt.Sprintf("%f", r.Rate)
		if err := pgRate.Scan(strRate); err != nil {
			return fmt.Errorf("failed to scan rate: %w", err)
		}

		// Вычисляем change_percentage (передаем "0" как строку)
		changePct := pgtype.Numeric{}
		if err := changePct.Scan("0"); err != nil {
			return fmt.Errorf("failed to scan change percentage: %w", err)
		}

		_, err = s.queries.CreateRate(ctx, db.CreateRateParams{
			CurrencyCode:     r.CharCode,
			Rate:             pgRate,
			RateDate:         pgDate,
			Source:           "cb_rf",
			ChangePercentage: changePct,
		})
		if err != nil {
			return fmt.Errorf("failed to upsert rate for %s: %w", r.CharCode, err)
		}
	}

	return nil
}

// SyncToday синхронизирует и валюты, и курсы на сегодня
func (s *Service) SyncToday(ctx context.Context) error {
	if err := s.SyncCurrencies(ctx); err != nil {
		return err
	}
	return s.SyncRates(ctx, time.Time{}) // Пустая дата = сегодня
}
