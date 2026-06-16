package handlers

import (
	"net/http"
	"strconv"
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
	rateMap, latestDate, err := getLatestRateMap(r.Context(), h.queries)
	if err != nil {
		response.InternalError(w, "Failed to fetch rates")
		return
	}

	result := make([]RateResponse, 0, len(rateMap))
	for code, info := range rateMap {
		pair := code + "_RUB"

		// Получаем change_percentage из БД
		changePct := h.getChangePercentage(r, code)

		result = append(result, RateResponse{
			Pair:             pair,
			Rate:             info.Rate,
			Date:             info.Date,
			Source:           info.Source,
			ChangePercentage: changePct,
		})
	}

	response.WriteSuccess(w, RatesResponse{
		Rates: result,
		Date:  latestDate,
	})
}

// getChangePercentage получает процент изменения для валюты
func (h *RatesHandler) getChangePercentage(r *http.Request, currencyCode string) float64 {
	rates, err := h.queries.GetLatestRates(r.Context())
	if err != nil {
		return 0
	}
	for _, rate := range rates {
		if rate.CurrencyCode == currencyCode && rate.ChangePercentage.Valid {
			if cp, err := rate.ChangePercentage.Float64Value(); err == nil && cp.Valid {
				return cp.Float64
			}
		}
	}
	return 0
}

// GetByPair возвращает курс для конкретной пары
// ИСПРАВЛЕНО: используем chi.URLParam и общую функцию parsePair
func (h *RatesHandler) GetByPair(w http.ResponseWriter, r *http.Request) {
	pair := extractPairFromURL(r)
	if pair == "" {
		response.BadRequest(w, "Currency pair is required")
		return
	}

	base, quote, err := parsePair(pair)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	// ИСПРАВЛЕНО: обработка RUB → RUB
	if base == quote {
		rateMap, dateStr, err := getLatestRateMap(r.Context(), h.queries)
		if err != nil {
			response.InternalError(w, "Failed to fetch rates")
			return
		}
		_ = rateMap
		response.WriteSuccess(w, RateResponse{
			Pair:   base + "_" + quote,
			Rate:   1.0,
			Date:   dateStr,
			Source: "identity",
		})
		return
	}

	if quote == "RUB" {
		h.getDirectRate(w, r, base)
		return
	}

	h.getCrossRate(w, r, base, quote)
}

// getDirectRate получает прямой курс к RUB
func (h *RatesHandler) getDirectRate(w http.ResponseWriter, r *http.Request, currencyCode string) {
	rateMap, _, err := getLatestRateMap(r.Context(), h.queries)
	if err != nil {
		response.InternalError(w, "Failed to fetch rates")
		return
	}

	info, ok := rateMap[currencyCode]
	if !ok {
		response.NotFound(w, "Rate not found for currency: "+currencyCode)
		return
	}

	changePct := h.getChangePercentage(r, currencyCode)

	response.WriteSuccess(w, RateResponse{
		Pair:             currencyCode + "_RUB",
		Rate:             info.Rate,
		Date:             info.Date,
		Source:           info.Source,
		ChangePercentage: changePct,
	})
}

// getCrossRate вычисляет кросс-курс через RUB
func (h *RatesHandler) getCrossRate(w http.ResponseWriter, r *http.Request, base, quote string) {
	rateMap, _, err := getLatestRateMap(r.Context(), h.queries)
	if err != nil {
		response.InternalError(w, "Failed to fetch rates")
		return
	}

	baseInfo, ok1 := rateMap[base]
	quoteInfo, ok2 := rateMap[quote]

	if !ok1 || !ok2 || quoteInfo.Rate == 0 {
		response.NotFound(w, "Rates not found for pair: "+base+"_"+quote)
		return
	}

	crossRate := baseInfo.Rate / quoteInfo.Rate

	response.WriteSuccess(w, RateResponse{
		Pair:   base + "_" + quote,
		Rate:   crossRate,
		Date:   baseInfo.Date,
		Source: "calculated",
	})
}

// GetHistory возвращает историю курсов
// ИСПРАВЛЕНО: используем chi.URLParam
func (h *RatesHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	pair := extractPairFromURL(r)
	if pair == "" {
		response.BadRequest(w, "Currency pair is required")
		return
	}

	base, quote, err := parsePair(pair)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	daysStr := r.URL.Query().Get("days")
	days := 30
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}

	fromDate := time.Now().AddDate(0, 0, -days)

	if quote == "RUB" {
		h.getDirectHistory(w, r, base, fromDate)
		return
	}

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

	quoteMap := make(map[string]float64)
	for _, q := range quoteHistory {
		rateFloat, err := q.Rate.Float64Value()
		if err != nil || !rateFloat.Valid {
			continue
		}
		quoteMap[q.RateDate.Time.Format("2006-01-02")] = rateFloat.Float64
	}

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
