package http

import (
	"encoding/json"
	"net/http"

	"evinizinkocu-backend/internal/application"
	"evinizinkocu-backend/internal/domain"
	"evinizinkocu-backend/internal/middleware"

	"github.com/go-chi/chi/v5"
)

type CatalogHandler struct {
	catalogService *application.CatalogService
}

func NewCatalogHandler(catalogService *application.CatalogService) *CatalogHandler {
	return &CatalogHandler{catalogService: catalogService}
}

func (h *CatalogHandler) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	// Public catalog fetch
	r.Get("/api/v1/catalog", h.GetFullCatalog)

	// Superadmin catalog management
	r.Group(func(sub chi.Router) {
		sub.Use(authMiddleware)
		sub.Use(middleware.RequireRole(domain.RoleSuperadmin))

		// Exams
		sub.Post("/api/v1/superadmin/catalog/exams", h.CreateExamType)
		sub.Put("/api/v1/superadmin/catalog/exams/{id}", h.UpdateExamType)
		sub.Delete("/api/v1/superadmin/catalog/exams/{id}", h.DeleteExamType)

		// Subjects
		sub.Post("/api/v1/superadmin/catalog/subjects", h.CreateSubject)
		sub.Put("/api/v1/superadmin/catalog/subjects/{id}", h.UpdateSubject)
		sub.Delete("/api/v1/superadmin/catalog/subjects/{id}", h.DeleteSubject)

		// Topics
		sub.Post("/api/v1/superadmin/catalog/topics", h.CreateTopic)
		sub.Put("/api/v1/superadmin/catalog/topics/{id}", h.UpdateTopic)
		sub.Delete("/api/v1/superadmin/catalog/topics/{id}", h.DeleteTopic)
	})
}

func (h *CatalogHandler) GetFullCatalog(w http.ResponseWriter, r *http.Request) {
	catalog, err := h.catalogService.GetFullCatalog(r.Context())
	if err != nil {
		http.Error(w, `{"error": "Katalog yüklenemedi"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(catalog)
}

type createExamRequest struct {
	Name    string `json:"name"`
	Divisor int    `json:"divisor"`
}

func (h *CatalogHandler) CreateExamType(w http.ResponseWriter, r *http.Request) {
	var req createExamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Divisor <= 0 {
		http.Error(w, `{"error": "Sınav adı ve pozitif bölen değeri zorunludur"}`, http.StatusBadRequest)
		return
	}

	res, err := h.catalogService.CreateExamType(r.Context(), req.Name, req.Divisor)
	if err != nil {
		http.Error(w, `{"error": "Sınav türü oluşturulamadı"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

func (h *CatalogHandler) UpdateExamType(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req createExamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Divisor <= 0 {
		http.Error(w, `{"error": "Geçersiz veriler"}`, http.StatusBadRequest)
		return
	}

	err := h.catalogService.UpdateExamType(r.Context(), id, req.Name, req.Divisor)
	if err != nil {
		http.Error(w, `{"error": "Sınav türü güncellenemedi"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *CatalogHandler) DeleteExamType(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := h.catalogService.DeleteExamType(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "Sınav türü silinemedi"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type createSubjectRequest struct {
	ExamTypeID    string `json:"exam_type_id"`
	Name          string `json:"name"`
	QuestionCount *int   `json:"question_count"`
}

func (h *CatalogHandler) CreateSubject(w http.ResponseWriter, r *http.Request) {
	var req createSubjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.ExamTypeID == "" || req.Name == "" {
		http.Error(w, `{"error": "Sınav türü ve ders adı zorunludur"}`, http.StatusBadRequest)
		return
	}

	res, err := h.catalogService.CreateSubject(r.Context(), req.ExamTypeID, req.Name, req.QuestionCount)
	if err != nil {
		http.Error(w, `{"error": "Ders oluşturulamadı"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

type updateSubjectRequest struct {
	Name          string `json:"name"`
	QuestionCount *int   `json:"question_count"`
}

func (h *CatalogHandler) UpdateSubject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req updateSubjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, `{"error": "Ders adı zorunludur"}`, http.StatusBadRequest)
		return
	}

	err := h.catalogService.UpdateSubject(r.Context(), id, req.Name, req.QuestionCount)
	if err != nil {
		http.Error(w, `{"error": "Ders güncellenemedi"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *CatalogHandler) DeleteSubject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := h.catalogService.DeleteSubject(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "Ders silinemedi"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type createTopicRequest struct {
	SubjectID string `json:"subject_id"`
	Name      string `json:"name"`
}

func (h *CatalogHandler) CreateTopic(w http.ResponseWriter, r *http.Request) {
	var req createTopicRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.SubjectID == "" || req.Name == "" {
		http.Error(w, `{"error": "Ders ve konu adı zorunludur"}`, http.StatusBadRequest)
		return
	}

	res, err := h.catalogService.CreateTopic(r.Context(), req.SubjectID, req.Name)
	if err != nil {
		http.Error(w, `{"error": "Konu oluşturulamadı"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

type updateTopicRequest struct {
	Name string `json:"name"`
}

func (h *CatalogHandler) UpdateTopic(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req updateTopicRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, `{"error": "Konu adı zorunludur"}`, http.StatusBadRequest)
		return
	}

	err := h.catalogService.UpdateTopic(r.Context(), id, req.Name)
	if err != nil {
		http.Error(w, `{"error": "Konu güncellenemedi"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *CatalogHandler) DeleteTopic(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := h.catalogService.DeleteTopic(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "Konu silinemedi"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
