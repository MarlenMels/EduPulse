package handlers

import (
	"net/http"
	"strconv"

	"edupulse/internal/service"
)

type AnalyticsHandler struct{ svc *service.AnalyticsService }

func NewAnalyticsHandler(svc *service.AnalyticsService) *AnalyticsHandler { return &AnalyticsHandler{svc: svc} }

func (h *AnalyticsHandler) ListSessionsByH3(w http.ResponseWriter, r *http.Request) {
	h3 := r.URL.Query().Get("h3")
	day := r.URL.Query().Get("day") // YYYY-MM-DD
	limit := 100
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	rows, err := h.svc.ListSessionsByH3(r.Context(), h3, day, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"items": rows,
		"count": len(rows),
	})
}