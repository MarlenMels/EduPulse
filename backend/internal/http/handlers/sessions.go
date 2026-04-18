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
	TeacherID int64  `json:"teacher_id,omitempty" example:"2"`
	Title     string `json:"title" example:"Math 101"`
	StartTime string `json:"start_time" example:"2026-05-01T10:00:00Z"`
}

// Create godoc
// @Summary      Create a session
// @Description  Create a new teaching session. Roles: admin, manager, teacher
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
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	st, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		writeError(w, http.StatusBadRequest, "start_time must be RFC3339, e.g. 2026-02-24T15:00:00Z")
		return
	}

	s, err := h.write.Create(r.Context(), uid, service.CreateSessionInput{
		TeacherID: req.TeacherID,
		Title:     req.Title,
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
// @Summary      List sessions
// @Description  Retrieve a list of teaching sessions
// @Tags         Sessions
// @Produce      json
// @Security     BearerAuth
// @Param        limit  query     int  false  "Max results"  default(50)
// @Success      200    {object}  listResponse
// @Failure      500    {object}  errorResponse
// @Router       /sessions [get]
func (h *SessionHandler) List(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	sessions, err := h.write.List(r.Context(), limit)
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
// @Description  Retrieve a single session by its ID
// @Tags         Sessions
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Session ID"
// @Success      200  {object}  repo.Session
// @Failure      400  {object}  errorResponse
// @Failure      404  {object}  errorResponse
// @Router       /sessions/{id} [get]
func (h *SessionHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
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
