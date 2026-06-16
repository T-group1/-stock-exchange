package handlers

import (
	"encoding/json"
	"net/http"
	"time"

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

	// ИСПРАВЛЕНО: обработка RUB → RUB
	if req.From == req.To {
		dateStr := time.Now().Format("2006-01-02")
		if req.Date != "" {
			dateStr = req.Date
		}
		response.WriteSuccess(w, ConversionResponse{
			From:   req.From,
			To:     req.To,
			Amount: req.Amount,
			Result: req.Amount,
			Rate:   1.0,
			Date:   dateStr,
		})
		return
	}

	// ИСПРАВЛЕНО: поддержка исторической конвертации через Date
	var rateMap map[string]rateInfo
	var dateStr string
	var err error

	if req.Date != "" {
		parsedDate, parseErr := time.Parse("2006-01-02", req.Date)
		if parseErr != nil {
			response.BadRequest(w, "Invalid date format. Use YYYY-MM-DD")
			return
		}
		rateMap, err = getRateMapByDate(r.Context(), h.queries, parsedDate)
		if err != nil {
			response.InternalError(w, "Failed to fetch historical rates: "+err.Error())
			return
		}
		dateStr = req.Date
	} else {
		rateMap, dateStr, err = getLatestRateMap(r.Context(), h.queries)
		if err != nil {
			response.InternalError(w, "Failed to fetch rates")
			return
		}
	}

	// Вычисляем курс конвертации
	var conversionRate float64

	if req.From == "RUB" {
		toInfo, ok := rateMap[req.To]
		if !ok || toInfo.Rate == 0 {
			response.NotFound(w, "Rate not found for currency: "+req.To)
			return
		}
		conversionRate = 1.0 / toInfo.Rate
	} else if req.To == "RUB" {
		fromInfo, ok := rateMap[req.From]
		if !ok {
			response.NotFound(w, "Rate not found for currency: "+req.From)
			return
		}
		conversionRate = fromInfo.Rate
	} else {
		fromInfo, ok1 := rateMap[req.From]
		toInfo, ok2 := rateMap[req.To]
		if !ok1 || !ok2 || toInfo.Rate == 0 {
			response.NotFound(w, "Rates not found for pair: "+req.From+"_"+req.To)
			return
		}
		conversionRate = fromInfo.Rate / toInfo.Rate
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
