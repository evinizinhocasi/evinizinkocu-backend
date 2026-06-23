package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"evinizinkocu-backend/internal/application"
	"evinizinkocu-backend/internal/domain"
	"evinizinkocu-backend/internal/middleware"

	"github.com/go-chi/chi/v5"
)

type NotificationHandler struct {
	notifService *application.NotificationService
}

func NewNotificationHandler(notifService *application.NotificationService) *NotificationHandler {
	return &NotificationHandler{notifService: notifService}
}

func (h *NotificationHandler) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Group(func(sub chi.Router) {
		sub.Use(authMiddleware)

		// Inbox
		sub.Get("/api/v1/notifications/inbox", h.GetInbox)
		sub.Put("/api/v1/notifications/inbox/{id}/read", h.MarkAsRead)
		sub.Post("/api/v1/notifications/inbox/read-all", h.MarkAllAsRead)

		// Coach notifications features
		sub.Group(func(coach chi.Router) {
			coach.Use(middleware.RequireRole(domain.RoleCoach))

			coach.Post("/api/v1/notifications/immediate", h.SendImmediateNotification)
			coach.Post("/api/v1/notifications/schedule", h.CreateSchedule)
			coach.Get("/api/v1/notifications/schedules", h.ListSchedules)
			coach.Delete("/api/v1/notifications/schedules/{id}", h.DeleteSchedule)
		})
	})
}

func (h *NotificationHandler) GetInbox(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	inbox, err := h.notifService.GetInbox(r.Context(), userID)
	if err != nil {
		http.Error(w, `{"error": "Bildirim kutusu yüklenemedi"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(inbox)
}

func (h *NotificationHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := h.notifService.MarkAsRead(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "İşlem başarısız oldu"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *NotificationHandler) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	err := h.notifService.MarkAllAsRead(r.Context(), userID)
	if err != nil {
		http.Error(w, `{"error": "İşlem başarısız oldu"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type immediateNotifRequest struct {
	Title           string   `json:"title"`
	Body            string   `json:"body"`
	TargetSelection string   `json:"target_selection"`
	TargetStudentIDs []string `json:"target_student_ids"`
}

func (h *NotificationHandler) SendImmediateNotification(w http.ResponseWriter, r *http.Request) {
	coachID := middleware.GetUserID(r.Context())
	var req immediateNotifRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Title == "" || req.Body == "" || req.TargetSelection == "" {
		http.Error(w, `{"error": "Başlık, mesaj ve hedef seçimi zorunludur"}`, http.StatusBadRequest)
		return
	}

	res, err := h.notifService.SendImmediateNotification(r.Context(), coachID, req.Title, req.Body, req.TargetSelection, req.TargetStudentIDs)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

type createScheduleRequest struct {
	Title            string   `json:"title"`
	Body             string   `json:"body"`
	TargetSelection  string   `json:"target_selection"`
	TargetStudentIDs []string `json:"target_student_ids"`
	ScheduleType     string   `json:"schedule_type"`
	SelectedWeekdays []int    `json:"selected_weekdays"`
	ScheduleTime     string   `json:"schedule_time"`
	StartDate        *string  `json:"start_date"` // YYYY-MM-DD
	EndDate          *string  `json:"end_date"`   // YYYY-MM-DD
}

func (h *NotificationHandler) CreateSchedule(w http.ResponseWriter, r *http.Request) {
	coachID := middleware.GetUserID(r.Context())
	var req createScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Title == "" || req.Body == "" || req.TargetSelection == "" || req.ScheduleType == "" || req.ScheduleTime == "" {
		http.Error(w, `{"error": "Gerekli alanlar eksik"}`, http.StatusBadRequest)
		return
	}

	var start, end *time.Time
	if req.StartDate != nil && *req.StartDate != "" {
		t, err := time.Parse("2006-01-02", *req.StartDate)
		if err == nil {
			start = &t
		}
	}
	if req.EndDate != nil && *req.EndDate != "" {
		t, err := time.Parse("2006-01-02", *req.EndDate)
		if err == nil {
			end = &t
		}
	}

	res, err := h.notifService.CreateSchedule(r.Context(), coachID, req.Title, req.Body, req.TargetSelection, req.TargetStudentIDs, req.ScheduleType, req.SelectedWeekdays, req.ScheduleTime, start, end)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

func (h *NotificationHandler) ListSchedules(w http.ResponseWriter, r *http.Request) {
	coachID := middleware.GetUserID(r.Context())
	list, err := h.notifService.ListSchedulesByCoach(r.Context(), coachID)
	if err != nil {
		http.Error(w, `{"error": "Planlanmış bildirimler alınamadı"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func (h *NotificationHandler) DeleteSchedule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	coachID := middleware.GetUserID(r.Context())

	// Verify ownership
	sched, err := h.notifService.GetScheduleByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "Plan bulunamadı"}`, http.StatusNotFound)
		return
	}

	if sched.CoachID != coachID {
		http.Error(w, `{"error": "Bu planı silme yetkiniz yok"}`, http.StatusForbidden)
		return
	}

	err = h.notifService.DeleteSchedule(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "Silinemedi"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
