package http

import (
	"encoding/json"
	"net/http"

	"evinizinkocu-backend/internal/application"
	"evinizinkocu-backend/internal/domain"
	"evinizinkocu-backend/internal/middleware"

	"github.com/go-chi/chi/v5"
)

type CMSHandler struct {
	cmsService *application.CMSService
}

func NewCMSHandler(cmsService *application.CMSService) *CMSHandler {
	return &CMSHandler{cmsService: cmsService}
}

func (h *CMSHandler) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	// Public routes
	r.Get("/api/v1/public/settings", h.GetPublicSettings)
	r.Get("/api/v1/public/legal/{slug}", h.GetLegalDocument)

	// Superadmin protected routes
	r.Group(func(sub chi.Router) {
		sub.Use(authMiddleware)
		sub.Use(middleware.RequireRole(domain.RoleSuperadmin))

		sub.Post("/api/v1/superadmin/cms/hero", h.SaveHeroSettings)
		sub.Post("/api/v1/superadmin/cms/faqs", h.CreateFAQ)
		sub.Put("/api/v1/superadmin/cms/faqs/{id}", h.UpdateFAQ)
		sub.Delete("/api/v1/superadmin/cms/faqs/{id}", h.DeleteFAQ)
		sub.Post("/api/v1/superadmin/cms/legal", h.SaveLegalDocument)
	})
}

type publicSettingsResponse struct {
	Hero *domain.HeroSettings `json:"hero"`
	FAQs []*domain.FAQ        `json:"faqs"`
}

func (h *CMSHandler) GetPublicSettings(w http.ResponseWriter, r *http.Request) {
	hero, err := h.cmsService.GetHeroSettings(r.Context())
	if err != nil {
		hero = &domain.HeroSettings{}
	}

	faqs, err := h.cmsService.ListFAQs(r.Context())
	if err != nil {
		faqs = []*domain.FAQ{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(publicSettingsResponse{
		Hero: hero,
		FAQs: faqs,
	})
}

func (h *CMSHandler) GetLegalDocument(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	doc, err := h.cmsService.GetLegalDocument(r.Context(), slug)
	if err != nil {
		http.Error(w, `{"error": "Belge bulunamadı"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(doc)
}

func (h *CMSHandler) SaveHeroSettings(w http.ResponseWriter, r *http.Request) {
	var req domain.HeroSettings
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	err := h.cmsService.SaveHeroSettings(r.Context(), &req)
	if err != nil {
		http.Error(w, `{"error": "Ayarlar kaydedilemedi"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type createFAQRequest struct {
	Question     string `json:"question"`
	Answer       string `json:"answer"`
	DisplayOrder int    `json:"display_order"`
}

func (h *CMSHandler) CreateFAQ(w http.ResponseWriter, r *http.Request) {
	var req createFAQRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Question == "" || req.Answer == "" {
		http.Error(w, `{"error": "Soru ve cevap zorunludur"}`, http.StatusBadRequest)
		return
	}

	res, err := h.cmsService.CreateFAQ(r.Context(), req.Question, req.Answer, req.DisplayOrder)
	if err != nil {
		http.Error(w, `{"error": "Soru oluşturulamadı"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

func (h *CMSHandler) UpdateFAQ(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req createFAQRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Question == "" || req.Answer == "" {
		http.Error(w, `{"error": "Soru ve cevap zorunludur"}`, http.StatusBadRequest)
		return
	}

	err := h.cmsService.UpdateFAQ(r.Context(), id, req.Question, req.Answer, req.DisplayOrder)
	if err != nil {
		http.Error(w, `{"error": "Soru güncellenemedi"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *CMSHandler) DeleteFAQ(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := h.cmsService.DeleteFAQ(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "Soru silinemedi"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type saveDocRequest struct {
	Slug    string `json:"slug"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (h *CMSHandler) SaveLegalDocument(w http.ResponseWriter, r *http.Request) {
	var req saveDocRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Slug == "" || req.Title == "" || req.Content == "" {
		http.Error(w, `{"error": "Eksik parametre"}`, http.StatusBadRequest)
		return
	}

	err := h.cmsService.SaveLegalDocument(r.Context(), req.Slug, req.Title, req.Content)
	if err != nil {
		http.Error(w, `{"error": "Doküman kaydedilemedi"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
