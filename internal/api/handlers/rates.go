package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"T_Project/internal/api/response"
	"T_Project/internal/db"

	"github.com/jackc/pgx/v5/pgtype"
)

// RatesHandler обработчик эндпоинтов курсов
type RatesHandler struct {
	queries db.Querier
}

// NewRatesHandler создаёт новый обработчик курсов
func NewRatesHandler(queries db.Querier) *RatesHandler {
	return &RatesHandler{queries: queries}
}

// RateResponse DTO для курса
type RateResponse struct {
	Pair             string  `json:"pair"`
	Rate             float64 `json:"rate"`
	Date             string  `json:"date"`
	Source           string  `json:"source"`
	ChangePercentage float64 `json:"change_percentage,omitempty"`
}

// RatesResponse DTO для списка курсов
type RatesResponse struct {
	Rates []RateResponse `json:"rates"`
	Date  string         `json:"date"`
}

// GetAll возвращает текущие курсы всех валют
func (h *RatesHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	rates, err := h.queries.GetLatestRates(r.Context())
	if err != nil {
		response.InternalError(w, "Failed to fetch rates")
		return
	}

	result := make([]RateResponse, 0, len(rates))
	var latestDate string

	for _, rate := range rates {
		// Конвертируем pgtype.Numeric в float64
		rateFloat, err := rate.Rate.Float64Value()
		if err != nil || !rateFloat.Valid {
			continue
		}

		// Конвертируем pgtype.Date в string
		dateStr := rate.RateDate.Time.Format("2006-01-02")
		if latestDate == "" {
			latestDate = dateStr
		}

		// Формируем пару (валюта_RUB)
		pair := rate.CurrencyCode + "_RUB"

		// Конвертируем change_percentage
		changePct := 0.0
		if rate.ChangePercentage.Valid {
			if cp, err := rate.ChangePercentage.Float64Value(); err == nil && cp.Valid {
				changePct = cp.Float64
			}
		}

		result = append(result, RateResponse{
			Pair:             pair,
			Rate:             rateFloat.Float64,
			Date:             dateStr,
			Source:           rate.Source,
			ChangePercentage: changePct,
		})
	}

	response.WriteSuccess(w, RatesResponse{
		Rates: result,
		Date:  latestDate,
	})
}

// GetByPair возвращает курс для конкретной пары
func (h *RatesHandler) GetByPair(w http.ResponseWriter, r *http.Request) {
	// Извлекаем pair из path (например, /rates/USD_RUB)
	pair := extractPairFromPath(r.URL.Path)
	if pair == "" {
		response.BadRequest(w, "Currency pair is required")
		return
	}

	// Парсим пару (например, USD_RUB -> USD, RUB)
	base, quote, err := parsePair(pair)
	if err != nil {
		response.BadRequest(w, "Invalid currency pair format. Use format: USD_RUB")
		return
	}

	// Если quote == RUB, просто берём курс из БД
	if quote == "RUB" {
		h.getDirectRate(w, r, base)
		return
	}

	// Иначе вычисляем кросс-курс через RUB
	h.getCrossRate(w, r, base, quote)
}

// getDirectRate получает прямой курс к RUB
func (h *RatesHandler) getDirectRate(w http.ResponseWriter, r *http.Request, currencyCode string) {
	rates, err := h.queries.GetLatestRates(r.Context())
	if err != nil {
		response.InternalError(w, "Failed to fetch rates")
		return
	}

	for _, rate := range rates {
		if rate.CurrencyCode == currencyCode {
			rateFloat, err := rate.Rate.Float64Value()
			if err != nil || !rateFloat.Valid {
				continue
			}

			changePct := 0.0
			if rate.ChangePercentage.Valid {
				if cp, err := rate.ChangePercentage.Float64Value(); err == nil && cp.Valid {
					changePct = cp.Float64
				}
			}

			response.WriteSuccess(w, RateResponse{
				Pair:             currencyCode + "_RUB",
				Rate:             rateFloat.Float64,
				Date:             rate.RateDate.Time.Format("2006-01-02"),
				Source:           rate.Source,
				ChangePercentage: changePct,
			})
			return
		}
	}

	response.NotFound(w, "Rate not found for currency: "+currencyCode)
}

// getCrossRate вычисляет кросс-курс через RUB
func (h *RatesHandler) getCrossRate(w http.ResponseWriter, r *http.Request, base, quote string) {
	rates, err := h.queries.GetLatestRates(r.Context())
	if err != nil {
		response.InternalError(w, "Failed to fetch rates")
		return
	}

	var baseRate, quoteRate float64
	var dateStr string

	for _, rate := range rates {
		rateFloat, err := rate.Rate.Float64Value()
		if err != nil || !rateFloat.Valid {
			continue
		}

		if rate.CurrencyCode == base {
			baseRate = rateFloat.Float64
			dateStr = rate.RateDate.Time.Format("2006-01-02")
		}
		if rate.CurrencyCode == quote {
			quoteRate = rateFloat.Float64
		}
	}

	if baseRate == 0 || quoteRate == 0 {
		response.NotFound(w, "Rates not found for pair: "+base+"_"+quote)
		return
	}

	// Кросс-курс: base/quote = (base/RUB) / (quote/RUB)
	crossRate := baseRate / quoteRate

	response.WriteSuccess(w, RateResponse{
		Pair:   base + "_" + quote,
		Rate:   crossRate,
		Date:   dateStr,
		Source: "calculated",
	})
}

// GetHistory возвращает историю курсов
func (h *RatesHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	// Извлекаем pair из path
	pair := extractPairFromPath(r.URL.Path)
	if pair == "" {
		response.BadRequest(w, "Currency pair is required")
		return
	}

	base, quote, err := parsePair(pair)
	if err != nil {
		response.BadRequest(w, "Invalid currency pair format")
		return
	}

	// Получаем параметры запроса
	daysStr := r.URL.Query().Get("days")
	days := 30 // по умолчанию 30 дней
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}

	fromDate := time.Now().AddDate(0, 0, -days)

	// Если quote == RUB, берём историю напрямую
	if quote == "RUB" {
		h.getDirectHistory(w, r, base, fromDate)
		return
	}

	// Иначе вычисляем кросс-курс историю
	h.getCrossHistory(w, r, base, quote, fromDate)
}

// getDirectHistory получает историю прямого курса
func (h *RatesHandler) getDirectHistory(w http.ResponseWriter, r *http.Request, currencyCode string, fromDate time.Time) {
	pgDate := pgtype.Date{}
	if err := pgDate.Scan(fromDate); err != nil {
		response.InternalError(w, "Invalid date format")
		return
	}

	history, err := h.queries.GetRateHistory(r.Context(), db.GetRateHistoryParams{
		CurrencyCode: currencyCode,
		RateDate:     pgDate,
	})
	if err != nil {
		response.InternalError(w, "Failed to fetch rate history")
		return
	}

	result := make([]RateResponse, 0, len(history))
	for _, h := range history {
		rateFloat, err := h.Rate.Float64Value()
		if err != nil || !rateFloat.Valid {
			continue
		}

		result = append(result, RateResponse{
			Pair:   currencyCode + "_RUB",
			Rate:   rateFloat.Float64,
			Date:   h.RateDate.Time.Format("2006-01-02"),
			Source: "cb_rf",
		})
	}

	response.WriteSuccess(w, result)
}

// getCrossHistory вычисляет историю кросс-курса
func (h *RatesHandler) getCrossHistory(w http.ResponseWriter, r *http.Request, base, quote string, fromDate time.Time) {
	pgDate := pgtype.Date{}
	if err := pgDate.Scan(fromDate); err != nil {
		response.InternalError(w, "Invalid date format")
		return
	}

	// Получаем историю для base и quote
	baseHistory, err := h.queries.GetRateHistory(r.Context(), db.GetRateHistoryParams{
		CurrencyCode: base,
		RateDate:     pgDate,
	})
	if err != nil {
		response.InternalError(w, "Failed to fetch base rate history")
		return
	}

	quoteHistory, err := h.queries.GetRateHistory(r.Context(), db.GetRateHistoryParams{
		CurrencyCode: quote,
		RateDate:     pgDate,
	})
	if err != nil {
		response.InternalError(w, "Failed to fetch quote rate history")
		return
	}

	// Создаём map для quote history
	quoteMap := make(map[string]float64)
	for _, q := range quoteHistory {
		rateFloat, err := q.Rate.Float64Value()
		if err != nil || !rateFloat.Valid {
			continue
		}
		quoteMap[q.RateDate.Time.Format("2006-01-02")] = rateFloat.Float64
	}

	// Вычисляем кросс-курс для каждой даты
	result := make([]RateResponse, 0, len(baseHistory))
	for _, b := range baseHistory {
		baseFloat, err := b.Rate.Float64Value()
		if err != nil || !baseFloat.Valid {
			continue
		}

		dateStr := b.RateDate.Time.Format("2006-01-02")
		quoteRate, ok := quoteMap[dateStr]
		if !ok || quoteRate == 0 {
			continue
		}

		crossRate := baseFloat.Float64 / quoteRate
		result = append(result, RateResponse{
			Pair:   base + "_" + quote,
			Rate:   crossRate,
			Date:   dateStr,
			Source: "calculated",
		})
	}

	response.WriteSuccess(w, result)
}

// extractPairFromPath извлекает пару из URL path
func extractPairFromPath(path string) string {
	// Ожидаем формат: /rates/USD_RUB или /rates/USD_RUB/history
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == "rates" && i+1 < len(parts) {
			pair := parts[i+1]
			// Убираем /history если есть
			if pair == "history" {
				return ""
			}
			return pair
		}
	}
	return ""
}

// parsePair парсит строку пары валют
func parsePair(pair string) (string, string, error) {
	parts := strings.Split(pair, "_")
	if len(parts) != 2 {
		return "", "", http.ErrBodyNotAllowed
	}
	return strings.ToUpper(parts[0]), strings.ToUpper(parts[1]), nil
}
