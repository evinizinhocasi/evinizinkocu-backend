package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"evinizinkocu-backend/internal/application"
	"evinizinkocu-backend/internal/domain"
	"evinizinkocu-backend/internal/middleware"

	"github.com/go-chi/chi/v5"
)

type StudentHandler struct {
	studentService *application.StudentService
}

func NewStudentHandler(studentService *application.StudentService) *StudentHandler {
	return &StudentHandler{studentService: studentService}
}

func (h *StudentHandler) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Group(func(sub chi.Router) {
		sub.Use(authMiddleware)

		// Create Student
		sub.Post("/api/v1/students", h.CreateStudent)

		// List Students
		sub.Get("/api/v1/students", h.ListStudents)

		// Get Single Student
		sub.Get("/api/v1/students/{id}", h.GetStudent)

		// Update Student
		sub.Put("/api/v1/students/{id}", h.UpdateStudent)

		// Transfer Student (Superadmin only)
		sub.With(middleware.RequireRole(domain.RoleSuperadmin)).Post("/api/v1/students/{id}/transfer", h.TransferStudent)

		// Resend Credentials (Superadmin & Coach)
		sub.Post("/api/v1/students/{id}/resend-credentials", h.ResendCredentials)
	})
}

type createStudentRequest struct {
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	ClassLevel  string `json:"class_level"`
	StudyTrack  string `json:"study_track"`
	ExamTypeID  string `json:"exam_type_id"`
	CoachID     string `json:"coach_id"` // Optional for coach, mandatory for superadmin
}

func (h *StudentHandler) CreateStudent(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	var req createStudentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.FirstName == "" || req.LastName == "" || req.Phone == "" || req.Email == "" || req.ClassLevel == "" || req.StudyTrack == "" || req.ExamTypeID == "" {
		http.Error(w, `{"error": "Tüm alanlar zorunludur"}`, http.StatusBadRequest)
		return
	}

	// Determine Coach ID
	targetCoachID := req.CoachID
	if role == domain.RoleCoach {
		// A coach can only create students for themselves
		targetCoachID = userID
	} else if role == domain.RoleSuperadmin {
		if targetCoachID == "" {
			http.Error(w, `{"error": "Süper admin için koç seçimi zorunludur"}`, http.StatusBadRequest)
			return
		}
	} else {
		http.Error(w, `{"error": "Yetkiniz yok"}`, http.StatusForbidden)
		return
	}

	res, err := h.studentService.CreateStudent(r.Context(), req.FirstName, req.LastName, req.Phone, req.Email, req.ClassLevel, req.StudyTrack, req.ExamTypeID, targetCoachID)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

func (h *StudentHandler) ListStudents(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	var students []*domain.Student
	var err error

	if role == domain.RoleCoach {
		students, err = h.studentService.ListStudentsByCoach(r.Context(), userID, true)
	} else if role == domain.RoleSuperadmin {
		students, err = h.studentService.ListAllStudents(r.Context())
	} else {
		http.Error(w, `{"error": "Forbidden"}`, http.StatusForbidden)
		return
	}

	if err != nil {
		http.Error(w, `{"error": "Öğrenciler listelenemedi"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(students)
}

func (h *StudentHandler) GetStudent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	student, err := h.studentService.GetStudentByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "Öğrenci bulunamadı"}`, http.StatusNotFound)
		return
	}

	// Enforce isolation
	if role == domain.RoleCoach && student.CoachID != userID {
		http.Error(w, `{"error": "Bu öğrenci size ait değil"}`, http.StatusForbidden)
		return
	} else if role == domain.RoleStudent && student.ID != userID {
		http.Error(w, `{"error": "Kendi profiliniz dışındaki profillere erişemezsiniz"}`, http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(student)
}

type updateStudentRequest struct {
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Phone      string `json:"phone"`
	ClassLevel string `json:"class_level"`
	StudyTrack string `json:"study_track"`
	ExamTypeID string `json:"exam_type_id"`
	IsArchived bool   `json:"is_archived"`
}

func (h *StudentHandler) UpdateStudent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	student, err := h.studentService.GetStudentByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "Öğrenci bulunamadı"}`, http.StatusNotFound)
		return
	}

	var req updateStudentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Enforce fields permissions
	if role == domain.RoleStudent {
		if student.ID != userID {
			http.Error(w, `{"error": "Yetkiniz yok"}`, http.StatusForbidden)
			return
		}
		// Students can only update personal fields, not class, track, exams or archiving
		err = h.studentService.UpdateStudentProfile(r.Context(), id, req.FirstName, req.LastName, req.Phone, student.ClassLevel, student.StudyTrack, student.ExamTypeID, student.IsArchived)
	} else if role == domain.RoleCoach {
		if student.CoachID != userID {
			http.Error(w, `{"error": "Bu öğrenci size ait değil"}`, http.StatusForbidden)
			return
		}
		err = h.studentService.UpdateStudentProfile(r.Context(), id, req.FirstName, req.LastName, req.Phone, req.ClassLevel, req.StudyTrack, req.ExamTypeID, req.IsArchived)
	} else if role == domain.RoleSuperadmin {
		err = h.studentService.UpdateStudentProfile(r.Context(), id, req.FirstName, req.LastName, req.Phone, req.ClassLevel, req.StudyTrack, req.ExamTypeID, req.IsArchived)
	} else {
		http.Error(w, `{"error": "Yetkiniz yok"}`, http.StatusForbidden)
		return
	}

	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Öğrenci bilgileri başarıyla güncellendi"}`))
}

type transferRequest struct {
	NewCoachID string `json:"new_coach_id"`
}

func (h *StudentHandler) TransferStudent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req transferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.NewCoachID == "" {
		http.Error(w, `{"error": "Yeni koç seçimi zorunludur"}`, http.StatusBadRequest)
		return
	}

	err := h.studentService.TransferStudent(r.Context(), id, req.NewCoachID)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Öğrenci başarıyla yeni koçuna transfer edildi"}`))
}

func (h *StudentHandler) ResendCredentials(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	err := h.studentService.ResendCredentials(r.Context(), id, userID, role)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Öğrenci giriş bilgileri tekrar gönderildi"}`))
}

