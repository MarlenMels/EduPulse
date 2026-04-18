package handlers

import (
	"net/http"
	"strings"

	"edupulse/internal/auth"
)

type AuthHandler struct {
	svc *auth.Service
}

func NewAuthHandler(svc *auth.Service) *AuthHandler { return &AuthHandler{svc: svc} }

type loginReq struct {
	Email    string `json:"email" example:"user@example.com"`
	Password string `json:"password" example:"secret123"`
}

// Login godoc
// @Summary      Login user
// @Description  Authenticate a user with email and password, returns a JWT token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      loginReq  true  "Login credentials"
// @Success      200   {object}  auth.LoginResult
// @Failure      400   {object}  errorResponse
// @Failure      401   {object}  errorResponse
// @Router       /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	res, err := h.svc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	writeJSON(w, http.StatusOK, res)
}

type registerReq struct {
	Email    string `json:"email" example:"user@example.com"`
	Password string `json:"password" example:"secret123"`
	Role     string `json:"role" example:"student" enums:"admin,manager,teacher,student,parent"`
}

// Register godoc
// @Summary      Register user
// @Description  Create a new user account and return a JWT token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      registerReq  true  "Registration data"
// @Success      201   {object}  auth.LoginResult
// @Failure      400   {object}  errorResponse
// @Failure      409   {object}  errorResponse
// @Router       /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerReq
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}
	if !auth.IsValidRole(req.Role) {
		writeError(w, http.StatusBadRequest, "invalid role, must be one of: admin, manager, teacher, student, parent")
		return
	}

	res, err := h.svc.Register(r.Context(), req.Email, req.Password, req.Role)
	if err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, res)
}
