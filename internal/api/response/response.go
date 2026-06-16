package response

import (
	"encoding/json"
	"net/http"
)

// Error структура ошибки
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// WriteJSON записывает JSON ответ
func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// WriteError записывает ошибку
func WriteError(w http.ResponseWriter, status int, code, message string) {
	WriteJSON(w, status, Error{
		Code:    code,
		Message: message,
	})
}

// WriteSuccess записывает успешный ответ
func WriteSuccess(w http.ResponseWriter, data interface{}) {
	WriteJSON(w, http.StatusOK, data)
}

// WriteCreated записывает ответ 201 Created
func WriteCreated(w http.ResponseWriter, data interface{}) {
	WriteJSON(w, http.StatusCreated, data)
}

// WriteNoContent записывает ответ 204 No Content
func WriteNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// BadRequest ошибка 400
func BadRequest(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusBadRequest, "BAD_REQUEST", message)
}

// NotFound ошибка 404
func NotFound(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusNotFound, "NOT_FOUND", message)
}

// InternalError ошибка 500
func InternalError(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", message)
}
