package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"T_Project/internal/db"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// rateInfo — нормализованная информация о курсе
type rateInfo struct {
	CurrencyCode     string
	Rate             float64
	Date             string
	Source           string
	ChangePercentage float64 // <-- ДОБАВЛЕНО: для устранения N+1 запросов
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

		// Безопасное извлечение change_percentage
		var changePct float64
		changePctVal, err := rate.ChangePercentage.Float64Value()
		if err == nil && changePctVal.Valid {
			changePct = changePctVal.Float64
		}

		rateMap[rate.CurrencyCode] = rateInfo{
			CurrencyCode:     rate.CurrencyCode,
			Rate:             rateFloat.Float64,
			Date:             dateStr,
			Source:           rate.Source,
			ChangePercentage: changePct, // <-- ДОБАВЛЕНО
		}
	}

	return rateMap, latestDate, nil
}

// getRateMapByDate возвращает map[CharCode]rateInfo для всех курсов на указанную дату
func getRateMapByDate(ctx context.Context, queries db.Querier, date time.Time) (map[string]rateInfo, error) {
	pgDate := pgtype.Date{
		Time:  date,
		Valid: true,
	}

	rates, err := queries.GetRatesByDate(ctx, pgDate)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch rates by date: %w", err)
	}

	rateMap := make(map[string]rateInfo, len(rates))
	for _, rate := range rates {
		rateFloat, err := rate.Rate.Float64Value()
		if err != nil || !rateFloat.Valid {
			continue
		}

		dateStr := rate.RateDate.Time.Format("2006-01-02")

		rateMap[rate.CurrencyCode] = rateInfo{
			CurrencyCode: rate.CurrencyCode,
			Rate:         rateFloat.Float64,
			Date:         dateStr,
			Source:       rate.Source,
		}
	}

	return rateMap, nil
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

// respondWithJSON отправляет JSON-ответ с указанным статусом
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "Failed to marshal JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// parseJSON парсит JSON из request body в указанный объект
func parseJSON(r *http.Request, v interface{}) error {
	if r.Body == nil {
		return fmt.Errorf("request body is empty")
	}
	defer r.Body.Close()

	// Читаем body (ограничение 1MB для безопасности)
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("failed to read request body: %w", err)
	}

	if len(body) == 0 {
		return fmt.Errorf("request body is empty")
	}

	if err := json.Unmarshal(body, v); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	return nil
}

// respondJSON отправляет JSON-ответ (обёртка над respondWithJSON для единообразия)
func respondJSON(w http.ResponseWriter, code int, payload interface{}) {
	respondWithJSON(w, code, payload)
}

// ErrorResponse структура для error responses (ИСПРАВЛЕНО: добавлено поле Code для соответствия OpenAPI)
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// respondError отправляет error response в JSON формате (ИСПРАВЛЕНО: добавлен параметр code)
func respondError(w http.ResponseWriter, httpCode int, errorCode string, message string) {
	respondWithJSON(w, httpCode, ErrorResponse{
		Error:   http.StatusText(httpCode),
		Message: message,
		Code:    errorCode,
	})
}
