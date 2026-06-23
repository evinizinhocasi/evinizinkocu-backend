package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"evinizinkocu-backend/internal/application"
	"evinizinkocu-backend/internal/domain"
	"evinizinkocu-backend/internal/middleware"

	"github.com/go-chi/chi/v5"
)

type AuthHandler struct {
	authService *application.AuthService
}

func NewAuthHandler(authService *application.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Post("/api/v1/auth/login", h.Login)
	r.Post("/api/v1/auth/refresh", h.Refresh)
	r.Post("/api/v1/auth/forgot-password", h.ForgotPassword)
	r.Post("/api/v1/auth/reset-password", h.ResetPassword)

	// Protected routes
	r.Group(func(sub chi.Router) {
		sub.Use(authMiddleware)
		sub.Post("/api/v1/auth/logout", h.Logout)
		sub.Post("/api/v1/auth/change-password", h.ChangePassword)
		sub.Post("/api/v1/device-tokens", h.RegisterDeviceToken)
	})
}

type loginRequest struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Identifier == "" || req.Password == "" {
		http.Error(w, `{"error": "E-posta/kullanıcı adı ve şifre zorunludur"}`, http.StatusBadRequest)
		return
	}

	res, err := h.authService.Login(r.Context(), req.Identifier, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			http.Error(w, `{"error": "Giriş bilgileri hatalı"}`, http.StatusUnauthorized)
			return
		}
		if errors.Is(err, domain.ErrPassiveAccount) {
			http.Error(w, `{"error": "Hesabınız pasif durumda. Yöneticinizle iletişime geçin."}`, http.StatusForbidden)
			return
		}
		http.Error(w, `{"error": "Giriş yapılamadı"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.RefreshToken == "" {
		http.Error(w, `{"error": "Refresh token is required"}`, http.StatusBadRequest)
		return
	}

	res, err := h.authService.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		http.Error(w, `{"error": "Geçersiz veya süresi dolmuş oturum"}`, http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token"`
	DeviceToken  string `json:"device_token"`
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req logoutRequest
	_ = json.NewDecoder(r.Body).Decode(&req) // Optional fields

	userID := middleware.GetUserID(r.Context())
	_ = h.authService.Logout(r.Context(), req.RefreshToken, req.DeviceToken, userID)

	w.WriteHeader(http.StatusNoContent)
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var req changePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.CurrentPassword == "" || req.NewPassword == "" {
		http.Error(w, `{"error": "Tüm alanlar zorunludur"}`, http.StatusBadRequest)
		return
	}

	userID := middleware.GetUserID(r.Context())
	err := h.authService.ChangePassword(r.Context(), userID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Şifreniz başarıyla güncellendi"}`))
}

type forgotPasswordRequest struct {
	Email string `json:"email"`
}

func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req forgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		http.Error(w, `{"error": "E-posta adresi zorunludur"}`, http.StatusBadRequest)
		return
	}

	err := h.authService.ForgotPassword(r.Context(), req.Email)
	if err != nil {
		http.Error(w, `{"error": "Şifre sıfırlama talebi başarısız oldu"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Şifre sıfırlama kodu e-posta adresinize gönderildi"}`))
}

type resetPasswordRequest struct {
	Email       string `json:"email"`
	Code        string `json:"code"`
	NewPassword string `json:"new_password"`
}

func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req resetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Code == "" || req.NewPassword == "" {
		http.Error(w, `{"error": "Tüm alanlar zorunludur"}`, http.StatusBadRequest)
		return
	}

	err := h.authService.ResetPassword(r.Context(), req.Email, req.Code, req.NewPassword)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Şifreniz başarıyla sıfırlandı"}`))
}

type registerDeviceTokenRequest struct {
	Token    string `json:"token"`
	Platform string `json:"platform"`
}

func (h *AuthHandler) RegisterDeviceToken(w http.ResponseWriter, r *http.Request) {
	var req registerDeviceTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Token == "" || req.Platform == "" {
		http.Error(w, `{"error": "Token ve platform zorunludur"}`, http.StatusBadRequest)
		return
	}

	userID := middleware.GetUserID(r.Context())
	err := h.authService.RegisterDeviceToken(r.Context(), userID, req.Token, req.Platform)
	if err != nil {
		http.Error(w, `{"error": "Token kaydedilemedi"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Custom fmt helper inside package to avoid compile issues
