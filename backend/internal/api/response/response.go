package response

import (
	"encoding/json"
	"net/http"
)

// Error структура ошибки (ИСПРАВЛЕНО: унифицировано с Error.yaml - добавлено поле Error)
type Error struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// WriteJSON записывает JSON ответ
func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// WriteError записывает ошибку (ИСПРАВЛЕНО: добавлен параметр errorText)
func WriteError(w http.ResponseWriter, status int, errorCode, errorText, message string) {
	WriteJSON(w, status, Error{
		Error:   errorText,
		Message: message,
		Code:    errorCode,
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

// BadRequest ошибка 400 (ИСПРАВЛЕНО: добавлен errorText)
func BadRequest(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusBadRequest, "BAD_REQUEST", http.StatusText(http.StatusBadRequest), message)
}

// NotFound ошибка 404 (ИСПРАВЛЕНО: добавлен errorText)
func NotFound(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusNotFound, "NOT_FOUND", http.StatusText(http.StatusNotFound), message)
}

// InternalError ошибка 500 (ИСПРАВЛЕНО: добавлен errorText)
func InternalError(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", http.StatusText(http.StatusInternalServerError), message)
}
