package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"evinizinkocu-backend/internal/application"
	"evinizinkocu-backend/internal/domain"
	"evinizinkocu-backend/internal/middleware"

	"github.com/go-chi/chi/v5"
)

type CoachHandler struct {
	coachService *application.CoachService
}

func NewCoachHandler(coachService *application.CoachService) *CoachHandler {
	return &CoachHandler{coachService: coachService}
}

func (h *CoachHandler) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	// Public routes
	r.Post("/api/v1/public/apply", h.CreateApplication)

	// Superadmin protected routes
	r.Group(func(sub chi.Router) {
		sub.Use(authMiddleware)
		sub.Use(middleware.RequireRole(domain.RoleSuperadmin))

		sub.Get("/api/v1/superadmin/applications", h.ListApplications)
		sub.Post("/api/v1/superadmin/applications/{id}/approve", h.ApproveApplication)
		sub.Post("/api/v1/superadmin/applications/{id}/reject", h.RejectApplication)
		sub.Get("/api/v1/superadmin/coaches", h.ListCoaches)
		sub.Post("/api/v1/superadmin/coaches", h.CreateCoach)
		sub.Put("/api/v1/superadmin/coaches/{id}/settings", h.UpdateCoachSettings)
		sub.Post("/api/v1/superadmin/coaches/{id}/resend-credentials", h.ResendCredentials)
	})

	// Coach protected routes
	r.Group(func(sub chi.Router) {
		sub.Use(authMiddleware)
		sub.Use(middleware.RequireRole(domain.RoleCoach))

		sub.Get("/api/v1/coach/profile", h.GetCoachProfile)
		sub.Put("/api/v1/coach/profile", h.UpdateCoachProfile)
	})
}

type applyRequest struct {
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	Phone          string `json:"phone"`
	Email          string `json:"email"`
	City           string `json:"city"`
	Specialization string `json:"specialization"`
	Explanation    string `json:"explanation"`
}

func (h *CoachHandler) CreateApplication(w http.ResponseWriter, r *http.Request) {
	var req applyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.FirstName == "" || req.LastName == "" || req.Phone == "" || req.Email == "" || req.City == "" || req.Specialization == "" || req.Explanation == "" {
		http.Error(w, `{"error": "Tüm alanlar zorunludur"}`, http.StatusBadRequest)
		return
	}

	err := h.coachService.CreateApplication(r.Context(), req.FirstName, req.LastName, req.Phone, req.Email, req.City, req.Specialization, req.Explanation)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message": "Başvurunuz başarıyla alındı"}`))
}

func (h *CoachHandler) ListApplications(w http.ResponseWriter, r *http.Request) {
	apps, err := h.coachService.ListApplications(r.Context())
	if err != nil {
		http.Error(w, `{"error": "Başvurular listelenemedi"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(apps)
}

type approveRequest struct {
	StudentCapacity         int    `json:"student_capacity"`
	AuthStartDate           string `json:"auth_start_date"` // YYYY-MM-DD
	AuthEndDate             string `json:"auth_end_date"`   // YYYY-MM-DD
	PermissionImmediatePush bool   `json:"permission_immediate_push"`
	PermissionScheduledPush bool   `json:"permission_scheduled_push"`
}

func (h *CoachHandler) ApproveApplication(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req approveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.StudentCapacity <= 0 {
		http.Error(w, `{"error": "Kapasite 0 veya negatif olamaz"}`, http.StatusBadRequest)
		return
	}

	start, err := time.Parse("2006-01-02", req.AuthStartDate)
	if err != nil {
		http.Error(w, `{"error": "Geçersiz başlangıç tarihi formatı. YYYY-MM-DD olmalı"}`, http.StatusBadRequest)
		return
	}

	end, err := time.Parse("2006-01-02", req.AuthEndDate)
	if err != nil {
		http.Error(w, `{"error": "Geçersiz bitiş tarihi formatı. YYYY-MM-DD olmalı"}`, http.StatusBadRequest)
		return
	}

	err = h.coachService.ApproveApplication(r.Context(), id, req.StudentCapacity, start, end, req.PermissionImmediatePush, req.PermissionScheduledPush)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Koç başvurusu onaylandı ve hesap oluşturuldu"}`))
}

func (h *CoachHandler) RejectApplication(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := h.coachService.RejectApplication(r.Context(), id)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Koç başvurusu reddedildi"}`))
}

func (h *CoachHandler) ListCoaches(w http.ResponseWriter, r *http.Request) {
	coaches, err := h.coachService.ListCoaches(r.Context())
	if err != nil {
		log.Printf("Error listing coaches: %v\n", err)
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(coaches)
}

func (h *CoachHandler) UpdateCoachSettings(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req approveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	start, err := time.Parse("2006-01-02", req.AuthStartDate)
	if err != nil {
		http.Error(w, `{"error": "Geçersiz başlangıç tarihi"}`, http.StatusBadRequest)
		return
	}

	end, err := time.Parse("2006-01-02", req.AuthEndDate)
	if err != nil {
		http.Error(w, `{"error": "Geçersiz bitiş tarihi"}`, http.StatusBadRequest)
		return
	}

	err = h.coachService.UpdateCoachSettings(r.Context(), id, req.StudentCapacity, start, end, req.PermissionImmediatePush, req.PermissionScheduledPush)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Koç ayarları başarıyla güncellendi"}`))
}

func (h *CoachHandler) ResendCredentials(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := h.coachService.ResendCredentials(r.Context(), id)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Giriş bilgileri tekrar gönderildi"}`))
}


func (h *CoachHandler) GetCoachProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	c, err := h.coachService.GetCoachProfile(r.Context(), userID)
	if err != nil {
		http.Error(w, `{"error": "Koç profili bulunamadı"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(c)
}

type updateProfileRequest struct {
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	Phone          string `json:"phone"`
	City           string `json:"city"`
	Biography      string `json:"biography"`
	Specialization string `json:"specialization"`
	SocialLinks    string `json:"social_links"`
}

func (h *CoachHandler) UpdateCoachProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	var req updateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.FirstName == "" || req.LastName == "" || req.Phone == "" {
		http.Error(w, `{"error": "Ad, soyad ve telefon zorunludur"}`, http.StatusBadRequest)
		return
	}

	err := h.coachService.UpdateCoachProfile(r.Context(), userID, req.FirstName, req.LastName, req.Phone, req.City, req.Biography, req.Specialization, req.SocialLinks)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Profiliniz başarıyla güncellendi"}`))
}

type createCoachRequest struct {
	FirstName               string `json:"first_name"`
	LastName                string `json:"last_name"`
	Phone                   string `json:"phone"`
	Email                   string `json:"email"`
	City                    string `json:"city"`
	Specialization          string `json:"specialization"`
	StudentCapacity         int    `json:"student_capacity"`
	AuthStartDate           string `json:"auth_start_date"` // YYYY-MM-DD
	AuthEndDate             string `json:"auth_end_date"`   // YYYY-MM-DD
	PermissionImmediatePush bool   `json:"permission_immediate_push"`
	PermissionScheduledPush bool   `json:"permission_scheduled_push"`
}

func (h *CoachHandler) CreateCoach(w http.ResponseWriter, r *http.Request) {
	var req createCoachRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.FirstName == "" || req.LastName == "" || req.Email == "" || req.Phone == "" {
		http.Error(w, `{"error": "Ad, soyad, e-posta ve telefon zorunludur"}`, http.StatusBadRequest)
		return
	}

	if req.StudentCapacity <= 0 {
		http.Error(w, `{"error": "Kapasite 0 veya negatif olamaz"}`, http.StatusBadRequest)
		return
	}

	start, err := time.Parse("2006-01-02", req.AuthStartDate)
	if err != nil {
		http.Error(w, `{"error": "Geçersiz başlangıç tarihi formatı. YYYY-MM-DD olmalı"}`, http.StatusBadRequest)
		return
	}

	end, err := time.Parse("2006-01-02", req.AuthEndDate)
	if err != nil {
		http.Error(w, `{"error": "Geçersiz bitiş tarihi formatı. YYYY-MM-DD olmalı"}`, http.StatusBadRequest)
		return
	}

	err = h.coachService.CreateCoach(
		r.Context(),
		req.FirstName,
		req.LastName,
		req.Phone,
		req.Email,
		req.City,
		req.Specialization,
		req.StudentCapacity,
		start,
		end,
		req.PermissionImmediatePush,
		req.PermissionScheduledPush,
	)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message": "Koç hesabı başarıyla oluşturuldu"}`))
}
