package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"edupulse/internal/middleware"
	"edupulse/internal/service"
)

type AssignmentHandler struct {
	svc    *service.AssignmentService
	manage *service.HomeworkManageService
}

func NewAssignmentHandler(svc *service.AssignmentService, manage *service.HomeworkManageService) *AssignmentHandler {
	return &AssignmentHandler{svc: svc, manage: manage}
}

type createAssignmentReq struct {
	SessionID   int64  `json:"session_id" example:"1"`
	Title       string `json:"title" example:"Read chapter 4 and answer 3 questions"`
	Description string `json:"description" example:"Long-form description (optional)"`
}

// Create godoc
// @Summary      Create an assignment for a session
// @Tags         Assignments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      createAssignmentReq  true  "Assignment"
// @Success      201   {object}  repo.Assignment
// @Router       /assignments [post]
func (h *AssignmentHandler) Create(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	role, _ := middleware.RoleFromContext(r.Context())

	var req createAssignmentReq
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, badJSONMessage(err))
		return
	}
	a, err := h.svc.Create(r.Context(), uid, service.CreateAssignmentInput{
		SessionID:   req.SessionID,
		Title:       req.Title,
		Description: req.Description,
		ActorRole:   role,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, a)
}

// List godoc
// @Summary      List assignments visible to the current user
// @Description  admin/manager see all (with creator email), teacher sees own + their courses', student sees enrolled courses'.
// @Tags         Assignments
// @Produce      json
// @Security     BearerAuth
// @Param        limit  query  int  false  "Max results"  default(100)
// @Success      200    {object}  listResponse
// @Router       /assignments [get]
func (h *AssignmentHandler) List(w http.ResponseWriter, r *http.Request) {
	uid, _ := middleware.UserIDFromContext(r.Context())
	role, _ := middleware.RoleFromContext(r.Context())
	limit, err := parseLimitParam(r, 100)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	items, err := h.svc.ListForActor(r.Context(), uid, role, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "count": len(items)})
}

// Submissions returns all submissions for an assignment.
// @Summary      List submissions for an assignment
// @Tags         Assignments
// @Produce      json
// @Security     BearerAuth
// @Param        id   path  int  true  "Assignment ID"
// @Success      200  {object}  listResponse
// @Router       /assignments/{id}/submissions [get]
func (h *AssignmentHandler) Submissions(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	items, err := h.manage.Submissions(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "count": len(items)})
}

// Delete godoc
// @Summary      Delete an assignment
// @Tags         Assignments
// @Security     BearerAuth
// @Param        id   path  int  true  "Assignment ID"
// @Success      204
// @Router       /assignments/{id} [delete]
func (h *AssignmentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	role, _ := middleware.RoleFromContext(r.Context())
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.svc.Delete(r.Context(), uid, role, id); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
