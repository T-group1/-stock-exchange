package handlers

import (
	"encoding/json"
	"net/http"

	"T_Project/internal/api/response"
	"T_Project/internal/db"
)

// ConvertHandler обработчик эндпоинта конвертации
type ConvertHandler struct {
	queries db.Querier
}

// NewConvertHandler создаёт новый обработчик конвертации
func NewConvertHandler(queries db.Querier) *ConvertHandler {
	return &ConvertHandler{queries: queries}
}

// ConversionRequest DTO запроса
type ConversionRequest struct {
	From   string  `json:"from"`
	To     string  `json:"to"`
	Amount float64 `json:"amount"`
	Date   string  `json:"date,omitempty"` // опционально, формат YYYY-MM-DD
}

// ConversionResponse DTO ответа
type ConversionResponse struct {
	From   string  `json:"from"`
	To     string  `json:"to"`
	Amount float64 `json:"amount"`
	Result float64 `json:"result"`
	Rate   float64 `json:"rate"`
	Date   string  `json:"date"`
}

// Convert выполняет конвертацию валюты
func (h *ConvertHandler) Convert(w http.ResponseWriter, r *http.Request) {
	var req ConversionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if req.From == "" || req.To == "" || req.Amount <= 0 {
		response.BadRequest(w, "Fields 'from', 'to' and 'amount' are required")
		return
	}

	// Получаем курсы
	rates, err := h.queries.GetLatestRates(r.Context())
	if err != nil {
		response.InternalError(w, "Failed to fetch rates")
		return
	}

	// Создаём map курсов
	rateMap := make(map[string]float64)
	var dateStr string
	for _, rate := range rates {
		rateFloat, err := rate.Rate.Float64Value()
		if err != nil || !rateFloat.Valid {
			continue
		}
		rateMap[rate.CurrencyCode] = rateFloat.Float64
		if dateStr == "" {
			dateStr = rate.RateDate.Time.Format("2006-01-02")
		}
	}

	// Вычисляем курс конвертации
	var conversionRate float64

	if req.From == "RUB" {
		// Из RUB в другую валюту
		toRate, ok := rateMap[req.To]
		if !ok || toRate == 0 {
			response.NotFound(w, "Rate not found for currency: "+req.To)
			return
		}
		conversionRate = 1.0 / toRate
	} else if req.To == "RUB" {
		// Из другой валюты в RUB
		fromRate, ok := rateMap[req.From]
		if !ok {
			response.NotFound(w, "Rate not found for currency: "+req.From)
			return
		}
		conversionRate = fromRate
	} else {
		// Кросс-курс через RUB
		fromRate, ok1 := rateMap[req.From]
		toRate, ok2 := rateMap[req.To]
		if !ok1 || !ok2 || toRate == 0 {
			response.NotFound(w, "Rates not found for pair: "+req.From+"_"+req.To)
			return
		}
		conversionRate = fromRate / toRate
	}

	result := req.Amount * conversionRate

	response.WriteSuccess(w, ConversionResponse{
		From:   req.From,
		To:     req.To,
		Amount: req.Amount,
		Result: result,
		Rate:   conversionRate,
		Date:   dateStr,
	})
}
