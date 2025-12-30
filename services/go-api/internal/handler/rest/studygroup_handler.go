package rest

import (
	"encoding/json"
	"log"
	"net/http"

	"devjournal/internal/middleware"
	"devjournal/internal/service"
	"devjournal/pkg/httputil"

	"github.com/google/uuid"
)

// StudyGroupHandler handles study group HTTP requests
type StudyGroupHandler struct {
	groupService *service.StudyGroupService
}

// NewStudyGroupHandler creates a new study group handler
func NewStudyGroupHandler(groupService *service.StudyGroupService) *StudyGroupHandler {
	return &StudyGroupHandler{groupService: groupService}
}

// List returns all study groups for the current user
func (h *StudyGroupHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserUUID(r.Context())
	if userID == uuid.Nil {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	groups, err := h.groupService.ListByUser(r.Context(), userID)
	if err != nil {
		log.Printf("ERROR: ListByUser failed for user %s: %v", userID, err)
		httputil.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	httputil.JSON(w, http.StatusOK, groups)
}

// ListPublic returns all public study groups for discovery
func (h *StudyGroupHandler) ListPublic(w http.ResponseWriter, r *http.Request) {
	groups, total, err := h.groupService.ListPublic(r.Context(), 50, 0)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	httputil.JSON(w, http.StatusOK, map[string]interface{}{
		"data":  groups,
		"total": total,
	})
}

// Get returns a single study group by ID
func (h *StudyGroupHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid group ID")
		return
	}

	group, err := h.groupService.GetByID(r.Context(), id)
	if err != nil {
		httputil.Error(w, http.StatusNotFound, "group not found")
		return
	}

	// Get member count
	memberCount, _ := h.groupService.GetMemberCount(r.Context(), id)

	httputil.JSON(w, http.StatusOK, map[string]interface{}{
		"group":       group,
		"memberCount": memberCount,
	})
}

// Create creates a new study group
func (h *StudyGroupHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserUUID(r.Context())
	if userID == uuid.Nil {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req service.CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	group, err := h.groupService.Create(r.Context(), userID, &req)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	httputil.JSON(w, http.StatusCreated, group)
}

// Join adds the current user to a study group
func (h *StudyGroupHandler) Join(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserUUID(r.Context())
	if userID == uuid.Nil {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := r.PathValue("id")
	groupID, err := uuid.Parse(idStr)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid group ID")
		return
	}

	if err := h.groupService.Join(r.Context(), groupID, userID); err != nil {
		httputil.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	httputil.JSON(w, http.StatusOK, map[string]string{"message": "joined successfully"})
}

// Leave removes the current user from a study group
func (h *StudyGroupHandler) Leave(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserUUID(r.Context())
	if userID == uuid.Nil {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := r.PathValue("id")
	groupID, err := uuid.Parse(idStr)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid group ID")
		return
	}

	if err := h.groupService.Leave(r.Context(), groupID, userID); err != nil {
		httputil.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	httputil.JSON(w, http.StatusOK, map[string]string{"message": "left successfully"})
}

// GetMembers returns all members of a study group
func (h *StudyGroupHandler) GetMembers(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	groupID, err := uuid.Parse(idStr)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid group ID")
		return
	}

	members, err := h.groupService.GetMembers(r.Context(), groupID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	httputil.JSON(w, http.StatusOK, members)
}

// Delete removes a study group (only by owner)
func (h *StudyGroupHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserUUID(r.Context())
	if userID == uuid.Nil {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := r.PathValue("id")
	groupID, err := uuid.Parse(idStr)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid group ID")
		return
	}

	if err := h.groupService.Delete(r.Context(), groupID, userID); err != nil {
		httputil.Error(w, http.StatusForbidden, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
