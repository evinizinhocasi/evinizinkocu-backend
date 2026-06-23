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

type TrackersHandler struct {
	trackersService *application.TrackersService
	studentService  *application.StudentService
}

func NewTrackersHandler(
	trackersService *application.TrackersService,
	studentService *application.StudentService,
) *TrackersHandler {
	return &TrackersHandler{
		trackersService: trackersService,
		studentService:  studentService,
	}
}

func (h *TrackersHandler) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Group(func(sub chi.Router) {
		sub.Use(authMiddleware)

		// Question Solving
		sub.Post("/api/v1/question-solving", h.CreateQuestionSolving)
		sub.Get("/api/v1/question-solving", h.ListQuestionSolving)
		sub.Put("/api/v1/question-solving/{id}", h.UpdateQuestionSolving)
		sub.Delete("/api/v1/question-solving/{id}", h.DeleteQuestionSolving)

		// Homework
		sub.Post("/api/v1/homework", h.CreateHomework)
		sub.Get("/api/v1/homework", h.ListHomework)
		sub.Put("/api/v1/homework/{id}/status", h.UpdateHomeworkStatus)
		sub.Delete("/api/v1/homework/{id}", h.DeleteHomework)

		// Resources
		sub.Post("/api/v1/resources", h.CreateResource)
		sub.Get("/api/v1/resources", h.ListResources)
		sub.Put("/api/v1/resources/{id}/progress", h.UpdateResourceProgress)
		sub.Delete("/api/v1/resources/{id}", h.DeleteResource)

		// Timetable
		sub.Post("/api/v1/timetable", h.SaveTimetable)
		sub.Get("/api/v1/timetable", h.GetTimetable)

		// Weekly Plan
		sub.Post("/api/v1/weekly-plans", h.CreateWeeklyPlanItem)
		sub.Get("/api/v1/weekly-plans", h.ListWeeklyPlanItems)
		sub.Put("/api/v1/weekly-plans/{id}", h.UpdateWeeklyPlanItem)
		sub.Delete("/api/v1/weekly-plans/{id}", h.DeleteWeeklyPlanItem)
		sub.Post("/api/v1/weekly-plans/copy", h.CopyWeeklyPlan)

		// Missing Topics
		sub.Post("/api/v1/missing-topics", h.CreateMissingTopic)
		sub.Get("/api/v1/missing-topics", h.ListMissingTopics)
		sub.Put("/api/v1/missing-topics/{id}", h.UpdateMissingTopic)
		sub.Delete("/api/v1/missing-topics/{id}", h.DeleteMissingTopic)

		// Goals
		sub.Post("/api/v1/goals", h.CreateGoal)
		sub.Get("/api/v1/goals", h.ListGoals)
		sub.Put("/api/v1/goals/{id}/progress", h.UpdateGoalProgress)
		sub.Delete("/api/v1/goals/{id}", h.DeleteGoal)

		// Meetings
		sub.Post("/api/v1/meetings", h.CreateMeeting)
		sub.Get("/api/v1/meetings", h.ListMeetings)
		sub.Put("/api/v1/meetings/{id}", h.UpdateMeeting)
		sub.Delete("/api/v1/meetings/{id}", h.DeleteMeeting)
	})
}

// Tenant checking helper
func (h *TrackersHandler) checkTenant(w http.ResponseWriter, r *http.Request, studentID string) (*domain.Student, bool) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	student, err := h.studentService.GetStudentByID(r.Context(), studentID)
	if err != nil {
		http.Error(w, `{"error": "Öğrenci bulunamadı"}`, http.StatusNotFound)
		return nil, false
	}

	if role == domain.RoleCoach && student.CoachID != userID {
		http.Error(w, `{"error": "Bu öğrenci size ait değil"}`, http.StatusForbidden)
		return nil, false
	}

	if role == domain.RoleStudent && student.ID != userID {
		http.Error(w, `{"error": "Yetkiniz yok"}`, http.StatusForbidden)
		return nil, false
	}

	return student, true
}

// --- Question Solving Tracker ---

type createQuestionRequest struct {
	StudentID string  `json:"student_id"`
	Date      string  `json:"date"` // YYYY-MM-DD
	SubjectID string  `json:"subject_id"`
	TopicID   *string `json:"topic_id"`
	Correct   int     `json:"correct"`
	Incorrect int     `json:"incorrect"`
	Blank     int     `json:"blank"`
	Note      string  `json:"note"`
}

func (h *TrackersHandler) CreateQuestionSolving(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	var req createQuestionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.StudentID == "" || req.Date == "" || req.SubjectID == "" {
		http.Error(w, `{"error": "Gerekli alanlar eksik"}`, http.StatusBadRequest)
		return
	}

	if _, ok := h.checkTenant(w, r, req.StudentID); !ok {
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		http.Error(w, `{"error": "Tarih formatı YYYY-MM-DD olmalı"}`, http.StatusBadRequest)
		return
	}

	res, err := h.trackersService.CreateQuestionSolving(r.Context(), req.StudentID, userID, date, req.SubjectID, req.TopicID, req.Correct, req.Incorrect, req.Blank, req.Note)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

func (h *TrackersHandler) ListQuestionSolving(w http.ResponseWriter, r *http.Request) {
	studentID := r.URL.Query().Get("student_id")
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	if studentID == "" {
		http.Error(w, `{"error": "student_id parametresi zorunludur"}`, http.StatusBadRequest)
		return
	}

	if _, ok := h.checkTenant(w, r, studentID); !ok {
		return
	}

	var startDate, endDate *time.Time
	if startDateStr != "" {
		t, err := time.Parse("2006-01-02", startDateStr)
		if err == nil {
			startDate = &t
		}
	}
	if endDateStr != "" {
		t, err := time.Parse("2006-01-02", endDateStr)
		if err == nil {
			endDate = &t
		}
	}

	list, err := h.trackersService.ListQuestionSolvingByStudent(r.Context(), studentID, startDate, endDate)
	if err != nil {
		http.Error(w, `{"error": "Soru çözümleri listelenemedi"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func (h *TrackersHandler) UpdateQuestionSolving(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	entry, err := h.trackersService.GetQuestionSolvingByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "Bulunamadı"}`, http.StatusNotFound)
		return
	}

	if _, ok := h.checkTenant(w, r, entry.StudentID); !ok {
		return
	}

	var req createQuestionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		http.Error(w, `{"error": "Tarih formatı YYYY-MM-DD olmalı"}`, http.StatusBadRequest)
		return
	}

	err = h.trackersService.UpdateQuestionSolving(r.Context(), id, date, req.SubjectID, req.TopicID, req.Correct, req.Incorrect, req.Blank, req.Note)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *TrackersHandler) DeleteQuestionSolving(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	entry, err := h.trackersService.GetQuestionSolvingByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "Bulunamadı"}`, http.StatusNotFound)
		return
	}

	if _, ok := h.checkTenant(w, r, entry.StudentID); !ok {
		return
	}

	err = h.trackersService.DeleteQuestionSolving(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "Silinemedi"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// --- Homework Tracker ---

type createHomeworkRequest struct {
	StudentID   string  `json:"student_id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	SubjectID   string  `json:"subject_id"`
	TopicID     *string `json:"topic_id"`
	Source      string  `json:"source"`
	PageRange   string  `json:"page_range"`
	URL         string  `json:"url"`
	StartDate   string  `json:"start_date"` // YYYY-MM-DD
	DueDate     string  `json:"due_date"`   // YYYY-MM-DD
}

func (h *TrackersHandler) CreateHomework(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())
	if role != domain.RoleCoach && role != domain.RoleSuperadmin {
		http.Error(w, `{"error": "Sadece koçlar ödev verebilir"}`, http.StatusForbidden)
		return
	}

	var req createHomeworkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.StudentID == "" || req.Title == "" || req.SubjectID == "" || req.StartDate == "" || req.DueDate == "" {
		http.Error(w, `{"error": "Gerekli alanlar eksik"}`, http.StatusBadRequest)
		return
	}

	if _, ok := h.checkTenant(w, r, req.StudentID); !ok {
		return
	}

	start, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		http.Error(w, `{"error": "Başlangıç tarihi geçersiz"}`, http.StatusBadRequest)
		return
	}

	due, err := time.Parse("2006-01-02", req.DueDate)
	if err != nil {
		http.Error(w, `{"error": "Bitiş tarihi geçersiz"}`, http.StatusBadRequest)
		return
	}

	res, err := h.trackersService.CreateHomework(r.Context(), req.StudentID, userID, req.Title, req.Description, req.SubjectID, req.TopicID, req.Source, req.PageRange, req.URL, start, due)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

func (h *TrackersHandler) ListHomework(w http.ResponseWriter, r *http.Request) {
	studentID := r.URL.Query().Get("student_id")
	if studentID == "" {
		http.Error(w, `{"error": "student_id parametresi zorunludur"}`, http.StatusBadRequest)
		return
	}

	if _, ok := h.checkTenant(w, r, studentID); !ok {
		return
	}

	list, err := h.trackersService.ListHomeworkByStudent(r.Context(), studentID)
	if err != nil {
		http.Error(w, `{"error": "Ödevler listelenemedi"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

type updateHomeworkStatusRequest struct {
	Status      string `json:"status"`
	Explanation string `json:"explanation"` // for rejection notes
}

func (h *TrackersHandler) UpdateHomeworkStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	var req updateHomeworkStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	err := h.trackersService.UpdateHomeworkStatus(r.Context(), id, userID, role, req.Status, req.Explanation)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *TrackersHandler) DeleteHomework(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	// Since delete is coach only
	role := middleware.GetUserRole(r.Context())
	if role != domain.RoleCoach && role != domain.RoleSuperadmin {
		http.Error(w, `{"error": "Forbidden"}`, http.StatusForbidden)
		return
	}

	err := h.trackersService.DeleteHomework(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "Silinemedi"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// --- Student Resources ---

type createResourceRequest struct {
	StudentID      string `json:"student_id"`
	Name           string `json:"name"`
	Publisher      string `json:"publisher"`
	SubjectID      string `json:"subject_id"`
	Description    string `json:"description"`
	TotalPages     int    `json:"total_pages"`
	CompletedPages int    `json:"completed_pages"`
	Status         string `json:"status"`
}

func (h *TrackersHandler) CreateResource(w http.ResponseWriter, r *http.Request) {
	var req createResourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.StudentID == "" || req.Name == "" || req.Publisher == "" || req.SubjectID == "" || req.TotalPages <= 0 {
		http.Error(w, `{"error": "Eksik alanlar var"}`, http.StatusBadRequest)
		return
	}

	if _, ok := h.checkTenant(w, r, req.StudentID); !ok {
		return
	}

	res, err := h.trackersService.CreateResource(r.Context(), req.StudentID, req.Name, req.Publisher, req.SubjectID, req.Description, req.TotalPages, req.CompletedPages, req.Status)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

func (h *TrackersHandler) ListResources(w http.ResponseWriter, r *http.Request) {
	studentID := r.URL.Query().Get("student_id")
	if studentID == "" {
		http.Error(w, `{"error": "student_id parametresi zorunludur"}`, http.StatusBadRequest)
		return
	}

	if _, ok := h.checkTenant(w, r, studentID); !ok {
		return
	}

	list, err := h.trackersService.ListResourcesByStudent(r.Context(), studentID)
	if err != nil {
		http.Error(w, `{"error": "Kaynaklar listelenemedi"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

type updateResourceRequest struct {
	CompletedPages     int    `json:"completed_pages"`
	ProgressPercentage int    `json:"progress_percentage"`
	Status             string `json:"status"`
}

func (h *TrackersHandler) UpdateResourceProgress(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req updateResourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	err := h.trackersService.UpdateResourceProgress(r.Context(), id, req.CompletedPages, req.ProgressPercentage, req.Status)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *TrackersHandler) DeleteResource(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := h.trackersService.DeleteResource(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "Silinemedi"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// --- School Timetable ---

type saveTimetableRequest struct {
	StudentID string                        `json:"student_id"`
	Entries   []*domain.SchoolTimetableEntry `json:"entries"`
}

func (h *TrackersHandler) SaveTimetable(w http.ResponseWriter, r *http.Request) {
	var req saveTimetableRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.StudentID == "" {
		http.Error(w, `{"error": "student_id zorunludur"}`, http.StatusBadRequest)
		return
	}

	if _, ok := h.checkTenant(w, r, req.StudentID); !ok {
		return
	}

	err := h.trackersService.SaveTimetable(r.Context(), req.StudentID, req.Entries)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *TrackersHandler) GetTimetable(w http.ResponseWriter, r *http.Request) {
	studentID := r.URL.Query().Get("student_id")
	if studentID == "" {
		http.Error(w, `{"error": "student_id parametresi zorunludur"}`, http.StatusBadRequest)
		return
	}

	if _, ok := h.checkTenant(w, r, studentID); !ok {
		return
	}

	list, err := h.trackersService.GetTimetableByStudent(r.Context(), studentID)
	if err != nil {
		http.Error(w, `{"error": "Haftalık ders programı alınamadı"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

// --- Weekly Coaching Plan ---

type createPlanItemRequest struct {
	StudentID string  `json:"student_id"`
	Date      string  `json:"date"` // YYYY-MM-DD
	StartTime string  `json:"start_time"`
	EndTime   string  `json:"end_time"`
	Title     string  `json:"title"`
	SubjectID *string `json:"subject_id"`
	TopicID   *string `json:"topic_id"`
	Note      string  `json:"note"`
	URL       string  `json:"url"`
}

func (h *TrackersHandler) CreateWeeklyPlanItem(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())
	if role != domain.RoleCoach && role != domain.RoleSuperadmin {
		http.Error(w, `{"error": "Yalnızca koç haftalık plan oluşturabilir"}`, http.StatusForbidden)
		return
	}

	var req createPlanItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.StudentID == "" || req.Date == "" || req.StartTime == "" || req.EndTime == "" || req.Title == "" {
		http.Error(w, `{"error": "Gerekli alanlar eksik"}`, http.StatusBadRequest)
		return
	}

	if _, ok := h.checkTenant(w, r, req.StudentID); !ok {
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		http.Error(w, `{"error": "Tarih formatı YYYY-MM-DD olmalı"}`, http.StatusBadRequest)
		return
	}

	res, err := h.trackersService.CreateWeeklyPlanItem(r.Context(), req.StudentID, userID, date, req.StartTime, req.EndTime, req.Title, req.SubjectID, req.TopicID, req.Note, req.URL)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

func (h *TrackersHandler) ListWeeklyPlanItems(w http.ResponseWriter, r *http.Request) {
	studentID := r.URL.Query().Get("student_id")
	startStr := r.URL.Query().Get("start_date")
	endStr := r.URL.Query().Get("end_date")

	if studentID == "" || startStr == "" || endStr == "" {
		http.Error(w, `{"error": "student_id, start_date ve end_date zorunludur"}`, http.StatusBadRequest)
		return
	}

	if _, ok := h.checkTenant(w, r, studentID); !ok {
		return
	}

	start, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		http.Error(w, `{"error": "Geçersiz başlangıç tarihi"}`, http.StatusBadRequest)
		return
	}

	end, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		http.Error(w, `{"error": "Geçersiz bitiş tarihi"}`, http.StatusBadRequest)
		return
	}

	list, err := h.trackersService.ListWeeklyPlanItems(r.Context(), studentID, start, end)
	if err != nil {
		http.Error(w, `{"error": "Haftalık plan listelenemedi"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func (h *TrackersHandler) UpdateWeeklyPlanItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	// Since update is coach only
	role := middleware.GetUserRole(r.Context())
	if role != domain.RoleCoach && role != domain.RoleSuperadmin {
		http.Error(w, `{"error": "Forbidden"}`, http.StatusForbidden)
		return
	}

	var req createPlanItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		http.Error(w, `{"error": "Geçersiz tarih"}`, http.StatusBadRequest)
		return
	}

	err = h.trackersService.UpdateWeeklyPlanItem(r.Context(), id, date, req.StartTime, req.EndTime, req.Title, req.SubjectID, req.TopicID, req.Note, req.URL)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *TrackersHandler) DeleteWeeklyPlanItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	role := middleware.GetUserRole(r.Context())
	if role != domain.RoleCoach && role != domain.RoleSuperadmin {
		http.Error(w, `{"error": "Forbidden"}`, http.StatusForbidden)
		return
	}

	err := h.trackersService.DeleteWeeklyPlanItem(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "Silinemedi"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type copyWeekRequest struct {
	StudentID string `json:"student_id"`
	FromStart string `json:"from_start"` // YYYY-MM-DD
	ToStart   string `json:"to_start"`   // YYYY-MM-DD
}

func (h *TrackersHandler) CopyWeeklyPlan(w http.ResponseWriter, r *http.Request) {
	role := middleware.GetUserRole(r.Context())
	if role != domain.RoleCoach && role != domain.RoleSuperadmin {
		http.Error(w, `{"error": "Forbidden"}`, http.StatusForbidden)
		return
	}

	var req copyWeekRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.StudentID == "" || req.FromStart == "" || req.ToStart == "" {
		http.Error(w, `{"error": "student_id, from_start ve to_start zorunludur"}`, http.StatusBadRequest)
		return
	}

	if _, ok := h.checkTenant(w, r, req.StudentID); !ok {
		return
	}

	from, err := time.Parse("2006-01-02", req.FromStart)
	if err != nil {
		http.Error(w, `{"error": "from_start tarihi geçersiz"}`, http.StatusBadRequest)
		return
	}

	to, err := time.Parse("2006-01-02", req.ToStart)
	if err != nil {
		http.Error(w, `{"error": "to_start tarihi geçersiz"}`, http.StatusBadRequest)
		return
	}

	err = h.trackersService.CopyWeekPlan(r.Context(), req.StudentID, from, to)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Haftalık program başarıyla kopyalandı"}`))
}

// --- Missing Topics ---

type createMissingTopicRequest struct {
	StudentID    string  `json:"student_id"`
	SubjectID    string  `json:"subject_id"`
	TopicID      string  `json:"topic_id"`
	Description  string  `json:"description"`
	Priority     string  `json:"priority"` // 'low', 'medium', 'high'
	Status       string  `json:"status"`   // 'identified', 'in_progress', 'resolved'
	TargetDate   *string `json:"target_date"`
	SolutionText string  `json:"solution_text"`
	URL          string  `json:"url"`
}

func (h *TrackersHandler) CreateMissingTopic(w http.ResponseWriter, r *http.Request) {
	role := middleware.GetUserRole(r.Context())
	if role != domain.RoleCoach && role != domain.RoleSuperadmin {
		http.Error(w, `{"error": "Yalnızca koç zayıf konu ekleyebilir"}`, http.StatusForbidden)
		return
	}

	var req createMissingTopicRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.StudentID == "" || req.SubjectID == "" || req.TopicID == "" {
		http.Error(w, `{"error": "Ders ve konu zorunludur"}`, http.StatusBadRequest)
		return
	}

	if _, ok := h.checkTenant(w, r, req.StudentID); !ok {
		return
	}

	var targetDate *time.Time
	if req.TargetDate != nil && *req.TargetDate != "" {
		t, err := time.Parse("2006-01-02", *req.TargetDate)
		if err == nil {
			targetDate = &t
		}
	}

	res, err := h.trackersService.CreateMissingTopic(r.Context(), req.StudentID, req.SubjectID, req.TopicID, req.Description, req.Priority, req.Status, targetDate, req.SolutionText, req.URL)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

func (h *TrackersHandler) ListMissingTopics(w http.ResponseWriter, r *http.Request) {
	studentID := r.URL.Query().Get("student_id")
	if studentID == "" {
		http.Error(w, `{"error": "student_id parametresi zorunludur"}`, http.StatusBadRequest)
		return
	}

	if _, ok := h.checkTenant(w, r, studentID); !ok {
		return
	}

	list, err := h.trackersService.ListMissingTopicsByStudent(r.Context(), studentID)
	if err != nil {
		http.Error(w, `{"error": "Eksik konular listelenemedi"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func (h *TrackersHandler) UpdateMissingTopic(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	role := middleware.GetUserRole(r.Context())
	if role != domain.RoleCoach && role != domain.RoleSuperadmin {
		http.Error(w, `{"error": "Forbidden"}`, http.StatusForbidden)
		return
	}

	var req createMissingTopicRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	var targetDate *time.Time
	if req.TargetDate != nil && *req.TargetDate != "" {
		t, err := time.Parse("2006-01-02", *req.TargetDate)
		if err == nil {
			targetDate = &t
		}
	}

	err := h.trackersService.UpdateMissingTopic(r.Context(), id, req.SubjectID, req.TopicID, req.Description, req.Priority, req.Status, targetDate, req.SolutionText, req.URL)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *TrackersHandler) DeleteMissingTopic(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	role := middleware.GetUserRole(r.Context())
	if role != domain.RoleCoach && role != domain.RoleSuperadmin {
		http.Error(w, `{"error": "Forbidden"}`, http.StatusForbidden)
		return
	}

	err := h.trackersService.DeleteMissingTopic(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "Silinemedi"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// --- Goals ---

type createGoalRequest struct {
	StudentID   string  `json:"student_id"`
	Type        string  `json:"type"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	TargetValue float64 `json:"target_value"`
	Unit        string  `json:"unit"`
	StartDate   string  `json:"start_date"`  // YYYY-MM-DD
	TargetDate  string  `json:"target_date"` // YYYY-MM-DD
}

func (h *TrackersHandler) CreateGoal(w http.ResponseWriter, r *http.Request) {
	role := middleware.GetUserRole(r.Context())
	if role != domain.RoleCoach && role != domain.RoleSuperadmin {
		http.Error(w, `{"error": "Yalnızca koç hedef oluşturabilir"}`, http.StatusForbidden)
		return
	}

	var req createGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.StudentID == "" || req.Type == "" || req.Title == "" || req.TargetValue <= 0 || req.StartDate == "" || req.TargetDate == "" {
		http.Error(w, `{"error": "Gerekli alanlar eksik"}`, http.StatusBadRequest)
		return
	}

	if _, ok := h.checkTenant(w, r, req.StudentID); !ok {
		return
	}

	start, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		http.Error(w, `{"error": "Başlangıç tarihi geçersiz"}`, http.StatusBadRequest)
		return
	}

	target, err := time.Parse("2006-01-02", req.TargetDate)
	if err != nil {
		http.Error(w, `{"error": "Bitiş tarihi geçersiz"}`, http.StatusBadRequest)
		return
	}

	res, err := h.trackersService.CreateGoal(r.Context(), req.StudentID, req.Type, req.Title, req.Description, req.TargetValue, req.Unit, start, target)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

func (h *TrackersHandler) ListGoals(w http.ResponseWriter, r *http.Request) {
	studentID := r.URL.Query().Get("student_id")
	if studentID == "" {
		http.Error(w, `{"error": "student_id parametresi zorunludur"}`, http.StatusBadRequest)
		return
	}

	if _, ok := h.checkTenant(w, r, studentID); !ok {
		return
	}

	list, err := h.trackersService.ListGoalsByStudent(r.Context(), studentID)
	if err != nil {
		http.Error(w, `{"error": "Hedefler listelenemedi"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

type updateGoalProgressRequest struct {
	CurrentValue float64 `json:"current_value"`
	Status       string  `json:"status"`
}

func (h *TrackersHandler) UpdateGoalProgress(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req updateGoalProgressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	err := h.trackersService.UpdateGoalManualProgress(r.Context(), id, req.CurrentValue, req.Status)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *TrackersHandler) DeleteGoal(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	role := middleware.GetUserRole(r.Context())
	if role != domain.RoleCoach && role != domain.RoleSuperadmin {
		http.Error(w, `{"error": "Forbidden"}`, http.StatusForbidden)
		return
	}

	err := h.trackersService.DeleteGoal(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "Silinemedi"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// --- Meetings ---

type createMeetingRequest struct {
	StudentID       string  `json:"student_id"`
	MeetingDate     string  `json:"meeting_date"` // YYYY-MM-DD HH:MM:SS (interpreted in Europe/Istanbul by presentation, parsed as UTC/time object)
	DurationMinutes int     `json:"duration_minutes"`
	Notes           string  `json:"notes"`
	MeetingURL      string  `json:"meeting_url"`
	NextMeetingDate *string `json:"next_meeting_date"`
}

func (h *TrackersHandler) CreateMeeting(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())
	if role != domain.RoleCoach && role != domain.RoleSuperadmin {
		http.Error(w, `{"error": "Forbidden"}`, http.StatusForbidden)
		return
	}

	var req createMeetingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.StudentID == "" || req.MeetingDate == "" || req.DurationMinutes <= 0 {
		http.Error(w, `{"error": "Gerekli alanlar eksik"}`, http.StatusBadRequest)
		return
	}

	if _, ok := h.checkTenant(w, r, req.StudentID); !ok {
		return
	}

	date, err := time.Parse(time.RFC3339, req.MeetingDate)
	if err != nil {
		// Fallback parse
		date, err = time.Parse("2006-01-02 15:04:05", req.MeetingDate)
		if err != nil {
			http.Error(w, `{"error": "Geçersiz görüşme tarihi. RFC3339 veya YYYY-MM-DD HH:MM:SS olmalı"}`, http.StatusBadRequest)
			return
		}
	}

	var nextMeeting *time.Time
	if req.NextMeetingDate != nil && *req.NextMeetingDate != "" {
		t, err := time.Parse(time.RFC3339, *req.NextMeetingDate)
		if err == nil {
			nextMeeting = &t
		} else {
			t2, err2 := time.Parse("2006-01-02 15:04:05", *req.NextMeetingDate)
			if err2 == nil {
				nextMeeting = &t2
			}
		}
	}

	res, err := h.trackersService.CreateMeeting(r.Context(), req.StudentID, userID, date, req.DurationMinutes, req.Notes, req.MeetingURL, nextMeeting)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

func (h *TrackersHandler) ListMeetings(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	studentID := r.URL.Query().Get("student_id")

	var list []*domain.Meeting
	var err error

	if studentID != "" {
		if _, ok := h.checkTenant(w, r, studentID); !ok {
			return
		}
		list, err = h.trackersService.ListMeetingsByStudent(r.Context(), studentID)
	} else {
		// If no student specified, return all coach meetings if logged-in is coach
		if role == domain.RoleCoach {
			list, err = h.trackersService.ListMeetingsByCoach(r.Context(), userID)
		} else {
			http.Error(w, `{"error": "student_id parametresi zorunludur"}`, http.StatusBadRequest)
			return
		}
	}

	if err != nil {
		http.Error(w, `{"error": "Görüşmeler listelenemedi"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

type updateMeetingRequest struct {
	MeetingDate     string  `json:"meeting_date"`
	DurationMinutes int     `json:"duration_minutes"`
	Notes           string  `json:"notes"`
	MeetingURL      string  `json:"meeting_url"`
	NextMeetingDate *string `json:"next_meeting_date"`
	Status          string  `json:"status"`
}

func (h *TrackersHandler) UpdateMeeting(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	role := middleware.GetUserRole(r.Context())
	if role != domain.RoleCoach && role != domain.RoleSuperadmin {
		http.Error(w, `{"error": "Forbidden"}`, http.StatusForbidden)
		return
	}

	var req updateMeetingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	date, err := time.Parse(time.RFC3339, req.MeetingDate)
	if err != nil {
		date, _ = time.Parse("2006-01-02 15:04:05", req.MeetingDate)
	}

	var nextMeeting *time.Time
	if req.NextMeetingDate != nil && *req.NextMeetingDate != "" {
		t, err := time.Parse(time.RFC3339, *req.NextMeetingDate)
		if err == nil {
			nextMeeting = &t
		} else {
			t2, _ := time.Parse("2006-01-02 15:04:05", *req.NextMeetingDate)
			nextMeeting = &t2
		}
	}

	err = h.trackersService.UpdateMeeting(r.Context(), id, date, req.DurationMinutes, req.Notes, req.MeetingURL, nextMeeting, req.Status)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *TrackersHandler) DeleteMeeting(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	role := middleware.GetUserRole(r.Context())
	if role != domain.RoleCoach && role != domain.RoleSuperadmin {
		http.Error(w, `{"error": "Forbidden"}`, http.StatusForbidden)
		return
	}

	err := h.trackersService.DeleteMeeting(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "Silinemedi"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
