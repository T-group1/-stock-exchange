package handlers

import (
	"encoding/json"
	"net/http"

	customMiddleware "T_Project/internal/api/middleware"
	"T_Project/internal/api/response"
	"T_Project/internal/db"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type FavoritesHandler struct {
	queries db.Querier
}

func NewFavoritesHandler(queries db.Querier) *FavoritesHandler {
	return &FavoritesHandler{queries: queries}
}

type FavoritesResponse struct {
	FavoritePairs []string `json:"favorite_pairs"`
}

type AddFavoriteRequest struct {
	CurrencyPair string `json:"currency_pair"`
}

// GET /favorites
func (h *FavoritesHandler) List(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := customMiddleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", http.StatusText(http.StatusUnauthorized), "User not authenticated")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.BadRequest(w, "Invalid user ID")
		return
	}

	pgUUID := pgtype.UUID{Bytes: [16]byte(userID), Valid: true}
	pairs, err := h.queries.GetUserFavoritePairs(r.Context(), pgUUID)
	if err != nil {
		response.InternalError(w, "Failed to fetch favorites")
		return
	}

	response.WriteSuccess(w, FavoritesResponse{FavoritePairs: pairs})
}

// POST /favorites
func (h *FavoritesHandler) Add(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := customMiddleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", http.StatusText(http.StatusUnauthorized), "User not authenticated")
		return
	}

	var req AddFavoriteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if req.CurrencyPair == "" {
		response.BadRequest(w, "currency_pair is required")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.BadRequest(w, "Invalid user ID")
		return
	}

	pgUUID := pgtype.UUID{Bytes: [16]byte(userID), Valid: true}
	err = h.queries.AddFavorite(r.Context(), db.AddFavoriteParams{
		UserID:       pgUUID,
		CurrencyPair: req.CurrencyPair,
	})
	if err != nil {
		response.InternalError(w, "Failed to add favorite")
		return
	}

	response.WriteCreated(w, map[string]string{"status": "added"})
}

// DELETE /favorites/{pair}
func (h *FavoritesHandler) Remove(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := customMiddleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", http.StatusText(http.StatusUnauthorized), "User not authenticated")
		return
	}

	pair := chi.URLParam(r, "pair")
	if pair == "" {
		response.BadRequest(w, "pair is required")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.BadRequest(w, "Invalid user ID")
		return
	}

	pgUUID := pgtype.UUID{Bytes: [16]byte(userID), Valid: true}
	err = h.queries.RemoveFavorite(r.Context(), db.RemoveFavoriteParams{
		UserID:       pgUUID,
		CurrencyPair: pair,
	})
	if err != nil {
		response.InternalError(w, "Failed to remove favorite")
		return
	}

	response.WriteNoContent(w)
}
