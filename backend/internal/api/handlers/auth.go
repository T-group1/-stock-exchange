package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"T_Project/internal/config"
	"T_Project/internal/db"
	"T_Project/internal/service/auth"

	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	queries      db.Querier
	jwtService   *auth.JWTService
	emailService *auth.EmailService
}

func NewAuthHandler(queries db.Querier, cfg *config.Config) *AuthHandler {
	jwtService := auth.NewJWTService(
		cfg.JWT.Secret,
		cfg.JWT.AccessTokenExpiry,
		cfg.JWT.RefreshTokenExpiry,
	)

	emailService := auth.NewEmailService(
		cfg.SMTP.Host,
		cfg.SMTP.Port,
		cfg.SMTP.Username,
		cfg.SMTP.Password,
		cfg.SMTP.From,
	)

	return &AuthHandler{
		queries:      queries,
		jwtService:   jwtService,
		emailService: emailService,
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
	AccessToken  string     `json:"access_token"`
	RefreshToken string     `json:"refresh_token"`
	TokenType    string     `json:"token_type"`
	ExpiresIn    int64      `json:"expires_in"`
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

	// Генерируем токен верификации
	token := generateToken()
	tokenExpiry := time.Now().Add(24 * time.Hour).Unix()

	// Создаем пользователя
	user, err := h.queries.CreateUser(r.Context(), db.CreateUserParams{
		Email:        req.Email,
		Name:         req.Name,
		PasswordHash: string(hashedPassword),
		IsVerified:   pgtype.Bool{Bool: false, Valid: true},
		VerificationToken: pgtype.Text{String: token, Valid: true},
		VerificationTokenExpires: pgtype.Int8{Int64: tokenExpiry, Valid: true},
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// Отправляем email с верификацией
	if err := h.emailService.SendVerificationEmail(req.Email, token); err != nil {
		// Не прерываем регистрацию, если email не отправился
		// Пользователь может запросить повторную отправку позже
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

	// Проверяем, подтвержден ли email
	if !user.IsVerified.Bool {
		respondError(w, http.StatusForbidden, "Please verify your email before logging in")
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

func (h *AuthHandler) Verify(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		respondError(w, http.StatusBadRequest, "Verification token is required")
		return
	}

	currentTime := time.Now().Unix()

	// Верифицируем пользователя
	user, err := h.queries.VerifyUser(r.Context(), db.VerifyUserParams{
		VerificationToken:        token,
		VerificationTokenExpires: currentTime,
	})
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid or expired verification token")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Email successfully verified",
		"user": UserResponse{
			ID:        user.ID.String(),
			Email:     user.Email,
			Name:      user.Name,
			CreatedAt: user.CreatedAt,
		},
	})
}

func generateToken() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback на время если rand не работает
		return hex.EncodeToString([]byte(time.Now().String()))
	}
	return hex.EncodeToString(bytes)
}