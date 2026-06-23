package http

import (
	"encoding/json"
	"net/http"

	"evinizinkocu-backend/internal/application"
	"evinizinkocu-backend/internal/domain"
	"evinizinkocu-backend/internal/middleware"

	"github.com/go-chi/chi/v5"
)

type DashboardHandler struct {
	dashboardService *application.DashboardService
}

func NewDashboardHandler(dashboardService *application.DashboardService) *DashboardHandler {
	return &DashboardHandler{dashboardService: dashboardService}
}

func (h *DashboardHandler) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Group(func(sub chi.Router) {
		sub.Use(authMiddleware)

		sub.With(middleware.RequireRole(domain.RoleSuperadmin)).Get("/api/v1/superadmin/dashboard", h.GetSuperadminDashboard)
		sub.With(middleware.RequireRole(domain.RoleCoach)).Get("/api/v1/coach/dashboard", h.GetCoachDashboard)
		sub.With(middleware.RequireRole(domain.RoleStudent)).Get("/api/v1/student/dashboard", h.GetStudentDashboard)
	})
}

func (h *DashboardHandler) GetSuperadminDashboard(w http.ResponseWriter, r *http.Request) {
	data, err := h.dashboardService.GetSuperadminDashboard(r.Context())
	if err != nil {
		http.Error(w, `{"error": "Gösterge paneli yüklenemedi"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (h *DashboardHandler) GetCoachDashboard(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	data, err := h.dashboardService.GetCoachDashboard(r.Context(), userID)
	if err != nil {
		http.Error(w, `{"error": "Gösterge paneli yüklenemedi"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (h *DashboardHandler) GetStudentDashboard(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	data, err := h.dashboardService.GetStudentDashboard(r.Context(), userID)
	if err != nil {
		http.Error(w, `{"error": "Gösterge paneli yüklenemedi"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
