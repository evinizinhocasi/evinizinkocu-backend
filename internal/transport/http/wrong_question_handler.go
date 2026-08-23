package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"evinizinkocu-backend/internal/application"
	"evinizinkocu-backend/internal/domain"
	"evinizinkocu-backend/internal/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type WrongQuestionHandler struct {
	service *application.WrongQuestionService
}

func NewWrongQuestionHandler(service *application.WrongQuestionService) *WrongQuestionHandler {
	return &WrongQuestionHandler{
		service: service,
	}
}

func (h *WrongQuestionHandler) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	// Public Media Proxy for R2 / local uploads
	r.Get("/media/{fileName}", h.MediaProxy)

	r.Group(func(sub chi.Router) {
		sub.Use(authMiddleware)

		sub.Post("/api/v1/wrong-questions", h.CreateWrongQuestion)
		sub.Get("/api/v1/wrong-questions", h.ListWrongQuestions)
		sub.Put("/api/v1/wrong-questions/{id}", h.UpdateWrongQuestion)
		sub.Put("/api/v1/wrong-questions/{id}/toggle", h.ToggleResolved)
		sub.Delete("/api/v1/wrong-questions/{id}", h.DeleteWrongQuestion)
	})
}

func (h *WrongQuestionHandler) CreateWrongQuestion(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())
	if userIDStr == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, `{"error":"invalid user id"}`, http.StatusUnauthorized)
		return
	}

	// Max 15 MB
	if err := r.ParseMultipartForm(15 << 20); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"invalid form data: %s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	studentIDStr := strings.TrimSpace(r.FormValue("studentId"))
	var studentID uuid.UUID

	if role == domain.RoleStudent {
		studentID = userID
	} else if studentIDStr != "" {
		studentID, err = uuid.Parse(studentIDStr)
		if err != nil {
			http.Error(w, `{"error":"invalid studentId format"}`, http.StatusBadRequest)
			return
		}
	} else {
		http.Error(w, `{"error":"studentId is required"}`, http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, `{"error":"image file is required"}`, http.StatusBadRequest)
		return
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, `{"error":"failed to read image file"}`, http.StatusInternalServerError)
		return
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg"
	}

	var subjectID *uuid.UUID
	subjectIDStr := strings.TrimSpace(r.FormValue("subjectId"))
	if subjectIDStr != "" {
		sid, err := uuid.Parse(subjectIDStr)
		if err == nil {
			subjectID = &sid
		}
	}

	title := strings.TrimSpace(r.FormValue("title"))
	note := strings.TrimSpace(r.FormValue("note"))

	question, err := h.service.Create(
		r.Context(),
		studentID,
		userID,
		string(role),
		subjectID,
		title,
		note,
		header.Filename,
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
		"message": "yanlış soru başarıyla yüklendi",
		"data":    question,
	})
}

func (h *WrongQuestionHandler) ListWrongQuestions(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())
	if userIDStr == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, `{"error":"invalid user id"}`, http.StatusUnauthorized)
		return
	}

	studentIDStr := strings.TrimSpace(r.URL.Query().Get("studentId"))
	var studentID uuid.UUID

	if role == domain.RoleStudent {
		studentID = userID
	} else if studentIDStr != "" {
		studentID, err = uuid.Parse(studentIDStr)
		if err != nil {
			http.Error(w, `{"error":"invalid studentId format"}`, http.StatusBadRequest)
			return
		}
	} else {
		http.Error(w, `{"error":"studentId is required"}`, http.StatusBadRequest)
		return
	}

	var subjectID *uuid.UUID
	subjectIDStr := strings.TrimSpace(r.URL.Query().Get("subjectId"))
	if subjectIDStr != "" {
		sid, err := uuid.Parse(subjectIDStr)
		if err == nil {
			subjectID = &sid
		}
	}

	list, err := h.service.ListByStudent(r.Context(), studentID, subjectID, userID, string(role))
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	if list == nil {
		list = []domain.WrongQuestion{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"wrongQuestions": list,
	})
}

func (h *WrongQuestionHandler) UpdateWrongQuestion(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())
	if userIDStr == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, `{"error":"invalid user id"}`, http.StatusUnauthorized)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid question id"}`, http.StatusBadRequest)
		return
	}

	var input domain.UpdateWrongQuestionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	input.ID = id

	if err := h.service.Update(r.Context(), input, userID, string(role)); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"message": "soru başarıyla güncellendi",
		"id":      id,
	})
}

func (h *WrongQuestionHandler) ToggleResolved(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())
	if userIDStr == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, `{"error":"invalid user id"}`, http.StatusUnauthorized)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid question id"}`, http.StatusBadRequest)
		return
	}

	var req struct {
		IsResolved bool `json:"isResolved"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if err := h.service.ToggleResolved(r.Context(), id, req.IsResolved, userID, string(role)); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"message":    "durum güncellendi",
		"id":         id,
		"isResolved": req.IsResolved,
	})
}

func (h *WrongQuestionHandler) DeleteWrongQuestion(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())
	if userIDStr == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, `{"error":"invalid user id"}`, http.StatusUnauthorized)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid question id"}`, http.StatusBadRequest)
		return
	}

	if err := h.service.Delete(r.Context(), id, userID, string(role)); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"message": "soru başarıyla silindi",
		"id":      id,
	})
}

func (h *WrongQuestionHandler) MediaProxy(w http.ResponseWriter, r *http.Request) {
	fileName := chi.URLParam(r, "fileName")
	fileName = filepath.Base(fileName)
	if fileName == "" || fileName == "." || fileName == "/" {
		http.Error(w, "invalid filename", http.StatusBadRequest)
		return
	}

	bytes, contentType, err := h.service.DownloadMedia(fileName)
	if err == nil && len(bytes) > 0 {
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Cache-Control", "public, max-age=86400")
		w.WriteHeader(http.StatusOK)
		w.Write(bytes)
		return
	}

	http.Error(w, fmt.Sprintf("Media not found: %v", err), http.StatusNotFound)
}
