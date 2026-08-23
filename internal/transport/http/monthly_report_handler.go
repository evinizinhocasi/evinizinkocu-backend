package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"evinizinkocu-backend/internal/application"
	"evinizinkocu-backend/internal/domain"
	"evinizinkocu-backend/internal/middleware"

	"github.com/go-chi/chi/v5"
)

type MonthlyReportHandler struct {
	service     *application.MonthlyReportService
	studentRepo domain.StudentRepository
}

func NewMonthlyReportHandler(service *application.MonthlyReportService, studentRepo domain.StudentRepository) *MonthlyReportHandler {
	return &MonthlyReportHandler{
		service:     service,
		studentRepo: studentRepo,
	}
}

func (h *MonthlyReportHandler) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Group(func(sub chi.Router) {
		sub.Use(authMiddleware)
		sub.Get("/api/v1/students/{id}/monthly-report", h.GetStudentMonthlyReport)
	})
}

func (h *MonthlyReportHandler) GetStudentMonthlyReport(w http.ResponseWriter, r *http.Request) {
	requestingUserID := middleware.GetUserID(r.Context())
	requestingRole := middleware.GetUserRole(r.Context())
	studentID := chi.URLParam(r, "id")

	// Authorization check
	if requestingRole == domain.RoleStudent && requestingUserID != studentID {
		http.Error(w, `{"error":"bu rapora erişim yetkiniz yok"}`, http.StatusForbidden)
		return
	}
	if requestingRole == domain.RoleCoach {
		student, err := h.studentRepo.GetStudentByID(r.Context(), studentID)
		if err != nil || student == nil || student.CoachID != requestingUserID {
			http.Error(w, `{"error":"bu öğrencinin raporuna erişim yetkiniz yok"}`, http.StatusForbidden)
			return
		}
	}

	year := time.Now().Year()
	month := int(time.Now().Month())

	if yStr := strings.TrimSpace(r.URL.Query().Get("year")); yStr != "" {
		if y, err := strconv.Atoi(yStr); err == nil && y > 2000 {
			year = y
		}
	}
	if mStr := strings.TrimSpace(r.URL.Query().Get("month")); mStr != "" {
		if m, err := strconv.Atoi(mStr); err == nil && m >= 1 && m <= 12 {
			month = m
		}
	}

	report, err := h.service.GetStudentMonthlyReport(r.Context(), studentID, year, month)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}
