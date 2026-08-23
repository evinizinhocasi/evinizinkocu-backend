package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"evinizinkocu-backend/internal/application"
	"evinizinkocu-backend/internal/domain"
	"evinizinkocu-backend/internal/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type TargetSchoolHandler struct {
	service *application.TargetSchoolService
}

func NewTargetSchoolHandler(service *application.TargetSchoolService) *TargetSchoolHandler {
	return &TargetSchoolHandler{
		service: service,
	}
}

func (h *TargetSchoolHandler) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Group(func(sub chi.Router) {
		sub.Use(authMiddleware)

		// Public/Authenticated access for listing target schools
		sub.Get("/api/v1/target-schools", h.ListTargetSchools)
		sub.Get("/api/v1/target-schools/{id}", h.GetTargetSchoolByID)

		// Admin management
		sub.Post("/api/v1/target-schools", h.CreateTargetSchool)
		sub.Put("/api/v1/target-schools/{id}", h.UpdateTargetSchool)
		sub.Delete("/api/v1/target-schools/{id}", h.DeleteTargetSchool)

		// Student Target School Assignment (Coach/Student/Admin)
		sub.Put("/api/v1/students/{id}/target-school", h.UpdateStudentTargetSchool)
	})
}

func (h *TargetSchoolHandler) ListTargetSchools(w http.ResponseWriter, r *http.Request) {
	schoolType := strings.TrimSpace(r.URL.Query().Get("type"))
	search := strings.TrimSpace(r.URL.Query().Get("search"))

	schools, err := h.service.List(r.Context(), schoolType, search)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"targetSchools": schools,
	})
}

func (h *TargetSchoolHandler) GetTargetSchoolByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}

	school, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
	if school == nil {
		http.Error(w, `{"error":"hedef okul bulunamadı"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"targetSchool": school,
	})
}

func (h *TargetSchoolHandler) CreateTargetSchool(w http.ResponseWriter, r *http.Request) {
	role := middleware.GetUserRole(r.Context())
	if role != domain.RoleSuperadmin {
		http.Error(w, `{"error":"bu işlem için yetkiniz yok"}`, http.StatusForbidden)
		return
	}

	if err := r.ParseMultipartForm(15 << 20); err != nil {
		// Try fallback if json was sent
		h.handleCreateJSON(w, r)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		http.Error(w, `{"error":"name is required"}`, http.StatusBadRequest)
		return
	}

	typeStr := strings.TrimSpace(r.FormValue("type"))
	if typeStr == "" {
		typeStr = "high_school"
	}
	schoolType := domain.TargetSchoolType(typeStr)

	city := strings.TrimSpace(r.FormValue("city"))
	department := strings.TrimSpace(r.FormValue("department"))
	minScore, _ := strconv.ParseFloat(strings.ReplaceAll(r.FormValue("minScore"), ",", "."), 64)
	percentile, _ := strconv.ParseFloat(strings.ReplaceAll(r.FormValue("percentile"), ",", "."), 64)
	ranking, _ := strconv.Atoi(r.FormValue("ranking"))

	var fileBytes []byte
	var fileName, contentType string

	file, header, err := r.FormFile("photo")
	if err == nil {
		defer file.Close()
		fileBytes, _ = io.ReadAll(file)
		fileName = header.Filename
		contentType = header.Header.Get("Content-Type")
	}

	school, err := h.service.Create(
		r.Context(),
		name,
		schoolType,
		city,
		minScore,
		percentile,
		ranking,
		department,
		fileName,
		fileBytes,
		contentType,
	)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"message": "hedef okul başarıyla oluşturuldu",
		"data":    school,
	})
}

func (h *TargetSchoolHandler) handleCreateJSON(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name       string  `json:"name"`
		Type       string  `json:"type"`
		City       string  `json:"city"`
		MinScore   float64 `json:"minScore"`
		Percentile float64 `json:"percentile"`
		Ranking    int     `json:"ranking"`
		Department string  `json:"department"`
		PhotoURL   string  `json:"photoUrl"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request format"}`, http.StatusBadRequest)
		return
	}

	typeStr := req.Type
	if typeStr == "" {
		typeStr = "high_school"
	}

	school, err := h.service.Create(
		r.Context(),
		req.Name,
		domain.TargetSchoolType(typeStr),
		req.City,
		req.MinScore,
		req.Percentile,
		req.Ranking,
		req.Department,
		"",
		nil,
		"",
	)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"message": "hedef okul başarıyla oluşturuldu",
		"data":    school,
	})
}

func (h *TargetSchoolHandler) UpdateTargetSchool(w http.ResponseWriter, r *http.Request) {
	role := middleware.GetUserRole(r.Context())
	if role != domain.RoleSuperadmin {
		http.Error(w, `{"error":"bu işlem için yetkiniz yok"}`, http.StatusForbidden)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}

	if err := r.ParseMultipartForm(15 << 20); err != nil {
		// Handle JSON update
		var req struct {
			Name       string  `json:"name"`
			Type       string  `json:"type"`
			City       string  `json:"city"`
			MinScore   float64 `json:"minScore"`
			Percentile float64 `json:"percentile"`
			Ranking    int     `json:"ranking"`
			Department string  `json:"department"`
		}
		if decErr := json.NewDecoder(r.Body).Decode(&req); decErr != nil {
			http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}
		typeStr := req.Type
		if typeStr == "" {
			typeStr = "high_school"
		}
		school, updateErr := h.service.Update(
			r.Context(),
			id,
			req.Name,
			domain.TargetSchoolType(typeStr),
			req.City,
			req.MinScore,
			req.Percentile,
			req.Ranking,
			req.Department,
			"",
			nil,
			"",
		)
		if updateErr != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%s"}`, updateErr.Error()), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"message": "hedef okul güncellendi",
			"data":    school,
		})
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	typeStr := strings.TrimSpace(r.FormValue("type"))
	if typeStr == "" {
		typeStr = "high_school"
	}
	schoolType := domain.TargetSchoolType(typeStr)

	city := strings.TrimSpace(r.FormValue("city"))
	department := strings.TrimSpace(r.FormValue("department"))
	minScore, _ := strconv.ParseFloat(strings.ReplaceAll(r.FormValue("minScore"), ",", "."), 64)
	percentile, _ := strconv.ParseFloat(strings.ReplaceAll(r.FormValue("percentile"), ",", "."), 64)
	ranking, _ := strconv.Atoi(r.FormValue("ranking"))

	var fileBytes []byte
	var fileName, contentType string

	file, header, err := r.FormFile("photo")
	if err == nil {
		defer file.Close()
		fileBytes, _ = io.ReadAll(file)
		fileName = header.Filename
		contentType = header.Header.Get("Content-Type")
	}

	school, err := h.service.Update(
		r.Context(),
		id,
		name,
		schoolType,
		city,
		minScore,
		percentile,
		ranking,
		department,
		fileName,
		fileBytes,
		contentType,
	)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"message": "hedef okul güncellendi",
		"data":    school,
	})
}

func (h *TargetSchoolHandler) DeleteTargetSchool(w http.ResponseWriter, r *http.Request) {
	role := middleware.GetUserRole(r.Context())
	if role != domain.RoleSuperadmin {
		http.Error(w, `{"error":"bu işlem için yetkiniz yok"}`, http.StatusForbidden)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"message": "hedef okul silindi",
		"id":      id,
	})
}

func (h *TargetSchoolHandler) UpdateStudentTargetSchool(w http.ResponseWriter, r *http.Request) {
	requestingUserID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())
	studentID := chi.URLParam(r, "id")

	// Allow multipart (for custom photo upload) or json
	var targetSchoolID *string
	var targetSchoolName, targetSchoolPhoto, targetSchoolType string
	var fileBytes []byte
	var fileName, contentType string

	if err := r.ParseMultipartForm(15 << 20); err == nil {
		idVal := strings.TrimSpace(r.FormValue("targetSchoolId"))
		if idVal != "" {
			targetSchoolID = &idVal
		}
		targetSchoolName = strings.TrimSpace(r.FormValue("targetSchoolName"))
		targetSchoolPhoto = strings.TrimSpace(r.FormValue("targetSchoolPhoto"))
		targetSchoolType = strings.TrimSpace(r.FormValue("targetSchoolType"))

		file, header, err := r.FormFile("photo")
		if err == nil {
			defer file.Close()
			fileBytes, _ = io.ReadAll(file)
			fileName = header.Filename
			contentType = header.Header.Get("Content-Type")
		}
	} else {
		var req struct {
			TargetSchoolID    *string `json:"targetSchoolId"`
			TargetSchoolName  string  `json:"targetSchoolName"`
			TargetSchoolPhoto string  `json:"targetSchoolPhoto"`
			TargetSchoolType  string  `json:"targetSchoolType"`
		}
		if decErr := json.NewDecoder(r.Body).Decode(&req); decErr == nil {
			targetSchoolID = req.TargetSchoolID
			targetSchoolName = req.TargetSchoolName
			targetSchoolPhoto = req.TargetSchoolPhoto
			targetSchoolType = req.TargetSchoolType
		}
	}

	err := h.service.UpdateStudentTargetSchool(
		r.Context(),
		studentID,
		requestingUserID,
		role,
		targetSchoolID,
		targetSchoolName,
		targetSchoolPhoto,
		targetSchoolType,
		fileName,
		fileBytes,
		contentType,
	)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"message":           "öğrenci hedef okulu başarıyla güncellendi",
		"targetSchoolName":  targetSchoolName,
		"targetSchoolPhoto": targetSchoolPhoto,
		"targetSchoolType":  targetSchoolType,
	})
}
