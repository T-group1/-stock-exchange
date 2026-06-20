package handlers

import (
	"net/http"

	"T_Project/internal/api/middleware"
	"T_Project/internal/api/response"
	"T_Project/internal/db"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// SubscriptionsHandler обработчик эндпоинтов подписок (алертов)
type SubscriptionsHandler struct {
	queries db.Querier
}

// NewSubscriptionsHandler создаёт новый обработчик подписок
func NewSubscriptionsHandler(queries db.Querier) *SubscriptionsHandler {
	return &SubscriptionsHandler{queries: queries}
}

// SubscriptionResponse DTO для ответа
type SubscriptionResponse struct {
	ID           string  `json:"id"`
	CurrencyCode string  `json:"currency_code"`
	RateValue    float64 `json:"rate_value"`
	Condition    string  `json:"condition"`
	IsActive     bool    `json:"is_active"`
	CreatedAt    int64   `json:"created_at"`
	TriggeredAt  *int64  `json:"triggered_at,omitempty"`
}

// CreateSubscriptionRequest запрос на создание подписки
type CreateSubscriptionRequest struct {
	CurrencyCode string  `json:"currency_code"`
	RateValue    float64 `json:"rate_value"`
	Condition    string  `json:"condition"` // "above" или "below"
}

// List возвращает список подписок текущего пользователя
func (h *SubscriptionsHandler) List(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	userID := pgtype.UUID{}
	if err := userID.Scan(userIDStr); err != nil {
		response.BadRequest(w, "Invalid user ID")
		return
	}

	subscriptions, err := h.queries.GetUserSubscriptions(r.Context(), userID)
	if err != nil {
		response.InternalError(w, "Failed to fetch subscriptions")
		return
	}

	result := make([]SubscriptionResponse, 0, len(subscriptions))
	for _, s := range subscriptions {
		rateFloat, _ := s.RateValue.Float64Value()

		resp := SubscriptionResponse{
			ID:           s.ID.String(),
			CurrencyCode: s.CurrencyCode,
			RateValue:    rateFloat.Float64,
			Condition:    s.Condition,
			IsActive:     s.IsActive.Bool,
			CreatedAt:    s.CreatedAt,
		}

		if s.TriggeredAt.Valid {
			resp.TriggeredAt = &s.TriggeredAt.Int64
		}

		result = append(result, resp)
	}

	response.WriteSuccess(w, map[string]interface{}{
		"subscriptions": result,
	})
}

// Create создаёт новую подписку (алерт)
func (h *SubscriptionsHandler) Create(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	userID := pgtype.UUID{}
	if err := userID.Scan(userIDStr); err != nil {
		response.BadRequest(w, "Invalid user ID")
		return
	}

	var req CreateSubscriptionRequest
	if err := parseJSON(r, &req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	// Валидация
	if req.CurrencyCode == "" {
		response.BadRequest(w, "Currency code is required")
		return
	}
	if req.RateValue <= 0 {
		response.BadRequest(w, "Rate value must be greater than 0")
		return
	}
	if req.Condition != "above" && req.Condition != "below" {
		response.BadRequest(w, "Condition must be 'above' or 'below'")
		return
	}

	// Проверяем что валюта существует
	_, err := h.queries.GetCurrencyByCode(r.Context(), req.CurrencyCode)
	if err != nil {
		response.NotFound(w, "Currency not found")
		return
	}

	// Конвертируем float64 в pgtype.Numeric
	rateValue := pgtype.Numeric{}
	if err := rateValue.Scan(req.RateValue); err != nil {
		response.InternalError(w, "Failed to process rate value")
		return
	}

	subscription, err := h.queries.CreateSubscription(r.Context(), db.CreateSubscriptionParams{
		UserID:       userID,
		CurrencyCode: req.CurrencyCode,
		RateValue:    rateValue,
		Condition:    req.Condition,
	})
	if err != nil {
		response.InternalError(w, "Failed to create subscription")
		return
	}

	rateFloat, _ := subscription.RateValue.Float64Value()

	response.WriteCreated(w, SubscriptionResponse{
		ID:           subscription.ID.String(),
		CurrencyCode: subscription.CurrencyCode,
		RateValue:    rateFloat.Float64,
		Condition:    subscription.Condition,
		IsActive:     subscription.IsActive.Bool,
		CreatedAt:    subscription.CreatedAt,
	})
}

// Delete удаляет подписку
func (h *SubscriptionsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	userID := pgtype.UUID{}
	if err := userID.Scan(userIDStr); err != nil {
		response.BadRequest(w, "Invalid user ID")
		return
	}

	subscriptionIDStr := chi.URLParam(r, "id")
	subscriptionID := pgtype.UUID{}
	if err := subscriptionID.Scan(subscriptionIDStr); err != nil {
		response.BadRequest(w, "Invalid subscription ID")
		return
	}

	// Проверяем что подписка принадлежит пользователю
	subscription, err := h.queries.GetSubscriptionByID(r.Context(), subscriptionID)
	if err != nil {
		response.NotFound(w, "Subscription not found")
		return
	}

	if subscription.UserID != userID {
		response.WriteError(w, http.StatusForbidden, "FORBIDDEN", "You don't have permission to delete this subscription")
		return
	}

	// Деактивируем подписку (мягкое удаление)
	err = h.queries.DeactivateSubscription(r.Context(), db.DeactivateSubscriptionParams{
		ID:          subscriptionID,
		TriggeredAt: pgtype.Int8{Valid: false},
	})
	if err != nil {
		response.InternalError(w, "Failed to delete subscription")
		return
	}

	response.WriteNoContent(w)
}
