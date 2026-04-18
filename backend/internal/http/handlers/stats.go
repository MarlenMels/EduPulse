package handlers

import (
	"net/http"

	"edupulse/internal/service"
)

type StatsHandler struct{ svc *service.StatsService }

func NewStatsHandler(svc *service.StatsService) *StatsHandler { return &StatsHandler{svc: svc} }

// Get godoc
// @Summary      Platform statistics
// @Description  User counts per role and online status. Roles: admin
// @Tags         Stats
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  service.StatsResult
// @Failure      500  {object}  errorResponse
// @Router       /stats [get]
func (h *StatsHandler) Get(w http.ResponseWriter, r *http.Request) {
	res, err := h.svc.Get(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	writeJSON(w, http.StatusOK, res)
}
