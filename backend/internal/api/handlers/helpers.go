package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"T_Project/internal/db"

	"github.com/go-chi/chi/v5"
)

// rateInfo — нормализованная информация о курсе
type rateInfo struct {
	CurrencyCode string
	Rate         float64
	Date         string
	Source       string
}

// getLatestRateMap возвращает map[CharCode]rateInfo для всех последних курсов
func getLatestRateMap(ctx context.Context, queries db.Querier) (map[string]rateInfo, string, error) {
	rates, err := queries.GetLatestRates(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch rates: %w", err)
	}

	rateMap := make(map[string]rateInfo, len(rates))
	var latestDate string

	for _, rate := range rates {
		rateFloat, err := rate.Rate.Float64Value()
		if err != nil || !rateFloat.Valid {
			continue
		}

		dateStr := rate.RateDate.Time.Format("2006-01-02")
		if latestDate == "" {
			latestDate = dateStr
		}

		rateMap[rate.CurrencyCode] = rateInfo{
			CurrencyCode: rate.CurrencyCode,
			Rate:         rateFloat.Float64,
			Date:         dateStr,
			Source:       rate.Source,
		}
	}

	return rateMap, latestDate, nil
}

// extractPairFromURL извлекает пару валют из URL path через chi
func extractPairFromURL(r *http.Request) string {
	return chi.URLParam(r, "pair")
}

// parsePair парсит строку пары валют
func parsePair(pair string) (string, string, error) {
	parts := strings.Split(pair, "_")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid pair format: %q, expected BASE_QUOTE", pair)
	}
	base := strings.TrimSpace(parts[0])
	quote := strings.TrimSpace(parts[1])
	if base == "" || quote == "" {
		return "", "", fmt.Errorf("empty currency code in pair: %q", pair)
	}
	return strings.ToUpper(base), strings.ToUpper(quote), nil
}
