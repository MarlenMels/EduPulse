package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"edupulse/internal/middleware"
	"edupulse/internal/service"
)

type SessionHandler struct {
	write *service.SessionService
	read  *service.SessionReadService
}

func NewSessionHandler(write *service.SessionService, read *service.SessionReadService) *SessionHandler {
	return &SessionHandler{write: write, read: read}
}

type createSessionReq struct {
	CourseID  int64  `json:"course_id" example:"1"`
	Title     string `json:"title" example:"Math 101 — lecture 1"`
	StartTime string `json:"start_time" example:"2026-05-01T10:00:00Z"`
}

// Create godoc
// @Summary      Create a session
// @Description  Create a new teaching session for a course. Roles: admin, manager, teacher (only on courses they teach).
// @Tags         Sessions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      createSessionReq  true  "Session data"
// @Success      201   {object}  repo.Session
// @Failure      400   {object}  errorResponse
// @Failure      401   {object}  errorResponse
// @Router       /sessions [post]
func (h *SessionHandler) Create(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	role, _ := middleware.RoleFromContext(r.Context())

	var req createSessionReq
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, badJSONMessage(err))
		return
	}
	title, err := normalizeRequiredText(req.Title, "title", 120)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	st, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		writeError(w, http.StatusBadRequest, "start_time must be RFC3339, e.g. 2026-02-24T15:00:00Z")
		return
	}

	s, err := h.write.Create(r.Context(), uid, service.CreateSessionInput{
		CourseID:  req.CourseID,
		Title:     title,
		StartTime: st,
		ActorRole: role,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, s)
}

// List godoc
// @Summary      List sessions visible to the current user
// @Description  admin/manager see all, teacher sees their courses' sessions, student sees enrolled courses' sessions.
// @Tags         Sessions
// @Produce      json
// @Security     BearerAuth
// @Param        limit  query     int  false  "Max results"  default(50)
// @Success      200    {object}  listResponse
// @Router       /sessions [get]
func (h *SessionHandler) List(w http.ResponseWriter, r *http.Request) {
	uid, _ := middleware.UserIDFromContext(r.Context())
	role, _ := middleware.RoleFromContext(r.Context())
	limit, err := parseLimitParam(r, 50)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	sessions, err := h.write.ListForActor(r.Context(), uid, role, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"items": sessions,
		"count": len(sessions),
	})
}

// Get godoc
// @Summary      Get session by ID
// @Tags         Sessions
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Session ID"
// @Success      200  {object}  repo.Session
// @Failure      404  {object}  errorResponse
// @Router       /sessions/{id} [get]
func (h *SessionHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	s, err := h.read.Get(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, s)
}

func (h *SessionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.write.Delete(r.Context(), uid, id); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
