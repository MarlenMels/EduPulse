package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"edupulse/internal/middleware"
	"edupulse/internal/service"
)

type HomeworkHandler struct {
	submit *service.HomeworkService
	manage *service.HomeworkManageService
}

func NewHomeworkHandler(submit *service.HomeworkService, manage *service.HomeworkManageService) *HomeworkHandler {
	return &HomeworkHandler{submit: submit, manage: manage}
}

type submitHomeworkReq struct {
	AssignmentID int64  `json:"assignment_id" example:"1"`
	Content      string `json:"content" example:"My homework answer"`
	Attachments  string `json:"attachments,omitempty" example:""`
}

// Submit godoc
// @Summary      Submit homework for an assignment
// @Tags         Homework
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      submitHomeworkReq  true  "Homework submission"
// @Success      201   {object}  repo.HomeworkSubmission
// @Router       /homework/submit [post]
func (h *HomeworkHandler) Submit(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req submitHomeworkReq
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, badJSONMessage(err))
		return
	}
	sub, err := h.submit.Submit(r.Context(), uid, service.SubmitHomeworkInput{
		AssignmentID: req.AssignmentID,
		Content:      req.Content,
		Attachments:  req.Attachments,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, sub)
}

// Mine returns the current student's submissions across all their assignments.
// @Summary      List my homework submissions
// @Tags         Homework
// @Produce      json
// @Security     BearerAuth
// @Param        limit  query  int  false  "Max results"  default(50)
// @Success      200    {object}  listResponse
// @Router       /homework/mine [get]
func (h *HomeworkHandler) Mine(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	limit, err := parseLimitParam(r, 50)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	items, err := h.submit.Mine(r.Context(), uid, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "count": len(items)})
}

type patchStatusReq struct {
	Status string `json:"status" example:"accepted"`
}

// UpdateStatus changes a submission's status.
// @Summary      Update homework status
// @Tags         Homework
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path  int             true  "Submission ID"
// @Param        body  body  patchStatusReq  true  "New status"
// @Success      200   {object}  repo.HomeworkSubmission
// @Router       /homework/{id}/status [patch]
// List godoc
// @Summary      List all homework submissions
// @Description  Get list of all homework submissions (admin/manager/teacher view)
// @Tags         Homework
// @Produce      json
// @Security     BearerAuth
// @Param        limit  query     int  false  "Max results"  default(50)
// @Success      200   {object}  map[string]interface{}{"items": []repo.HomeworkSubmission, "count": int}
// @Router       /homework [get]
func (h *HomeworkHandler) List(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	role, _ := middleware.RoleFromContext(r.Context())

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	var items []repo.HomeworkSubmission
	var err error

	// Admin/manager can see all submissions
	if role == "admin" || role == "manager" {
		items, err = h.manage.ListAll(r.Context(), limit)
	} else {
		// Teachers can see submissions for their courses
		items, err = h.manage.ListByTeacher(r.Context(), uid, limit)
	}

	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "count": len(items)})
}

func (h *HomeworkHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	role, _ := middleware.RoleFromContext(r.Context())

	subID, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if subID <= 0 {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req patchStatusReq
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, badJSONMessage(err))
		return
	}
	req.Status = strings.TrimSpace(req.Status)

	updated, err := h.manage.UpdateStatus(r.Context(), uid, role, subID, req.Status)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, updated)
}
