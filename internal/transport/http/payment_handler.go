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

type PaymentHandler struct {
	paymentService *application.PaymentService
	studentService *application.StudentService
}

func NewPaymentHandler(
	paymentService *application.PaymentService,
	studentService *application.StudentService,
) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
		studentService: studentService,
	}
}

func (h *PaymentHandler) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Group(func(sub chi.Router) {
		sub.Use(authMiddleware)

		// Coach's own payment history
		sub.With(middleware.RequireRole(domain.RoleCoach)).Get("/api/v1/coach/payments", h.GetCoachPaymentsSelf)

		// Superadmin coach payments management
		sub.Group(func(admin chi.Router) {
			admin.Use(middleware.RequireRole(domain.RoleSuperadmin))

			admin.Post("/api/v1/superadmin/payments/coaches", h.CreateCoachPayment)
			admin.Get("/api/v1/superadmin/payments/coaches", h.ListCoachPayments)
			admin.Put("/api/v1/superadmin/payments/coaches/{id}", h.UpdateCoachPayment)
			admin.Delete("/api/v1/superadmin/payments/coaches/{id}", h.DeleteCoachPayment)
		})

		// Student payments (accessible by students, coaches, superadmins)
		sub.Post("/api/v1/payments/students", h.CreateStudentPayment)
		sub.Get("/api/v1/payments/students", h.ListStudentPayments)
		sub.Put("/api/v1/payments/students/{id}", h.UpdateStudentPayment)
		sub.Delete("/api/v1/payments/students/{id}", h.DeleteStudentPayment)
	})
}

// Coach payments CRUD

type createCoachPaymentRequest struct {
	CoachID     string  `json:"coach_id"`
	Amount      float64 `json:"amount"`
	PaymentDate string  `json:"payment_date"` // YYYY-MM-DD
	Description string  `json:"description"`
	Status      string  `json:"status"`
}

func (h *PaymentHandler) CreateCoachPayment(w http.ResponseWriter, r *http.Request) {
	var req createCoachPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.CoachID == "" || req.Amount < 0 || req.PaymentDate == "" || req.Status == "" {
		http.Error(w, `{"error": "Gerekli alanlar eksik veya hatalı"}`, http.StatusBadRequest)
		return
	}

	date, err := time.Parse("2006-01-02", req.PaymentDate)
	if err != nil {
		http.Error(w, `{"error": "Tarih formatı YYYY-MM-DD olmalı"}`, http.StatusBadRequest)
		return
	}

	res, err := h.paymentService.CreateCoachPayment(r.Context(), req.CoachID, req.Amount, date, req.Description, req.Status)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

func (h *PaymentHandler) ListCoachPayments(w http.ResponseWriter, r *http.Request) {
	coachID := r.URL.Query().Get("coach_id") // Optional filter
	list, err := h.paymentService.ListCoachPayments(r.Context(), coachID)
	if err != nil {
		http.Error(w, `{"error": "Ödeme kayıtları alınamadı"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func (h *PaymentHandler) GetCoachPaymentsSelf(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	list, err := h.paymentService.ListCoachPayments(r.Context(), userID)
	if err != nil {
		http.Error(w, `{"error": "Ödeme geçmişiniz alınamadı"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func (h *PaymentHandler) UpdateCoachPayment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req createCoachPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	date, err := time.Parse("2006-01-02", req.PaymentDate)
	if err != nil {
		http.Error(w, `{"error": "Tarih formatı YYYY-MM-DD olmalı"}`, http.StatusBadRequest)
		return
	}

	err = h.paymentService.UpdateCoachPayment(r.Context(), id, req.Amount, date, req.Description, req.Status)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *PaymentHandler) DeleteCoachPayment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := h.paymentService.DeleteCoachPayment(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "Silinemedi"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Student Payments CRUD

type createStudentPaymentRequest struct {
	StudentID   string  `json:"student_id"`
	Amount      float64 `json:"amount"`
	PaymentDate string  `json:"payment_date"` // YYYY-MM-DD
	Description string  `json:"description"`
	Status      string  `json:"status"`
}

func (h *PaymentHandler) CreateStudentPayment(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	if role != domain.RoleCoach && role != domain.RoleSuperadmin {
		http.Error(w, `{"error": "Yalnızca koç veya süper yönetici öğrenci ödemesi ekleyebilir"}`, http.StatusForbidden)
		return
	}

	var req createStudentPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.StudentID == "" || req.Amount < 0 || req.PaymentDate == "" || req.Status == "" {
		http.Error(w, `{"error": "Eksik parametre"}`, http.StatusBadRequest)
		return
	}

	// Verify tenant
	student, err := h.studentService.GetStudentByID(r.Context(), req.StudentID)
	if err != nil {
		http.Error(w, `{"error": "Öğrenci bulunamadı"}`, http.StatusBadRequest)
		return
	}

	coachID := student.CoachID
	if role == domain.RoleCoach && student.CoachID != userID {
		http.Error(w, `{"error": "Bu öğrenci size ait değil"}`, http.StatusForbidden)
		return
	}

	date, err := time.Parse("2006-01-02", req.PaymentDate)
	if err != nil {
		http.Error(w, `{"error": "Geçersiz tarih"}`, http.StatusBadRequest)
		return
	}

	res, err := h.paymentService.CreateStudentPayment(r.Context(), req.StudentID, coachID, req.Amount, date, req.Description, req.Status)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

func (h *PaymentHandler) ListStudentPayments(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	studentID := r.URL.Query().Get("student_id")

	var list []*domain.StudentPayment
	var err error

	if role == domain.RoleStudent {
		list, err = h.paymentService.ListStudentPayments(r.Context(), userID)
	} else if role == domain.RoleCoach {
		if studentID != "" {
			student, err := h.studentService.GetStudentByID(r.Context(), studentID)
			if err != nil || student.CoachID != userID {
				http.Error(w, `{"error": "Erişim yetkiniz yok"}`, http.StatusForbidden)
				return
			}
			list, err = h.paymentService.ListStudentPayments(r.Context(), studentID)
		} else {
			list, err = h.paymentService.ListStudentPaymentsByCoach(r.Context(), userID)
		}
	} else if role == domain.RoleSuperadmin {
		if studentID != "" {
			list, err = h.paymentService.ListStudentPayments(r.Context(), studentID)
		} else {
			coachID := r.URL.Query().Get("coach_id")
			list, err = h.paymentService.ListStudentPaymentsByCoach(r.Context(), coachID)
		}
	} else {
		http.Error(w, `{"error": "Forbidden"}`, http.StatusForbidden)
		return
	}

	if err != nil {
		http.Error(w, `{"error": "Ödeme kayıtları listelenemedi"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func (h *PaymentHandler) UpdateStudentPayment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	if role != domain.RoleCoach && role != domain.RoleSuperadmin {
		http.Error(w, `{"error": "Forbidden"}`, http.StatusForbidden)
		return
	}

	p, err := h.paymentService.GetStudentPaymentByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "Ödeme kaydı bulunamadı"}`, http.StatusNotFound)
		return
	}

	if role == domain.RoleCoach && p.CoachID != userID {
		http.Error(w, `{"error": "Bu ödeme kaydını güncelleme yetkiniz yok"}`, http.StatusForbidden)
		return
	}

	var req createStudentPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	date, err := time.Parse("2006-01-02", req.PaymentDate)
	if err != nil {
		http.Error(w, `{"error": "Geçersiz tarih"}`, http.StatusBadRequest)
		return
	}

	err = h.paymentService.UpdateStudentPayment(r.Context(), id, req.Amount, date, req.Description, req.Status)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *PaymentHandler) DeleteStudentPayment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	if role != domain.RoleCoach && role != domain.RoleSuperadmin {
		http.Error(w, `{"error": "Forbidden"}`, http.StatusForbidden)
		return
	}

	p, err := h.paymentService.GetStudentPaymentByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "Ödeme kaydı bulunamadı"}`, http.StatusNotFound)
		return
	}

	if role == domain.RoleCoach && p.CoachID != userID {
		http.Error(w, `{"error": "Bu ödeme kaydını silme yetkiniz yok"}`, http.StatusForbidden)
		return
	}

	err = h.paymentService.DeleteStudentPayment(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "Silinemedi"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
