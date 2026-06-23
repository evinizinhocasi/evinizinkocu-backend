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

type ExamHandler struct {
	examService    *application.ExamService
	studentService *application.StudentService
}

func NewExamHandler(
	examService *application.ExamService,
	studentService *application.StudentService,
) *ExamHandler {
	return &ExamHandler{
		examService:    examService,
		studentService: studentService,
	}
}

func (h *ExamHandler) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Group(func(sub chi.Router) {
		sub.Use(authMiddleware)

		sub.Post("/api/v1/trial-exams", h.CreateTrialExam)
		sub.Get("/api/v1/trial-exams", h.ListTrialExams)
		sub.Put("/api/v1/trial-exams/{id}", h.UpdateTrialExam)
		sub.Delete("/api/v1/trial-exams/{id}", h.DeleteTrialExam)
	})
}

type trialExamCreateRequest struct {
	StudentID    string                            `json:"student_id"`
	ExamName     string                            `json:"exam_name"`
	ExamDate     string                            `json:"exam_date"` // YYYY-MM-DD
	ExamTypeID   string                            `json:"exam_type_id"`
	Score        *float64                          `json:"score"`
	Ranking      *int                              `json:"ranking"`
	CoachComment string                            `json:"coach_comment"`
	Results      []*application.SubjectResultInput `json:"results"`
}

func (h *ExamHandler) CreateTrialExam(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	var req trialExamCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.StudentID == "" || req.ExamName == "" || req.ExamDate == "" || req.ExamTypeID == "" || len(req.Results) == 0 {
		http.Error(w, `{"error": "Gerekli alanlar eksik"}`, http.StatusBadRequest)
		return
	}

	// Tenant check
	student, err := h.studentService.GetStudentByID(r.Context(), req.StudentID)
	if err != nil {
		http.Error(w, `{"error": "Öğrenci bulunamadı"}`, http.StatusBadRequest)
		return
	}

	if role == domain.RoleCoach && student.CoachID != userID {
		http.Error(w, `{"error": "Bu öğrenci size ait değil"}`, http.StatusForbidden)
		return
	} else if role == domain.RoleStudent && student.ID != userID {
		http.Error(w, `{"error": "Kendi profiliniz dışında veri ekleyemezsiniz"}`, http.StatusForbidden)
		return
	}

	date, err := time.Parse("2006-01-02", req.ExamDate)
	if err != nil {
		http.Error(w, `{"error": "Tarih formatı YYYY-MM-DD olmalı"}`, http.StatusBadRequest)
		return
	}

	res, err := h.examService.CreateTrialExam(r.Context(), req.StudentID, userID, req.ExamName, date, req.ExamTypeID, req.Score, req.Ranking, req.CoachComment, req.Results)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

func (h *ExamHandler) ListTrialExams(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	studentID := r.URL.Query().Get("student_id")
	examTypeID := r.URL.Query().Get("exam_type_id")
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	if studentID == "" {
		http.Error(w, `{"error": "student_id parametresi zorunludur"}`, http.StatusBadRequest)
		return
	}

	// Isolation check
	student, err := h.studentService.GetStudentByID(r.Context(), studentID)
	if err != nil {
		http.Error(w, `{"error": "Öğrenci bulunamadı"}`, http.StatusBadRequest)
		return
	}

	if role == domain.RoleCoach && student.CoachID != userID {
		http.Error(w, `{"error": "Bu öğrenci size ait değil"}`, http.StatusForbidden)
		return
	} else if role == domain.RoleStudent && student.ID != userID {
		http.Error(w, `{"error": "Kendi sınavlarınız dışındakilere erişemezsiniz"}`, http.StatusForbidden)
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

	exams, err := h.examService.ListExamsByStudent(r.Context(), studentID, examTypeID, startDate, endDate)
	if err != nil {
		http.Error(w, `{"error": "Sınavlar listelenemedi"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(exams)
}

type trialExamUpdateRequest struct {
	ExamName     string                            `json:"exam_name"`
	ExamDate     string                            `json:"exam_date"` // YYYY-MM-DD
	Score        *float64                          `json:"score"`
	Ranking      *int                              `json:"ranking"`
	CoachComment string                            `json:"coach_comment"`
	Results      []*application.SubjectResultInput `json:"results"`
}

func (h *ExamHandler) UpdateTrialExam(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	exam, err := h.examService.GetTrialExamByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "Sınav bulunamadı"}`, http.StatusNotFound)
		return
	}

	student, err := h.studentService.GetStudentByID(r.Context(), exam.StudentID)
	if err != nil {
		http.Error(w, `{"error": "Öğrenci bulunamadı"}`, http.StatusBadRequest)
		return
	}

	// Isolation Check
	if role == domain.RoleCoach && student.CoachID != userID {
		http.Error(w, `{"error": "Bu öğrenciye ait sınavı güncelleme yetkiniz yok"}`, http.StatusForbidden)
		return
	} else if role == domain.RoleStudent && student.ID != userID {
		http.Error(w, `{"error": "Kendi sınavınız dışındakileri güncelleyemezsiniz"}`, http.StatusForbidden)
		return
	}

	var req trialExamUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	date, err := time.Parse("2006-01-02", req.ExamDate)
	if err != nil {
		http.Error(w, `{"error": "Tarih formatı YYYY-MM-DD olmalı"}`, http.StatusBadRequest)
		return
	}

	err = h.examService.UpdateTrialExam(r.Context(), id, userID, req.ExamName, date, req.Score, req.Ranking, req.CoachComment, req.Results)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Sınav başarıyla güncellendi"}`))
}

func (h *ExamHandler) DeleteTrialExam(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	exam, err := h.examService.GetTrialExamByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "Sınav bulunamadı"}`, http.StatusNotFound)
		return
	}

	student, err := h.studentService.GetStudentByID(r.Context(), exam.StudentID)
	if err != nil {
		http.Error(w, `{"error": "Öğrenci bulunamadı"}`, http.StatusBadRequest)
		return
	}

	// Isolation Check
	if role == domain.RoleCoach && student.CoachID != userID {
		http.Error(w, `{"error": "Bu öğrenciye ait sınavı silme yetkiniz yok"}`, http.StatusForbidden)
		return
	} else if role == domain.RoleStudent && student.ID != userID {
		http.Error(w, `{"error": "Kendi sınavınız dışındakileri silemezsiniz"}`, http.StatusForbidden)
		return
	}

	err = h.examService.DeleteTrialExam(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "Sınav silinemedi"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Sınav başarıyla silindi"}`))
}
