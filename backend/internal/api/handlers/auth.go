package handlers

import (
	"net/http"

	"T_Project/internal/config"
	"T_Project/internal/db"
	"T_Project/internal/service/auth"

	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	queries    db.Querier
	jwtService *auth.JWTService
}

func NewAuthHandler(queries db.Querier, cfg *config.Config) *AuthHandler {
	jwtService := auth.NewJWTService(
		cfg.JWT.Secret,
		cfg.JWT.AccessTokenExpiry,
		cfg.JWT.RefreshTokenExpiry,
	)

	return &AuthHandler{
		queries:    queries,
		jwtService: jwtService,
	}
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	TokenType    string       `json:"token_type"`
	ExpiresIn    int64        `json:"expires_in"`
	User         UserResponse `json:"user"`
}

type UserResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	CreatedAt int64  `json:"created_at"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := parseJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" || req.Name == "" {
		respondError(w, http.StatusBadRequest, "Email, password and name are required")
		return
	}

	if len(req.Password) < 8 {
		respondError(w, http.StatusBadRequest, "Password must be at least 8 characters")
		return
	}

	// Проверяем, существует ли пользователь
	_, err := h.queries.GetUserByEmail(r.Context(), req.Email)
	if err == nil {
		respondError(w, http.StatusConflict, "User with this email already exists")
		return
	}

	// Хешируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	// Создаем пользователя (используем реальные поля из db.CreateUserParams)
	user, err := h.queries.CreateUser(r.Context(), db.CreateUserParams{
		Email:        req.Email,
		Name:         req.Name,
		PasswordHash: string(hashedPassword),
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// Генерируем токены
	accessToken, err := h.jwtService.GenerateAccessToken(user.ID.String(), user.Email)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate access token")
		return
	}

	refreshToken, err := h.jwtService.GenerateRefreshToken(user.ID.String(), user.Email)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate refresh token")
		return
	}

	respondJSON(w, http.StatusCreated, AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		User: UserResponse{
			ID:        user.ID.String(),
			Email:     user.Email,
			Name:      user.Name,
			CreatedAt: user.CreatedAt,
		},
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := parseJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	// Получаем пользователя
	user, err := h.queries.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// Проверяем пароль
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		respondError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// Генерируем токены
	accessToken, err := h.jwtService.GenerateAccessToken(user.ID.String(), user.Email)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate access token")
		return
	}

	refreshToken, err := h.jwtService.GenerateRefreshToken(user.ID.String(), user.Email)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate refresh token")
		return
	}

	respondJSON(w, http.StatusOK, AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		User: UserResponse{
			ID:        user.ID.String(),
			Email:     user.Email,
			Name:      user.Name,
			CreatedAt: user.CreatedAt,
		},
	})
}