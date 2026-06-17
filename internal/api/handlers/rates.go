package handlers

import (
	"fmt"
	"net/http"
	"time"

	"T_Project/internal/db"

	"github.com/jackc/pgx/v5/pgtype"
)

type RatesHandler struct {
	queries db.Querier
}

func NewRatesHandler(queries db.Querier) *RatesHandler {
	return &RatesHandler{queries: queries}
}

// RateResponse представляет ответ для текущего курса пары
type RateResponse struct {
	Pair   string  `json:"pair"`
	Rate   float64 `json:"rate"`
	Date   string  `json:"date"`
	Base   string  `json:"base"`
	Quote  string  `json:"quote"`
	Source string  `json:"source"`
}

// HistoryPoint представляет точку данных для графика
type HistoryPoint struct {
	Date string  `json:"date"`
	Rate float64 `json:"rate"`
}

// HistoryResponse представляет ответ для истории курса пары
type HistoryResponse struct {
	Pair    string         `json:"pair"`
	Base    string         `json:"base"`
	Quote   string         `json:"quote"`
	History []HistoryPoint `json:"history"`
}

// GetAll возвращает все последние курсы валют (ОПТИМИЗИРОВАНО: 1 запрос в БД)
func (h *RatesHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	rateMap, latestDate, err := getLatestRateMap(r.Context(), h.queries)
	if err != nil {
		http.Error(w, "Failed to fetch rates", http.StatusInternalServerError)
		return
	}

	response := make(map[string]interface{})
	for code, info := range rateMap {
		response[code] = map[string]interface{}{
			"rate":              info.Rate,
			"date":              info.Date,
			"source":            info.Source,
			"change_percentage": info.ChangePercentage,
		}
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"date":  latestDate,
		"rates": response,
	})
}

// GetByPair возвращает курс для конкретной пары валют
func (h *RatesHandler) GetByPair(w http.ResponseWriter, r *http.Request) {
	pair := extractPairFromURL(r)
	base, quote, err := parsePair(pair)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Edge case: одинаковые валюты — ИСПРАВЛЕНО: не делаем лишний запрос в БД
	if base == quote {
		respondWithJSON(w, http.StatusOK, RateResponse{
			Pair:   pair,
			Rate:   1.0,
			Date:   time.Now().Format("2006-01-02"), // Используем текущую дату
			Base:   base,
			Quote:  quote,
			Source: "INTERNAL",
		})
		return
	}

	rateMap, latestDate, err := getLatestRateMap(r.Context(), h.queries)
	if err != nil {
		http.Error(w, "Failed to fetch rates", http.StatusInternalServerError)
		return
	}

	// Обработка случая, когда базовая валюта - RUB
	if base == "RUB" {
		quoteCurr, err := h.queries.GetCurrencyByCode(r.Context(), quote)
		if err != nil {
			http.Error(w, fmt.Sprintf("Currency not found: %s", quote), http.StatusNotFound)
			return
		}

		quoteInfo, exists := rateMap[quote]
		if !exists {
			http.Error(w, fmt.Sprintf("Rate not found for currency: %s", quote), http.StatusNotFound)
			return
		}

		// ИСПРАВЛЕНО: проверка на ноль перед делением
		quoteRatePerUnit := quoteInfo.Rate / float64(quoteCurr.Nominal)
		if quoteRatePerUnit == 0 {
			http.Error(w, "Invalid rate data (division by zero)", http.StatusInternalServerError)
			return
		}

		resultRate := 1.0 / quoteRatePerUnit

		respondWithJSON(w, http.StatusOK, RateResponse{
			Pair:   pair,
			Rate:   resultRate,
			Date:   latestDate,
			Base:   base,
			Quote:  quote,
			Source: quoteInfo.Source,
		})
		return
	}

	// Стандартная логика (включая случай, когда quote == "RUB")
	baseInfo, baseExists := rateMap[base]
	quoteInfo, quoteExists := rateMap[quote]

	if !baseExists || !quoteExists {
		http.Error(w, fmt.Sprintf("Rates not found for pair: %s", pair), http.StatusNotFound)
		return
	}

	var resultRate float64
	var source string

	if quote == "RUB" {
		baseCurr, err := h.queries.GetCurrencyByCode(r.Context(), base)
		if err == nil {
			// ИСПРАВЛЕНО: проверка на ноль перед делением
			baseNominal := float64(baseCurr.Nominal)
			if baseNominal == 0 {
				http.Error(w, "Invalid currency nominal (division by zero)", http.StatusInternalServerError)
				return
			}
			resultRate = baseInfo.Rate / baseNominal
		} else {
			resultRate = baseInfo.Rate // fallback
		}
		source = baseInfo.Source
	} else {
		// Кросс-курс через рубли
		baseCurr, err1 := h.queries.GetCurrencyByCode(r.Context(), base)
		quoteCurr, err2 := h.queries.GetCurrencyByCode(r.Context(), quote)

		if err1 != nil || err2 != nil {
			http.Error(w, "Failed to fetch currency metadata", http.StatusInternalServerError)
			return
		}

		// ИСПРАВЛЕНО: проверки на ноль перед делением
		baseNominal := float64(baseCurr.Nominal)
		quoteNominal := float64(quoteCurr.Nominal)

		if baseNominal == 0 || quoteNominal == 0 {
			http.Error(w, "Invalid currency nominal (division by zero)", http.StatusInternalServerError)
			return
		}

		baseRatePerRub := baseInfo.Rate / baseNominal
		quoteRatePerRub := quoteInfo.Rate / quoteNominal

		if quoteRatePerRub == 0 {
			http.Error(w, "Invalid rate data (division by zero)", http.StatusInternalServerError)
			return
		}

		resultRate = baseRatePerRub / quoteRatePerRub
		source = fmt.Sprintf("%s/%s", baseInfo.Source, quoteInfo.Source)
	}

	respondWithJSON(w, http.StatusOK, RateResponse{
		Pair:   pair,
		Rate:   resultRate,
		Date:   latestDate,
		Base:   base,
		Quote:  quote,
		Source: source,
	})
}

// GetHistory возвращает историю курса для конкретной пары валют
func (h *RatesHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	pair := extractPairFromURL(r)
	base, quote, err := parsePair(pair)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	daysStr := r.URL.Query().Get("days")
	if daysStr == "" {
		daysStr = "30"
	}

	var days int
	if _, err := fmt.Sscanf(daysStr, "%d", &days); err != nil || days <= 0 || days > 365 {
		days = 30
	}

	startDate := time.Now().AddDate(0, 0, -days)
	pgDate := pgtype.Date{
		Time:  startDate,
		Valid: true,
	}

	// Edge case: одинаковые валюты
	if base == quote {
		var history []HistoryPoint
		for i := 0; i <= days; i++ {
			d := time.Now().AddDate(0, 0, -i)
			history = append(history, HistoryPoint{
				Date: d.Format("2006-01-02"),
				Rate: 1.0,
			})
		}
		respondWithJSON(w, http.StatusOK, HistoryResponse{
			Pair:    pair,
			Base:    base,
			Quote:   quote,
			History: history,
		})
		return
	}

	if base == "RUB" {
		quoteCurr, err := h.queries.GetCurrencyByCode(r.Context(), quote)
		if err != nil {
			http.Error(w, fmt.Sprintf("Currency not found: %s", quote), http.StatusNotFound)
			return
		}

		historyRows, err := h.queries.GetRateHistory(r.Context(), db.GetRateHistoryParams{
			CurrencyCode: quote,
			RateDate:     pgDate,
		})
		if err != nil {
			http.Error(w, "Failed to fetch rate history", http.StatusInternalServerError)
			return
		}

		var history []HistoryPoint
		quoteNominal := float64(quoteCurr.Nominal)

		// ИСПРАВЛЕНО: проверка на ноль
		if quoteNominal == 0 {
			http.Error(w, "Invalid currency nominal", http.StatusInternalServerError)
			return
		}

		for _, row := range historyRows {
			rateFloat, err := row.Rate.Float64Value()
			if err != nil || !rateFloat.Valid {
				continue
			}

			quoteRatePerUnit := rateFloat.Float64 / quoteNominal
			if quoteRatePerUnit == 0 {
				continue
			}

			history = append(history, HistoryPoint{
				Date: row.RateDate.Time.Format("2006-01-02"),
				Rate: 1.0 / quoteRatePerUnit,
			})
		}

		respondWithJSON(w, http.StatusOK, HistoryResponse{
			Pair:    pair,
			Base:    base,
			Quote:   quote,
			History: history,
		})
		return
	}

	// Стандартная логика истории
	baseHistory, err := h.queries.GetRateHistory(r.Context(), db.GetRateHistoryParams{
		CurrencyCode: base,
		RateDate:     pgDate,
	})
	if err != nil {
		http.Error(w, "Failed to fetch base currency history", http.StatusInternalServerError)
		return
	}

	quoteHistory, err := h.queries.GetRateHistory(r.Context(), db.GetRateHistoryParams{
		CurrencyCode: quote,
		RateDate:     pgDate,
	})
	if err != nil {
		http.Error(w, "Failed to fetch quote currency history", http.StatusInternalServerError)
		return
	}

	baseCurr, _ := h.queries.GetCurrencyByCode(r.Context(), base)
	quoteCurr, _ := h.queries.GetCurrencyByCode(r.Context(), quote)
	baseNominal := float64(baseCurr.Nominal)
	quoteNominal := float64(quoteCurr.Nominal)

	// ИСПРАВЛЕНО: проверки на ноль
	if baseNominal == 0 || quoteNominal == 0 {
		http.Error(w, "Invalid currency nominal", http.StatusInternalServerError)
		return
	}

	quoteMap := make(map[string]float64)
	for _, row := range quoteHistory {
		rateFloat, err := row.Rate.Float64Value()
		if err == nil && rateFloat.Valid {
			quoteMap[row.RateDate.Time.Format("2006-01-02")] = rateFloat.Float64 / quoteNominal
		}
	}

	var history []HistoryPoint
	for _, row := range baseHistory {
		rateFloat, err := row.Rate.Float64Value()
		if err != nil || !rateFloat.Valid {
			continue
		}

		dateStr := row.RateDate.Time.Format("2006-01-02")
		baseRatePerRub := rateFloat.Float64 / baseNominal

		quoteRatePerRub := 1.0
		if quote != "RUB" {
			if qRate, exists := quoteMap[dateStr]; exists {
				quoteRatePerRub = qRate
			} else {
				continue
			}
		}

		if quoteRatePerRub == 0 {
			continue
		}

		history = append(history, HistoryPoint{
			Date: dateStr,
			Rate: baseRatePerRub / quoteRatePerRub,
		})
	}

	respondWithJSON(w, http.StatusOK, HistoryResponse{
		Pair:    pair,
		Base:    base,
		Quote:   quote,
		History: history,
	})
}
