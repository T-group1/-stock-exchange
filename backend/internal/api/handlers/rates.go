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
	Date             string  `json:"date"`
	Rate             float64 `json:"rate"`
	Source           string  `json:"source,omitempty"`
	ChangePercentage float64 `json:"change_percentage,omitempty"`
}

// HistoryResponse представляет ответ для истории курса пары
type HistoryResponse struct {
	Pair   string         `json:"pair"`
	Period string         `json:"period"`
	Data   []HistoryPoint `json:"data"`
}

// fillHistoryGaps заполняет пропуски в истории курсов (выходные, праздники)
func fillHistoryGaps(history []HistoryPoint, days int) []HistoryPoint {
	if len(history) == 0 {
		return history
	}

	// Создаем карту существующих данных
	historyMap := make(map[string]HistoryPoint)
	for _, point := range history {
		historyMap[point.Date] = point
	}

	// Определяем диапазон дат
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	// Генерируем полный список дат
	var filledHistory []HistoryPoint
	var lastPoint HistoryPoint

	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")

		if point, exists := historyMap[dateStr]; exists {
			filledHistory = append(filledHistory, point)
			lastPoint = point
		} else if lastPoint.Rate > 0 {
			// Заполняем пропуск последним известным курсом
			filledHistory = append(filledHistory, HistoryPoint{
				Date:             dateStr,
				Rate:             lastPoint.Rate,
				Source:           lastPoint.Source,
				ChangePercentage: 0.0, // При заполнении пропуска изменение = 0
			})
		}
	}

	return filledHistory
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

	// Пропускаем проверку quoteExists, если запрашиваем рубли
	if !baseExists || (!quoteExists && quote != "RUB") {
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

	// ✅ ИСПРАВЛЕНО: Обработка параметра period вместо days
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "1m"
	}

	var days int
	switch period {
	case "1d":
		days = 1
	case "1w":
		days = 7
	case "1m":
		days = 30
	case "3m":
		days = 90
	default:
		days = 30
	}

	// ✅ ИСПРАВЛЕНО: Обработка параметра fill_gaps
	fillGaps := r.URL.Query().Get("fill_gaps") == "true"

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
				Date:             d.Format("2006-01-02"),
				Rate:             1.0,
				Source:           "INTERNAL",
				ChangePercentage: 0.0,
			})
		}

		if fillGaps {
			history = fillHistoryGaps(history, days)
		}

		respondWithJSON(w, http.StatusOK, HistoryResponse{
			Pair:   pair,
			Period: period,
			Data:   history,
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

			changePercent := 0.0
			if cp, err := row.ChangePercentage.Float64Value(); err == nil && cp.Valid {
				changePercent = cp.Float64
			}

			history = append(history, HistoryPoint{
				Date:             row.RateDate.Time.Format("2006-01-02"),
				Rate:             1.0 / quoteRatePerUnit,
				Source:           row.Source,
				ChangePercentage: changePercent,
			})
		}

		if fillGaps {
			history = fillHistoryGaps(history, days)
		}

		respondWithJSON(w, http.StatusOK, HistoryResponse{
			Pair:   pair,
			Period: period,
			Data:   history,
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

	if quote == "RUB" {
		quoteNominal = 1.0 // RUB не имеет записи в БД, ставим номинал 1
	}

	// ИСПРАВЛЕНО: проверки на ноль
	if baseNominal == 0 || quoteNominal == 0 {
		http.Error(w, "Invalid currency nominal", http.StatusInternalServerError)
		return
	}

	// Получаем source из первой записи baseHistory
	baseSource := ""
	if len(baseHistory) > 0 {
		baseSource = baseHistory[0].Source
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

		changePercent := 0.0
		if cp, err := row.ChangePercentage.Float64Value(); err == nil && cp.Valid {
			changePercent = cp.Float64
		}

		history = append(history, HistoryPoint{
			Date:             dateStr,
			Rate:             baseRatePerRub / quoteRatePerRub,
			Source:           baseSource,
			ChangePercentage: changePercent,
		})
	}

	if fillGaps {
		history = fillHistoryGaps(history, days)
	}

	respondWithJSON(w, http.StatusOK, HistoryResponse{
		Pair:   pair,
		Period: period,
		Data:   history,
	})
}
