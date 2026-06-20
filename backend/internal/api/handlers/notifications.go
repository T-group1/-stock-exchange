package handlers

import (
	"net/http"
	"strconv"
	"time"

	"T_Project/internal/api/middleware"
	"T_Project/internal/api/response"
	"T_Project/internal/db"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// NotificationsHandler обработчик эндпоинтов уведомлений
type NotificationsHandler struct {
	queries db.Querier
}

// NewNotificationsHandler создаёт новый обработчик уведомлений
func NewNotificationsHandler(queries db.Querier) *NotificationsHandler {
	return &NotificationsHandler{queries: queries}
}

// NotificationResponse DTO для ответа
type NotificationResponse struct {
	ID             string `json:"id"`
	SubscriptionID string `json:"subscription_id"`
	Type           string `json:"type"`
	Title          string `json:"title"`
	Message        string `json:"message"`
	IsRead         bool   `json:"is_read"`
	CreatedAt      string `json:"created_at"` // ISO 8601
}

// List возвращает список уведомлений текущего пользователя
func (h *NotificationsHandler) List(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", http.StatusText(http.StatusUnauthorized), "User not authenticated")
		return
	}

	// Парсим UUID
	userID := pgtype.UUID{}
	if err := userID.Scan(userIDStr); err != nil {
		response.BadRequest(w, "Invalid user ID")
		return
	}

	// Пагинация
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	unreadOnlyStr := r.URL.Query().Get("unread_only")

	limit := int32(20)
	offset := int32(0)
	unreadOnly := false

	if limitStr != "" {
		if l, err := strconv.ParseInt(limitStr, 10, 32); err == nil {
			limit = int32(l)
		}
	}
	if offsetStr != "" {
		if o, err := strconv.ParseInt(offsetStr, 10, 32); err == nil {
			offset = int32(o)
		}
	}
	if unreadOnlyStr == "true" {
		unreadOnly = true
	}

	// Получаем уведомления в зависимости от unread_only
	var notifications []db.Notification
	var err error

	if unreadOnly {
		notifications, err = h.queries.GetUserNotificationsUnread(r.Context(), db.GetUserNotificationsUnreadParams{
			UserID: userID,
			Limit:  limit,
			Offset: offset,
		})
	} else {
		notifications, err = h.queries.GetUserNotifications(r.Context(), db.GetUserNotificationsParams{
			UserID: userID,
			Limit:  limit,
			Offset: offset,
		})
	}

	if err != nil {
		response.InternalError(w, "Failed to fetch notifications")
		return
	}

	// Получаем total и unread_count
	total, err := h.queries.GetUserNotificationsCount(r.Context(), userID)
	if err != nil {
		response.InternalError(w, "Failed to get total count")
		return
	}

	unreadCount, err := h.queries.GetUnreadCount(r.Context(), userID)
	if err != nil {
		response.InternalError(w, "Failed to get unread count")
		return
	}

	// Формируем ответ
	result := make([]NotificationResponse, len(notifications))
	for i, n := range notifications {
		// Конвертируем Unix timestamp в ISO 8601
		createdAt := time.Unix(n.CreatedAt, 0).UTC().Format(time.RFC3339)

		result[i] = NotificationResponse{
			ID:             n.ID.String(),
			SubscriptionID: n.SubscriptionID.String(),
			Type:           n.Type,
			Title:          n.Title,
			Message:        n.Message,
			IsRead:         n.IsRead.Bool,
			CreatedAt:      createdAt,
		}
	}

	// Возвращаем total и unread_count
	response.WriteSuccess(w, map[string]interface{}{
		"notifications": result,
		"total":         total,
		"unread_count":  unreadCount,
	})
}

// MarkAsRead помечает уведомление как прочитанное
func (h *NotificationsHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
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

	notificationIDStr := chi.URLParam(r, "id")
	notificationID := pgtype.UUID{}
	if err := notificationID.Scan(notificationIDStr); err != nil {
		response.BadRequest(w, "Invalid notification ID")
		return
	}

	// ИСПРАВЛЕНО: получаем обновлённый объект уведомления
	notification, err := h.queries.MarkNotificationAsRead(r.Context(), db.MarkNotificationAsReadParams{
		ID:     notificationID,
		UserID: userID,
	})
	if err != nil {
		response.InternalError(w, "Failed to mark notification as read")
		return
	}

	// ИСПРАВЛЕНО: возвращаем полный объект уведомления
	createdAt := time.Unix(notification.CreatedAt, 0).UTC().Format(time.RFC3339)

	response.WriteSuccess(w, NotificationResponse{
		ID:             notification.ID.String(),
		SubscriptionID: notification.SubscriptionID.String(),
		Type:           notification.Type,
		Title:          notification.Title,
		Message:        notification.Message,
		IsRead:         notification.IsRead.Bool,
		CreatedAt:      createdAt,
	})
}

// GetUnreadCount возвращает количество непрочитанных уведомлений
func (h *NotificationsHandler) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
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

	count, err := h.queries.GetUnreadCount(r.Context(), userID)
	if err != nil {
		response.InternalError(w, "Failed to get unread count")
		return
	}

	response.WriteSuccess(w, map[string]int64{
		"unread_count": count,
	})
}
