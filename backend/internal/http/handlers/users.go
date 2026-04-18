package handlers

import (
	"net/http"

	"edupulse/internal/middleware"
	"edupulse/internal/service"
)

type UserHandler struct {
	svc *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// Me godoc
// @Summary      Get current user profile
// @Description  Returns the profile of the authenticated user
// @Tags         Users
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  repo.User
// @Failure      401  {object}  errorResponse
// @Failure      404  {object}  errorResponse
// @Router       /users/me [get]
func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	u, err := h.svc.Me(r.Context(), uid)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, u)
}