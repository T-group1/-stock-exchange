package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"T_Project/internal/db"
)

type ConvertHandler struct {
	queries db.Querier
}

func NewConvertHandler(queries db.Querier) *ConvertHandler {
	return &ConvertHandler{queries: queries}
}

// ConvertRequest представляет запрос на конвертацию
type ConvertRequest struct {
	From   string  `json:"from"`
	To     string  `json:"to"`
	Amount float64 `json:"amount"`
	Date   string  `json:"date,omitempty"` // опционально, формат YYYY-MM-DD
}

// ConvertResponse представляет ответ конвертации
type ConvertResponse struct {
	From   string  `json:"from"`
	To     string  `json:"to"`
	Amount float64 `json:"amount"`
	Result float64 `json:"result"`
	Rate   float64 `json:"rate"`
	Date   string  `json:"date"`
	Source string  `json:"source"`
}

// Convert конвертирует одну валюту в другую
func (h *ConvertHandler) Convert(w http.ResponseWriter, r *http.Request) {
	var req ConvertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.From == "" || req.To == "" || req.Amount <= 0 {
		http.Error(w, "Invalid parameters: from, to, and amount > 0 are required", http.StatusBadRequest)
		return
	}

	// Edge case: одинаковые валюты
	if req.From == req.To {
		respondWithJSON(w, http.StatusOK, ConvertResponse{
			From:   req.From,
			To:     req.To,
			Amount: req.Amount,
			Result: req.Amount,
			Rate:   1.0,
			Date:   time.Now().Format("2006-01-02"),
			Source: "INTERNAL",
		})
		return
	}

	var rateMap map[string]rateInfo
	var dateStr string
	var err error

	// Если указана дата, получаем курсы на эту дату
	if req.Date != "" {
		parsedDate, err := time.Parse("2006-01-02", req.Date)
		if err != nil {
			http.Error(w, "Invalid date format, expected YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		rateMap, err = getRateMapByDate(r.Context(), h.queries, parsedDate)
		if err != nil {
			http.Error(w, "Failed to fetch rates for date", http.StatusInternalServerError)
			return
		}
		dateStr = parsedDate.Format("2006-01-02")
	} else {
		// Иначе получаем последние курсы
		rateMap, dateStr, err = getLatestRateMap(r.Context(), h.queries)
		if err != nil {
			http.Error(w, "Failed to fetch latest rates", http.StatusInternalServerError)
			return
		}
	}

	// Специальный случай: конвертация из RUB
	if req.From == "RUB" {
		toCurr, err := h.queries.GetCurrencyByCode(r.Context(), req.To)
		if err != nil {
			http.Error(w, fmt.Sprintf("Currency not found: %s", req.To), http.StatusNotFound)
			return
		}

		toInfo, exists := rateMap[req.To]
		if !exists {
			http.Error(w, fmt.Sprintf("Rate not found for currency: %s", req.To), http.StatusNotFound)
			return
		}

		// ИСПРАВЛЕНО: проверка на ноль
		toNominal := float64(toCurr.Nominal)
		if toNominal == 0 {
			http.Error(w, "Invalid currency nominal (division by zero)", http.StatusInternalServerError)
			return
		}

		toRatePerUnit := toInfo.Rate / toNominal
		if toRatePerUnit == 0 {
			http.Error(w, "Invalid rate data (division by zero)", http.StatusInternalServerError)
			return
		}

		rate := 1.0 / toRatePerUnit
		result := req.Amount * rate

		respondWithJSON(w, http.StatusOK, ConvertResponse{
			From:   req.From,
			To:     req.To,
			Amount: req.Amount,
			Result: result,
			Rate:   rate,
			Date:   dateStr,
			Source: toInfo.Source,
		})
		return
	}

	// Специальный случай: конвертация в RUB
	if req.To == "RUB" {
		fromCurr, err := h.queries.GetCurrencyByCode(r.Context(), req.From)
		if err != nil {
			http.Error(w, fmt.Sprintf("Currency not found: %s", req.From), http.StatusNotFound)
			return
		}

		fromInfo, exists := rateMap[req.From]
		if !exists {
			http.Error(w, fmt.Sprintf("Rate not found for currency: %s", req.From), http.StatusNotFound)
			return
		}

		// ИСПРАВЛЕНО: проверка на ноль
		fromNominal := float64(fromCurr.Nominal)
		if fromNominal == 0 {
			http.Error(w, "Invalid currency nominal (division by zero)", http.StatusInternalServerError)
			return
		}

		rate := fromInfo.Rate / fromNominal
		result := req.Amount * rate

		respondWithJSON(w, http.StatusOK, ConvertResponse{
			From:   req.From,
			To:     req.To,
			Amount: req.Amount,
			Result: result,
			Rate:   rate,
			Date:   dateStr,
			Source: fromInfo.Source,
		})
		return
	}

	// Кросс-курс через рубли
	fromInfo, fromExists := rateMap[req.From]
	toInfo, toExists := rateMap[req.To]

	if !fromExists || !toExists {
		http.Error(w, fmt.Sprintf("Rates not found for pair: %s_%s", req.From, req.To), http.StatusNotFound)
		return
	}

	fromCurr, err1 := h.queries.GetCurrencyByCode(r.Context(), req.From)
	toCurr, err2 := h.queries.GetCurrencyByCode(r.Context(), req.To)

	if err1 != nil || err2 != nil {
		http.Error(w, "Failed to fetch currency metadata", http.StatusInternalServerError)
		return
	}

	// ИСПРАВЛЕНО: все проверки на ноль
	fromNominal := float64(fromCurr.Nominal)
	toNominal := float64(toCurr.Nominal)

	if fromNominal == 0 || toNominal == 0 {
		http.Error(w, "Invalid currency nominal (division by zero)", http.StatusInternalServerError)
		return
	}

	fromRatePerRub := fromInfo.Rate / fromNominal
	toRatePerRub := toInfo.Rate / toNominal

	if toRatePerRub == 0 {
		http.Error(w, "Invalid rate data (division by zero)", http.StatusInternalServerError)
		return
	}

	rate := fromRatePerRub / toRatePerRub
	result := req.Amount * rate

	respondWithJSON(w, http.StatusOK, ConvertResponse{
		From:   req.From,
		To:     req.To,
		Amount: req.Amount,
		Result: result,
		Rate:   rate,
		Date:   dateStr,
		Source: fmt.Sprintf("%s/%s", fromInfo.Source, toInfo.Source),
	})
}
