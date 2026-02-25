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
	BranchID  int64   `json:"branch_id"`
	TeacherID int64   `json:"teacher_id,omitempty"`
	Title     string  `json:"title"`
	StartTime string  `json:"start_time"`
	Lat       float64 `json:"lat"`
	Lng       float64 `json:"lng"`
}

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
		BranchID:  req.BranchID,
		TeacherID: req.TeacherID,
		Title:     req.Title,
		StartTime: st,
		Lat:       req.Lat,
		Lng:       req.Lng,
		ActorRole: role,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, s)
}

func (h *SessionHandler) List(w http.ResponseWriter, r *http.Request) {
	h3 := r.URL.Query().Get("h3")
	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	sessions, err := h.write.List(r.Context(), h3, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"items": sessions,
		"count": len(sessions),
	})
}

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

func (h *SessionHandler) Nearby(w http.ResponseWriter, r *http.Request) {
	center := r.URL.Query().Get("h3")
	k := 1
	if v := r.URL.Query().Get("k"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			k = n
		}
	}
	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}

	items, err := h.read.NearbyByH3(r.Context(), center, k, limit)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"items": items,
		"count": len(items),
	})
}