package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"edupulse/internal/middleware"
	"edupulse/internal/service"
)

type VideoHandler struct{ svc *service.VideoService }

func NewVideoHandler(svc *service.VideoService) *VideoHandler { return &VideoHandler{svc: svc} }

// Upload godoc
// @Summary      Upload video for a lesson
// @Description  Upload a video file for a lesson. Conversion to HLS runs asynchronously. Max 500MB. Allowed formats: mp4, mov, mkv. Roles: admin, manager, teacher
// @Tags         Videos
// @Accept       mpfd
// @Produce      json
// @Security     BearerAuth
// @Param        id     path      int   true  "Lesson ID"
// @Param        video  formData  file  true  "Video file"
// @Success      202    {object}  repo.VideoUpload
// @Failure      400    {object}  errorResponse
// @Failure      401    {object}  errorResponse
// @Failure      413    {object}  errorResponse
// @Failure      503    {object}  errorResponse
// @Router       /lessons/{id}/video [post]
func (h *VideoHandler) Upload(w http.ResponseWriter, r *http.Request) {
	lessonID, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if lessonID <= 0 {
		writeError(w, http.StatusBadRequest, "invalid lesson id")
		return
	}

	const maxMemory int64 = 32 << 20 // 32 MB in-memory, rest goes to tmp files
	if err := r.ParseMultipartForm(maxMemory); err != nil {
		writeError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}

	file, header, err := r.FormFile("video")
	if err != nil {
		writeError(w, http.StatusBadRequest, "video file field is required")
		return
	}
	defer file.Close()

	userID, _ := middleware.UserIDFromContext(r.Context())

	upload, err := h.svc.SaveAndConvert(r.Context(), lessonID, userID, file, header)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrFFmpegUnavailable):
			writeError(w, http.StatusServiceUnavailable, "video conversion unavailable")
		case errors.Is(err, service.ErrFileTooLarge):
			writeError(w, http.StatusRequestEntityTooLarge, "file too large (max 500MB)")
		case errors.Is(err, service.ErrUnsupportedFormat):
			writeError(w, http.StatusBadRequest, "unsupported format (allowed: mp4, mov, mkv)")
		default:
			writeError(w, http.StatusInternalServerError, "upload failed")
		}
		return
	}

	writeJSON(w, http.StatusAccepted, upload)
}

// Status godoc
// @Summary      Get video upload status for a lesson
// @Description  Returns the latest video upload record for the lesson.
// @Tags         Videos
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Lesson ID"
// @Success      200  {object}  repo.VideoUpload
// @Failure      404  {object}  errorResponse
// @Router       /lessons/{id}/video [get]
func (h *VideoHandler) Status(w http.ResponseWriter, r *http.Request) {
	lessonID, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if lessonID <= 0 {
		writeError(w, http.StatusBadRequest, "invalid lesson id")
		return
	}
	upload, err := h.svc.GetStatus(r.Context(), lessonID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	if upload == nil {
		writeError(w, http.StatusNotFound, "no upload found")
		return
	}
	writeJSON(w, http.StatusOK, upload)
}
