package handlers

import (
	"net/http"

	"T_Project/internal/api/middleware"
	"T_Project/internal/api/response"
	"T_Project/internal/db"

	"github.com/jackc/pgx/v5/pgtype"
)

// NotificationSettingsHandler обработчик эндпоинтов настроек уведомлений
type NotificationSettingsHandler struct {
	queries db.Querier
}

// NewNotificationSettingsHandler создаёт новый обработчик настроек
func NewNotificationSettingsHandler(queries db.Querier) *NotificationSettingsHandler {
	return &NotificationSettingsHandler{queries: queries}
}

// NotificationSettingsResponse DTO для ответа
type NotificationSettingsResponse struct {
	EmailEnabled    bool    `json:"email_enabled"`
	BrowserEnabled  bool    `json:"browser_enabled"`
	QuietHoursStart *string `json:"quiet_hours_start,omitempty"`
	QuietHoursEnd   *string `json:"quiet_hours_end,omitempty"`
}

// UpdateSettingsRequest запрос на обновление настроек
type UpdateSettingsRequest struct {
	EmailEnabled    *bool   `json:"email_enabled"`
	BrowserEnabled  *bool   `json:"browser_enabled"`
	QuietHoursStart *string `json:"quiet_hours_start"`
	QuietHoursEnd   *string `json:"quiet_hours_end"`
}

// Get возвращает текущие настройки уведомлений
func (h *NotificationSettingsHandler) Get(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", http.StatusText(http.StatusUnauthorized), "User not authenticated")
		return
	}

	userID := pgtype.UUID{}
	if err := userID.Scan(userIDStr); err != nil {
		response.BadRequest(w, "Invalid user ID")
		return
	}

	settings, err := h.queries.GetNotificationSettings(r.Context(), userID)
	if err != nil {
		// Если настроек нет, возвращаем значения по умолчанию
		response.WriteSuccess(w, NotificationSettingsResponse{
			EmailEnabled:   true,
			BrowserEnabled: true,
		})
		return
	}

	resp := NotificationSettingsResponse{
		EmailEnabled:   settings.EmailEnabled.Bool,
		BrowserEnabled: settings.BrowserEnabled.Bool,
	}

	if settings.QuietHoursStart.Valid {
		resp.QuietHoursStart = &settings.QuietHoursStart.String
	}
	if settings.QuietHoursEnd.Valid {
		resp.QuietHoursEnd = &settings.QuietHoursEnd.String
	}

	response.WriteSuccess(w, resp)
}

// Update обновляет настройки уведомлений
func (h *NotificationSettingsHandler) Update(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", http.StatusText(http.StatusUnauthorized), "User not authenticated")
		return
	}

	userID := pgtype.UUID{}
	if err := userID.Scan(userIDStr); err != nil {
		response.BadRequest(w, "Invalid user ID")
		return
	}

	var req UpdateSettingsRequest
	if err := parseJSON(r, &req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	// Получаем текущие настройки (если есть)
	currentSettings, err := h.queries.GetNotificationSettings(r.Context(), userID)

	// Значения по умолчанию
	emailEnabled := true
	browserEnabled := true
	quietHoursStart := pgtype.Text{Valid: false}
	quietHoursEnd := pgtype.Text{Valid: false}

	// Если настройки уже существуют, используем их как базу
	if err == nil {
		emailEnabled = currentSettings.EmailEnabled.Bool
		browserEnabled = currentSettings.BrowserEnabled.Bool
		quietHoursStart = currentSettings.QuietHoursStart
		quietHoursEnd = currentSettings.QuietHoursEnd
	}

	// Обновляем только те поля, которые пришли в запросе
	if req.EmailEnabled != nil {
		emailEnabled = *req.EmailEnabled
	}
	if req.BrowserEnabled != nil {
		browserEnabled = *req.BrowserEnabled
	}
	if req.QuietHoursStart != nil {
		quietHoursStart = pgtype.Text{String: *req.QuietHoursStart, Valid: true}
	}
	if req.QuietHoursEnd != nil {
		quietHoursEnd = pgtype.Text{String: *req.QuietHoursEnd, Valid: true}
	}

	settings, err := h.queries.UpsertNotificationSettings(r.Context(), db.UpsertNotificationSettingsParams{
		UserID:          userID,
		EmailEnabled:    pgtype.Bool{Bool: emailEnabled, Valid: true},
		BrowserEnabled:  pgtype.Bool{Bool: browserEnabled, Valid: true},
		QuietHoursStart: quietHoursStart,
		QuietHoursEnd:   quietHoursEnd,
	})
	if err != nil {
		response.InternalError(w, "Failed to update notification settings")
		return
	}

	resp := NotificationSettingsResponse{
		EmailEnabled:   settings.EmailEnabled.Bool,
		BrowserEnabled: settings.BrowserEnabled.Bool,
	}

	if settings.QuietHoursStart.Valid {
		resp.QuietHoursStart = &settings.QuietHoursStart.String
	}
	if settings.QuietHoursEnd.Valid {
		resp.QuietHoursEnd = &settings.QuietHoursEnd.String
	}

	response.WriteSuccess(w, resp)
}
