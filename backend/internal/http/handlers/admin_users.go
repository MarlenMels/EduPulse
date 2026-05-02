package handlers

import (
	"net/http"
	"strconv"

	"edupulse/internal/auth"
	"edupulse/internal/middleware"
	"edupulse/internal/repo"

	"github.com/go-chi/chi/v5"
)

type AdminUsersHandler struct {
	users *repo.UserRepo
}

func NewAdminUsersHandler(users *repo.UserRepo) *AdminUsersHandler {
	return &AdminUsersHandler{users: users}
}

type adminCreateUserReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

// List godoc
// @Summary      List all users (admin)
// @Tags         AdminUsers
// @Produce      json
// @Security     BearerAuth
// @Param        limit  query     int  false  "Max results"  default(100)
// @Success      200    {object}  listResponse
// @Router       /admin/users [get]
func (h *AdminUsersHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, err := parseLimitParam(r, 100)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	items, err := h.users.List(r.Context(), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"items": items,
		"count": len(items),
	})
}

// Create godoc
// @Summary      Create user with any role (admin)
// @Tags         AdminUsers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      adminCreateUserReq  true  "User data"
// @Success      201   {object}  repo.User
// @Failure      400   {object}  errorResponse
// @Failure      409   {object}  errorResponse
// @Router       /admin/users [post]
func (h *AdminUsersHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req adminCreateUserReq
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, badJSONMessage(err))
		return
	}
	email, err := normalizeRequiredText(req.Email, "email", 254)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if !emailRegex.MatchString(email) {
		writeError(w, http.StatusBadRequest, "invalid email format")
		return
	}
	if len(req.Password) < minPasswordLen {
		writeError(w, http.StatusBadRequest, "password must be at least 6 characters")
		return
	}
	if len(req.Password) > maxPasswordLen {
		writeError(w, http.StatusBadRequest, "password must be at most 72 characters")
		return
	}
	if !auth.IsValidRole(req.Role) {
		writeError(w, http.StatusBadRequest, "invalid role, must be one of: admin, manager, teacher, student, parent")
		return
	}

	existing, err := h.users.GetByEmail(r.Context(), email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	if existing != nil {
		writeError(w, http.StatusConflict, "email already registered")
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "hash error")
		return
	}
	u, err := h.users.Create(r.Context(), email, hash, req.Role)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	writeJSON(w, http.StatusCreated, u)
}

// Delete godoc
// @Summary      Delete user (admin)
// @Tags         AdminUsers
// @Security     BearerAuth
// @Param        id   path  int  true  "User ID"
// @Success      204
// @Failure      400  {object}  errorResponse
// @Router       /admin/users/{id} [delete]
func (h *AdminUsersHandler) Delete(w http.ResponseWriter, r *http.Request) {
	actorID, _ := middleware.UserIDFromContext(r.Context())
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if id == actorID {
		writeError(w, http.StatusBadRequest, "cannot delete yourself")
		return
	}
	if err := h.users.Delete(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
