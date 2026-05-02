package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"edupulse/internal/repo"
	"edupulse/internal/service"
)

type lessonInput struct {
	Title       string `json:"title" example:"Introduction to Go"`
	Description string `json:"description" example:"In this lesson you will learn the basics"`
	VideoURL    string `json:"video_url" example:"/videos/lesson1.mp4"`
	FileURL     string `json:"file_url" example:"/uploads/lesson1.pdf"`
	SortOrder   int    `json:"sort_order" example:"1"`
}

type createCourseReq struct {
	Title       string        `json:"title" example:"Go Programming"`
	Description string        `json:"description" example:"Learn Go from scratch"`
	ImageURL    string        `json:"image_url" example:"/uploads/go.png"`
	Lessons     []lessonInput `json:"lessons"`
}

type updateLessonReq struct {
	Title       string `json:"title" example:"Updated lesson title"`
	Description string `json:"description" example:"Updated description"`
	VideoURL    string `json:"video_url" example:"/videos/lesson1.mp4"`
	FileURL     string `json:"file_url" example:"/uploads/lesson1.pdf"`
	SortOrder   int    `json:"sort_order" example:"1"`
}

type lessonAssetReq struct {
	Type             string `json:"type" example:"file"`
	URL              string `json:"url" example:"/uploads/material.pdf"`
	OriginalFilename string `json:"original_filename" example:"material.pdf"`
}

type CourseHandler struct {
	svc *service.CourseService
}

func NewCourseHandler(svc *service.CourseService) *CourseHandler {
	return &CourseHandler{svc: svc}
}

// Create godoc
// @Summary      Create a course with lessons
// @Description  Create a new course, optionally with lessons in one request. Roles: admin, manager, teacher
// @Tags         Courses
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      createCourseReq  true  "Course data with optional lessons array"
// @Success      201   {object}  service.CourseWithLessons
// @Failure      400   {object}  errorResponse
// @Failure      401   {object}  errorResponse
// @Router       /courses [post]
func (h *CourseHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createCourseReq
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, badJSONMessage(err))
		return
	}
	title, err := normalizeRequiredText(req.Title, "title", 120)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	description, err := normalizeOptionalText(req.Description, 500)
	if err != nil {
		writeError(w, http.StatusBadRequest, "description must be 500 characters or less")
		return
	}
	imageURL, err := normalizeOptionalText(req.ImageURL, 255)
	if err != nil {
		writeError(w, http.StatusBadRequest, "image_url must be 255 characters or less")
		return
	}

	lessons := make([]repo.Lesson, 0, len(req.Lessons))
	for _, l := range req.Lessons {
		lessonTitle, err := normalizeRequiredText(l.Title, "lesson title", 120)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		lessons = append(lessons, repo.Lesson{
			Title:       lessonTitle,
			Description: strings.TrimSpace(l.Description),
			VideoURL:    strings.TrimSpace(l.VideoURL),
			FileURL:     strings.TrimSpace(l.FileURL),
			SortOrder:   l.SortOrder,
		})
	}

	result, err := h.svc.CreateWithLessons(r.Context(), repo.Course{
		Title:       title,
		Description: description,
		ImageURL:    imageURL,
	}, lessons)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

// AddLesson godoc
// @Summary      Add a lesson to a course
// @Description  Add a new lesson to an existing course. Roles: admin, manager, teacher
// @Tags         Courses
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int          true  "Course ID"
// @Param        body  body      lessonInput  true  "Lesson data"
// @Success      201   {object}  repo.Lesson
// @Failure      400   {object}  errorResponse
// @Failure      401   {object}  errorResponse
// @Router       /courses/{id}/lessons [post]
func (h *CourseHandler) AddLesson(w http.ResponseWriter, r *http.Request) {
	courseID, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if courseID <= 0 {
		writeError(w, http.StatusBadRequest, "invalid course id")
		return
	}

	var req lessonInput
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, badJSONMessage(err))
		return
	}
	title, err := normalizeRequiredText(req.Title, "title", 120)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	l, err := h.svc.AddLesson(r.Context(), repo.Lesson{
		CourseID:    courseID,
		Title:       title,
		Description: strings.TrimSpace(req.Description),
		VideoURL:    strings.TrimSpace(req.VideoURL),
		FileURL:     strings.TrimSpace(req.FileURL),
		SortOrder:   req.SortOrder,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, l)
}

// UpdateLesson godoc
// @Summary      Update a lesson
// @Description  Update an existing lesson's content. Roles: admin, manager, teacher
// @Tags         Courses
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id         path      int              true  "Course ID"
// @Param        lessonId   path      int              true  "Lesson ID"
// @Param        body       body      updateLessonReq  true  "Updated lesson data"
// @Success      200        {object}  repo.Lesson
// @Failure      400        {object}  errorResponse
// @Failure      401        {object}  errorResponse
// @Router       /courses/{id}/lessons/{lessonId} [put]
func (h *CourseHandler) UpdateLesson(w http.ResponseWriter, r *http.Request) {
	courseID, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	lessonID, _ := strconv.ParseInt(chi.URLParam(r, "lessonId"), 10, 64)
	if courseID <= 0 || lessonID <= 0 {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req updateLessonReq
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, badJSONMessage(err))
		return
	}
	title, err := normalizeRequiredText(req.Title, "title", 120)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	l, err := h.svc.UpdateLesson(r.Context(), repo.Lesson{
		ID:          lessonID,
		CourseID:    courseID,
		Title:       title,
		Description: strings.TrimSpace(req.Description),
		VideoURL:    strings.TrimSpace(req.VideoURL),
		FileURL:     strings.TrimSpace(req.FileURL),
		SortOrder:   req.SortOrder,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, l)
}

// List godoc
// @Summary      List courses with lessons
// @Description  Retrieve courses with their lessons
// @Tags         Courses
// @Produce      json
// @Security     BearerAuth
// @Param        limit  query     int  false  "Max results"  default(50)
// @Success      200    {object}  listResponse
// @Failure      500    {object}  errorResponse
// @Router       /courses [get]
func (h *CourseHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, err := parseLimitParam(r, 50)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	search, err := normalizeSearchQuery(r.URL.Query().Get("search"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	courses, err := h.svc.List(r.Context(), limit, search)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"items": courses,
		"count": len(courses),
	})
}

func (h *CourseHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid course id")
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *CourseHandler) DeleteLesson(w http.ResponseWriter, r *http.Request) {
	courseID, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	lessonID, _ := strconv.ParseInt(chi.URLParam(r, "lessonId"), 10, 64)
	if courseID <= 0 || lessonID <= 0 {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.svc.DeleteLesson(r.Context(), courseID, lessonID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *CourseHandler) AddLessonAsset(w http.ResponseWriter, r *http.Request) {
	lessonID, _ := strconv.ParseInt(chi.URLParam(r, "lessonId"), 10, 64)
	if lessonID <= 0 {
		writeError(w, http.StatusBadRequest, "invalid lesson id")
		return
	}
	var req lessonAssetReq
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, badJSONMessage(err))
		return
	}
	asset, err := h.svc.AddLessonAsset(r.Context(), repo.LessonAsset{
		LessonID:         lessonID,
		Type:             req.Type,
		URL:              req.URL,
		OriginalFilename: req.OriginalFilename,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, asset)
}

func (h *CourseHandler) DeleteLessonAsset(w http.ResponseWriter, r *http.Request) {
	lessonID, _ := strconv.ParseInt(chi.URLParam(r, "lessonId"), 10, 64)
	assetID, _ := strconv.ParseInt(chi.URLParam(r, "assetId"), 10, 64)
	if lessonID <= 0 || assetID <= 0 {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.svc.DeleteLessonAsset(r.Context(), lessonID, assetID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
