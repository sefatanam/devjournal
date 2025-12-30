package rest

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"devjournal/internal/domain"
	"devjournal/internal/middleware"
	"devjournal/internal/service"
	"devjournal/pkg/httputil"

	"github.com/google/uuid"
)

// JournalHandler handles journal entry endpoints
type JournalHandler struct {
	journalService  *service.JournalService
	progressService *service.ProgressService
}

// NewJournalHandler creates a new journal handler
func NewJournalHandler(journalService *service.JournalService, progressService *service.ProgressService) *JournalHandler {
	return &JournalHandler{
		journalService:  journalService,
		progressService: progressService,
	}
}

// List handles GET /api/entries
func (h *JournalHandler) List(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httputil.Error(w, http.StatusUnauthorized, "invalid user ID")
		return
	}

	// Parse query parameters - support both page/pageSize and limit/offset
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	mood := r.URL.Query().Get("mood")
	search := r.URL.Query().Get("search")

	// Default values
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	// Convert to limit/offset for internal use
	limit := pageSize
	offset := (page - 1) * pageSize

	var entries []domain.JournalEntry
	var total int

	if search != "" {
		entries, err = h.journalService.Search(r.Context(), userID, search, limit, offset)
		total = len(entries)
	} else if mood != "" {
		entries, err = h.journalService.ListByMood(r.Context(), userID, mood, limit, offset)
		total = len(entries)
	} else {
		entries, total, err = h.journalService.List(r.Context(), userID, limit, offset)
	}

	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to list entries")
		return
	}

	// Calculate total pages
	totalPages := (total + pageSize - 1) / pageSize

	// Return format matching Angular's PaginatedResponse
	response := map[string]interface{}{
		"data":       entries,
		"total":      total,
		"page":       page,
		"pageSize":   pageSize,
		"totalPages": totalPages,
	}

	httputil.JSON(w, http.StatusOK, response)
}

// Get handles GET /api/entries/{id}
func (h *JournalHandler) Get(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httputil.Error(w, http.StatusUnauthorized, "invalid user ID")
		return
	}

	entryIDStr := r.PathValue("id")
	entryID, err := uuid.Parse(entryIDStr)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid entry ID")
		return
	}

	entry, err := h.journalService.GetByID(r.Context(), entryID, userID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to get entry")
		return
	}
	if entry == nil {
		httputil.Error(w, http.StatusNotFound, "entry not found")
		return
	}

	httputil.JSON(w, http.StatusOK, entry)
}

// Create handles POST /api/entries
func (h *JournalHandler) Create(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httputil.Error(w, http.StatusUnauthorized, "invalid user ID")
		return
	}

	var req domain.CreateJournalEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate request
	if req.Title == "" || req.Content == "" {
		httputil.Error(w, http.StatusBadRequest, "title and content are required")
		return
	}

	entry, err := h.journalService.Create(r.Context(), userID, &req)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to create entry")
		return
	}

	// Record journal entry for progress tracking
	if err := h.progressService.RecordJournalEntry(r.Context(), userID); err != nil {
		log.Printf("WARN: Failed to record journal entry for progress: %v", err)
		// Don't fail the request, progress tracking is secondary
	}

	httputil.JSON(w, http.StatusCreated, entry)
}

// Update handles PUT /api/entries/{id}
func (h *JournalHandler) Update(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httputil.Error(w, http.StatusUnauthorized, "invalid user ID")
		return
	}

	entryIDStr := r.PathValue("id")
	entryID, err := uuid.Parse(entryIDStr)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid entry ID")
		return
	}

	var req domain.UpdateJournalEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate request
	if req.Title == "" || req.Content == "" {
		httputil.Error(w, http.StatusBadRequest, "title and content are required")
		return
	}

	entry, err := h.journalService.Update(r.Context(), entryID, userID, &req)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to update entry")
		return
	}

	httputil.JSON(w, http.StatusOK, entry)
}

// Delete handles DELETE /api/entries/{id}
func (h *JournalHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httputil.Error(w, http.StatusUnauthorized, "invalid user ID")
		return
	}

	entryIDStr := r.PathValue("id")
	entryID, err := uuid.Parse(entryIDStr)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid entry ID")
		return
	}

	if err := h.journalService.Delete(r.Context(), entryID, userID); err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to delete entry")
		return
	}

	httputil.JSON(w, http.StatusOK, map[string]bool{"success": true})
}
