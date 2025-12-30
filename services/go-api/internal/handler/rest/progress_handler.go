package rest

import (
	"net/http"

	"devjournal/internal/middleware"
	"devjournal/internal/service"
	"devjournal/pkg/httputil"

	"github.com/google/uuid"
)

// ProgressHandler handles progress tracking endpoints
// @REVIEW - Phase 7: Progress Tracking REST handler
type ProgressHandler struct {
	progressService *service.ProgressService
}

// NewProgressHandler creates a new progress handler
func NewProgressHandler(progressService *service.ProgressService) *ProgressHandler {
	return &ProgressHandler{progressService: progressService}
}

// GetSummary handles GET /api/progress/summary
func (h *ProgressHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httputil.Error(w, http.StatusUnauthorized, "invalid user ID")
		return
	}

	summary, err := h.progressService.GetSummary(r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to get progress summary")
		return
	}

	httputil.JSON(w, http.StatusOK, summary)
}

// GetToday handles GET /api/progress/today
func (h *ProgressHandler) GetToday(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httputil.Error(w, http.StatusUnauthorized, "invalid user ID")
		return
	}

	progress, err := h.progressService.GetTodayProgress(r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to get today's progress")
		return
	}

	httputil.JSON(w, http.StatusOK, progress)
}

// GetWeekly handles GET /api/progress/weekly
func (h *ProgressHandler) GetWeekly(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httputil.Error(w, http.StatusUnauthorized, "invalid user ID")
		return
	}

	progressList, err := h.progressService.GetWeeklyProgress(r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to get weekly progress")
		return
	}

	httputil.JSON(w, http.StatusOK, map[string]interface{}{
		"progress": progressList,
		"period":   "weekly",
	})
}

// GetMonthly handles GET /api/progress/monthly
func (h *ProgressHandler) GetMonthly(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httputil.Error(w, http.StatusUnauthorized, "invalid user ID")
		return
	}

	progressList, err := h.progressService.GetMonthlyProgress(r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to get monthly progress")
		return
	}

	httputil.JSON(w, http.StatusOK, map[string]interface{}{
		"progress": progressList,
		"period":   "monthly",
	})
}

// GetStreak handles GET /api/progress/streak
func (h *ProgressHandler) GetStreak(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httputil.Error(w, http.StatusUnauthorized, "invalid user ID")
		return
	}

	streak, err := h.progressService.GetCurrentStreak(r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to get streak")
		return
	}

	httputil.JSON(w, http.StatusOK, map[string]interface{}{
		"currentStreak": streak,
	})
}
