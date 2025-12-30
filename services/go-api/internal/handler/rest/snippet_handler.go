package rest

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"devjournal/internal/domain"
	"devjournal/internal/middleware"
	"devjournal/internal/service"
	"devjournal/pkg/httputil"

	"github.com/google/uuid"
)

// SnippetHandler handles code snippet endpoints
type SnippetHandler struct {
	snippetService  *service.SnippetService
	progressService *service.ProgressService
}

// NewSnippetHandler creates a new snippet handler
func NewSnippetHandler(snippetService *service.SnippetService, progressService *service.ProgressService) *SnippetHandler {
	return &SnippetHandler{
		snippetService:  snippetService,
		progressService: progressService,
	}
}

// List handles GET /api/snippets
func (h *SnippetHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "invalid user ID")
		return
	}

	// Parse query parameters - support both page/pageSize and limit/offset
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	language := r.URL.Query().Get("language")
	tagsParam := r.URL.Query().Get("tags")
	search := r.URL.Query().Get("search")

	// Default values
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	// Convert to limit/offset for internal use
	limit := int64(pageSize)
	offset := int64((page - 1) * pageSize)

	var snippets []domain.Snippet
	var total int64
	var err error

	if search != "" {
		snippets, err = h.snippetService.Search(r.Context(), userID, search, limit, offset)
		total = int64(len(snippets))
	} else if tagsParam != "" {
		tags := strings.Split(tagsParam, ",")
		snippets, err = h.snippetService.ListByTags(r.Context(), userID, tags, limit, offset)
		total = int64(len(snippets))
	} else if language != "" {
		snippets, err = h.snippetService.ListByLanguage(r.Context(), userID, language, limit, offset)
		total = int64(len(snippets))
	} else {
		snippets, total, err = h.snippetService.List(r.Context(), userID, limit, offset)
	}

	if err != nil {
		log.Printf("ERROR: Failed to list snippets for user %s: %v", userID, err)
		httputil.Error(w, http.StatusInternalServerError, "failed to list snippets")
		return
	}

	// Calculate total pages
	totalPages := (int(total) + pageSize - 1) / pageSize

	// Return format matching Angular's PaginatedResponse
	response := map[string]interface{}{
		"data":       snippets,
		"total":      total,
		"page":       page,
		"pageSize":   pageSize,
		"totalPages": totalPages,
	}

	httputil.JSON(w, http.StatusOK, response)
}

// Get handles GET /api/snippets/{id}
func (h *SnippetHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "invalid user ID")
		return
	}

	snippetID := r.PathValue("id")
	if snippetID == "" {
		httputil.Error(w, http.StatusBadRequest, "invalid snippet ID")
		return
	}

	snippet, err := h.snippetService.GetByID(r.Context(), snippetID, userID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to get snippet")
		return
	}
	if snippet == nil {
		httputil.Error(w, http.StatusNotFound, "snippet not found")
		return
	}

	httputil.JSON(w, http.StatusOK, snippet)
}

// Create handles POST /api/snippets
func (h *SnippetHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "invalid user ID")
		return
	}

	var req domain.CreateSnippetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate request
	if req.Title == "" || req.Code == "" || req.Language == "" {
		httputil.Error(w, http.StatusBadRequest, "title, code, and language are required")
		return
	}

	snippet, err := h.snippetService.Create(r.Context(), userID, &req)
	if err != nil {
		log.Printf("ERROR: Failed to create snippet for user %s: %v", userID, err)
		httputil.Error(w, http.StatusInternalServerError, "failed to create snippet")
		return
	}

	// Record snippet creation for progress tracking
	if userUUID, err := uuid.Parse(userID); err == nil {
		if err := h.progressService.RecordSnippet(r.Context(), userUUID); err != nil {
			log.Printf("WARN: Failed to record snippet for progress: %v", err)
			// Don't fail the request, progress tracking is secondary
		}
	}

	httputil.JSON(w, http.StatusCreated, snippet)
}

// Update handles PUT /api/snippets/{id}
func (h *SnippetHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "invalid user ID")
		return
	}

	snippetID := r.PathValue("id")
	if snippetID == "" {
		httputil.Error(w, http.StatusBadRequest, "invalid snippet ID")
		return
	}

	var req domain.UpdateSnippetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate request
	if req.Title == "" || req.Code == "" || req.Language == "" {
		httputil.Error(w, http.StatusBadRequest, "title, code, and language are required")
		return
	}

	snippet, err := h.snippetService.Update(r.Context(), snippetID, userID, &req)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to update snippet")
		return
	}

	httputil.JSON(w, http.StatusOK, snippet)
}

// Delete handles DELETE /api/snippets/{id}
func (h *SnippetHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "invalid user ID")
		return
	}

	snippetID := r.PathValue("id")
	if snippetID == "" {
		httputil.Error(w, http.StatusBadRequest, "invalid snippet ID")
		return
	}

	if err := h.snippetService.Delete(r.Context(), snippetID, userID); err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to delete snippet")
		return
	}

	httputil.JSON(w, http.StatusOK, map[string]bool{"success": true})
}
