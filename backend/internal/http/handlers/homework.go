package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"edupulse/internal/middleware"
	"edupulse/internal/repo"
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
	SessionID int64  `json:"session_id" example:"1"`
	Content   string `json:"content" example:"My homework answer"`
}

// Submit godoc
// @Summary      Submit homework
// @Description  Submit a homework assignment. Roles: admin, student
// @Tags         Homework
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      submitHomeworkReq  true  "Homework submission"
// @Success      201   {object}  repo.HomeworkSubmission
// @Failure      400   {object}  errorResponse
// @Failure      401   {object}  errorResponse
// @Router       /homework/submit [post]
func (h *HomeworkHandler) Submit(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req submitHomeworkReq
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	sub, err := h.submit.Submit(r.Context(), uid, service.SubmitHomeworkInput{
		SessionID: req.SessionID,
		Content:   req.Content,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, sub)
}

// List godoc
// @Summary      List homework submissions
// @Description  Retrieve homework submissions with filters. Roles: admin, teacher, manager
// @Tags         Homework
// @Produce      json
// @Security     BearerAuth
// @Param        session_id  query     int     false  "Filter by session ID"
// @Param        student_id  query     int     false  "Filter by student ID"
// @Param        status      query     string  false  "Filter by status"
// @Param        limit       query     int     false  "Max results"  default(50)
// @Success      200         {object}  listResponse
// @Failure      500         {object}  errorResponse
// @Router       /homework [get]
func (h *HomeworkHandler) List(w http.ResponseWriter, r *http.Request) {
	var f repo.HomeworkListFilter
	if v := r.URL.Query().Get("session_id"); v != "" {
		f.SessionID, _ = strconv.ParseInt(v, 10, 64)
	}
	if v := r.URL.Query().Get("student_id"); v != "" {
		f.StudentID, _ = strconv.ParseInt(v, 10, 64)
	}
	f.Status = strings.TrimSpace(r.URL.Query().Get("status"))

	f.Limit = 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			f.Limit = n
		}
	}

	items, err := h.manage.List(r.Context(), f)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"items": items,
		"count": len(items),
	})
}

// Mine godoc
// @Summary      List my homework submissions
// @Description  Retrieve homework submissions for the authenticated student
// @Tags         Homework
// @Produce      json
// @Security     BearerAuth
// @Param        status  query     string  false  "Filter by status"
// @Param        limit   query     int     false  "Max results"  default(50)
// @Success      200     {object}  listResponse
// @Failure      400     {object}  errorResponse
// @Failure      401     {object}  errorResponse
// @Router       /homework/mine [get]
func (h *HomeworkHandler) Mine(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}

	items, err := h.manage.ListMine(r.Context(), uid, status, limit)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"items": items,
		"count": len(items),
	})
}

type patchStatusReq struct {
	Status string `json:"status" example:"approved"`
}

// UpdateStatus godoc
// @Summary      Update homework status
// @Description  Update the status of a homework submission. Roles: admin, teacher
// @Tags         Homework
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int             true  "Submission ID"
// @Param        body  body      patchStatusReq  true  "New status"
// @Success      200   {object}  repo.HomeworkSubmission
// @Failure      400   {object}  errorResponse
// @Failure      401   {object}  errorResponse
// @Router       /homework/{id}/status [patch]
func (h *HomeworkHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	teacherID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	subID, _ := strconv.ParseInt(idStr, 10, 64)
	if subID <= 0 {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req patchStatusReq
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	req.Status = strings.TrimSpace(req.Status)

	updated, err := h.manage.UpdateStatus(r.Context(), teacherID, subID, req.Status)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, updated)
}