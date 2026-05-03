package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"edupulse/internal/middleware"
	"edupulse/internal/service"
)

type EnrollmentHandler struct {
	svc *service.EnrollmentService
}

func NewEnrollmentHandler(svc *service.EnrollmentService) *EnrollmentHandler {
	return &EnrollmentHandler{svc: svc}
}

type addUserToCourseReq struct {
	UserID int64 `json:"user_id"`
}

// ListTeachers godoc
// @Summary      List teachers of a course
// @Tags         Enrollments
// @Produce      json
// @Security     BearerAuth
// @Param        id   path  int  true  "Course ID"
// @Success      200  {object}  listResponse
// @Router       /courses/{id}/teachers [get]
func (h *EnrollmentHandler) ListTeachers(w http.ResponseWriter, r *http.Request) {
	courseID, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	items, err := h.svc.TeachersByCourse(r.Context(), courseID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "count": len(items)})
}

// AddTeacher godoc
// @Summary      Add a teacher to a course
// @Tags         Enrollments
// @Accept       json
// @Security     BearerAuth
// @Param        id    path  int                  true  "Course ID"
// @Param        body  body  addUserToCourseReq   true  "Teacher user_id"
// @Success      204
// @Router       /admin/courses/{id}/teachers [post]
func (h *EnrollmentHandler) AddTeacher(w http.ResponseWriter, r *http.Request) {
	uid, _ := middleware.UserIDFromContext(r.Context())
	courseID, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	var req addUserToCourseReq
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, badJSONMessage(err))
		return
	}
	if err := h.svc.AddTeacher(r.Context(), uid, courseID, req.UserID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// RemoveTeacher godoc
// @Summary      Remove a teacher from a course
// @Tags         Enrollments
// @Security     BearerAuth
// @Param        id        path  int  true  "Course ID"
// @Param        teacherId path  int  true  "Teacher user ID"
// @Success      204
// @Router       /admin/courses/{id}/teachers/{teacherId} [delete]
func (h *EnrollmentHandler) RemoveTeacher(w http.ResponseWriter, r *http.Request) {
	uid, _ := middleware.UserIDFromContext(r.Context())
	courseID, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	teacherID, _ := strconv.ParseInt(chi.URLParam(r, "teacherId"), 10, 64)
	if err := h.svc.RemoveTeacher(r.Context(), uid, courseID, teacherID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ListStudents godoc
// @Summary      List students enrolled in a course
// @Tags         Enrollments
// @Produce      json
// @Security     BearerAuth
// @Param        id   path  int  true  "Course ID"
// @Success      200  {object}  listResponse
// @Router       /courses/{id}/students [get]
func (h *EnrollmentHandler) ListStudents(w http.ResponseWriter, r *http.Request) {
	courseID, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	items, err := h.svc.StudentsByCourse(r.Context(), courseID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "count": len(items)})
}

// EnrollStudent godoc
// @Summary      Enroll a student in a course
// @Tags         Enrollments
// @Accept       json
// @Security     BearerAuth
// @Param        id    path  int                  true  "Course ID"
// @Param        body  body  addUserToCourseReq   true  "Student user_id"
// @Success      204
// @Router       /admin/courses/{id}/students [post]
func (h *EnrollmentHandler) EnrollStudent(w http.ResponseWriter, r *http.Request) {
	uid, _ := middleware.UserIDFromContext(r.Context())
	courseID, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	var req addUserToCourseReq
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, badJSONMessage(err))
		return
	}
	if err := h.svc.EnrollStudent(r.Context(), uid, courseID, req.UserID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// UnenrollStudent godoc
// @Summary      Remove a student from a course
// @Tags         Enrollments
// @Security     BearerAuth
// @Param        id         path  int  true  "Course ID"
// @Param        studentId  path  int  true  "Student user ID"
// @Success      204
// @Router       /admin/courses/{id}/students/{studentId} [delete]
func (h *EnrollmentHandler) UnenrollStudent(w http.ResponseWriter, r *http.Request) {
	uid, _ := middleware.UserIDFromContext(r.Context())
	courseID, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	studentID, _ := strconv.ParseInt(chi.URLParam(r, "studentId"), 10, 64)
	if err := h.svc.UnenrollStudent(r.Context(), uid, courseID, studentID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// MyStudents godoc
// @Summary      List students of the current teacher (across all their courses)
// @Tags         Enrollments
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  listResponse
// @Router       /teachers/me/students [get]
func (h *EnrollmentHandler) MyStudents(w http.ResponseWriter, r *http.Request) {
	uid, _ := middleware.UserIDFromContext(r.Context())
	items, err := h.svc.StudentsByTeacher(r.Context(), uid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "count": len(items)})
}
