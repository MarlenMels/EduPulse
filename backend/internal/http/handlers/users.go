package handlers

import (
	"net/http"
	"strings"

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

type changePasswordReq struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req changePasswordReq
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	req.CurrentPassword = strings.TrimSpace(req.CurrentPassword)
	if req.CurrentPassword == "" || req.NewPassword == "" {
		writeError(w, http.StatusBadRequest, "current and new password are required")
		return
	}
	if len(req.NewPassword) < 6 {
		writeError(w, http.StatusBadRequest, "new password must be at least 6 characters")
		return
	}
	if len(req.NewPassword) > 72 {
		writeError(w, http.StatusBadRequest, "new password must be at most 72 characters")
		return
	}
	if err := h.svc.ChangePassword(r.Context(), uid, req.CurrentPassword, req.NewPassword); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
