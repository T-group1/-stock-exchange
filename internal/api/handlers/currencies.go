package handlers

import (
	"net/http"

	"T_Project/internal/api/response"
	"T_Project/internal/db"
)

// CurrenciesHandler обработчик эндпоинтов валют
type CurrenciesHandler struct {
	queries db.Querier
}

// NewCurrenciesHandler создаёт новый обработчик валют
func NewCurrenciesHandler(queries db.Querier) *CurrenciesHandler {
	return &CurrenciesHandler{queries: queries}
}

// CurrencyResponse DTO для ответа
type CurrencyResponse struct {
	Code    string `json:"code"`
	Name    string `json:"name"`
	Nominal int32  `json:"nominal"`
}

// List возвращает список всех валют
func (h *CurrenciesHandler) List(w http.ResponseWriter, r *http.Request) {
	currencies, err := h.queries.GetCurrencies(r.Context())
	if err != nil {
		response.InternalError(w, "Failed to fetch currencies")
		return
	}

	result := make([]CurrencyResponse, len(currencies))
	for i, c := range currencies {
		result[i] = CurrencyResponse{
			Code:    c.Code,
			Name:    c.Name,
			Nominal: c.Nominal,
		}
	}

	response.WriteSuccess(w, result)
}

// GetByCode возвращает валюту по коду
func (h *CurrenciesHandler) GetByCode(w http.ResponseWriter, r *http.Request) {
	// Код валюты передаётся через query параметр или path variable
	code := r.URL.Query().Get("code")
	if code == "" {
		response.BadRequest(w, "Currency code is required")
		return
	}

	currency, err := h.queries.GetCurrencyByCode(r.Context(), code)
	if err != nil {
		response.NotFound(w, "Currency not found")
		return
	}

	response.WriteSuccess(w, CurrencyResponse{
		Code:    currency.Code,
		Name:    currency.Name,
		Nominal: currency.Nominal,
	})
}
