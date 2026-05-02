package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type UploadHandler struct {
	dir string
}

func NewUploadHandler(dir string) *UploadHandler {
	return &UploadHandler{dir: dir}
}

func (h *UploadHandler) Create(w http.ResponseWriter, r *http.Request) {
	const maxMemory int64 = 16 << 20
	const maxSize int64 = 50 << 20
	if err := r.ParseMultipartForm(maxMemory); err != nil {
		writeError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file field is required")
		return
	}
	defer file.Close()

	if header.Size <= 0 || header.Size > maxSize {
		writeError(w, http.StatusRequestEntityTooLarge, "file too large (max 50MB)")
		return
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext == "" {
		writeError(w, http.StatusBadRequest, "file extension is required")
		return
	}

	if err := os.MkdirAll(h.dir, 0o755); err != nil {
		writeError(w, http.StatusInternalServerError, "upload failed")
		return
	}

	name := fmt.Sprintf("file_%d%s", time.Now().UnixNano(), ext)
	path := filepath.Join(h.dir, name)
	dst, err := os.Create(path)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "upload failed")
		return
	}
	if _, err := io.Copy(dst, file); err != nil {
		dst.Close()
		_ = os.Remove(path)
		writeError(w, http.StatusInternalServerError, "upload failed")
		return
	}
	dst.Close()

	writeJSON(w, http.StatusCreated, map[string]any{
		"url":  "/uploads/" + name,
		"name": header.Filename,
		"size": header.Size,
	})
}
