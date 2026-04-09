package handlers

import (
	"net/http"
	"strconv"

	"edupulse/internal/service"
)

type NotificationHandler struct{ svc *service.NotificationService }

func NewNotificationHandler(svc *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

func (h *NotificationHandler) List(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	n, err := h.svc.ListRecent(r.Context(), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"items": n,
		"count": len(n),
	})
}