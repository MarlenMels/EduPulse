package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"edupulse/internal/repo"
)

var (
	ErrFFmpegUnavailable = errors.New("video conversion unavailable")
	ErrFileTooLarge      = errors.New("file too large")
	ErrUnsupportedFormat = errors.New("unsupported video format")
)

var allowedVideoExts = map[string]struct{}{
	".mp4": {},
	".mov": {},
	".mkv": {},
}

type VideoService struct {
	uploads  *repo.VideoRepo
	courses  *repo.CourseRepo
	audit    *AuditService
	videoDir string
	hlsDir   string
	maxSize  int64

	ffmpegAvailable bool

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewVideoService(uploads *repo.VideoRepo, courses *repo.CourseRepo, audit *AuditService, videoDir, hlsDir string, maxSize int64) *VideoService {
	ctx, cancel := context.WithCancel(context.Background())
	s := &VideoService{
		uploads:  uploads,
		courses:  courses,
		audit:    audit,
		videoDir: videoDir,
		hlsDir:   hlsDir,
		maxSize:  maxSize,
		ctx:      ctx,
		cancel:   cancel,
	}
	_, err := exec.LookPath("ffmpeg")
	s.ffmpegAvailable = err == nil
	return s
}

// Shutdown cancels any running ffmpeg processes and waits for goroutines to finish.
func (s *VideoService) Shutdown() {
	s.cancel()
	s.wg.Wait()
}

func (s *VideoService) FFmpegAvailable() bool { return s.ffmpegAvailable }

// CheckFFmpeg verifies that ffmpeg is on PATH. Called once at startup.
func CheckFFmpeg() error {
	_, err := exec.LookPath("ffmpeg")
	return err
}

func (s *VideoService) SaveAndConvert(ctx context.Context, lessonID, userID int64, file multipart.File, header *multipart.FileHeader) (repo.VideoUpload, error) {
	if !s.ffmpegAvailable {
		return repo.VideoUpload{}, ErrFFmpegUnavailable
	}
	if header.Size <= 0 || header.Size > s.maxSize {
		return repo.VideoUpload{}, ErrFileTooLarge
	}
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if _, ok := allowedVideoExts[ext]; !ok {
		return repo.VideoUpload{}, ErrUnsupportedFormat
	}

	if err := os.MkdirAll(s.videoDir, 0o755); err != nil {
		return repo.VideoUpload{}, err
	}

	uniqueName := fmt.Sprintf("lesson_%d_%d.mp4", lessonID, time.Now().Unix())
	storedPath := filepath.Join(s.videoDir, uniqueName)

	dst, err := os.Create(storedPath)
	if err != nil {
		return repo.VideoUpload{}, err
	}
	if _, err := io.Copy(dst, file); err != nil {
		dst.Close()
		_ = os.Remove(storedPath)
		return repo.VideoUpload{}, err
	}
	dst.Close()

	upload, err := s.uploads.Create(ctx, repo.VideoUpload{
		LessonID:         lessonID,
		OriginalFilename: header.Filename,
		StoredPath:       storedPath,
		Status:           "processing",
	})
	if err != nil {
		_ = os.Remove(storedPath)
		return repo.VideoUpload{}, err
	}

	if err := s.courses.UpdateVideoStatus(ctx, lessonID, "processing", ""); err != nil {
		log.Printf("video: lesson %d status update failed: %v", lessonID, err)
	}

	if s.audit != nil {
		_ = s.audit.Log(ctx, userID, "video.upload", "lesson", lessonID, map[string]any{
			"upload_id": upload.ID,
			"filename":  header.Filename,
			"size":      header.Size,
		})
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.convertToHLS(upload.ID, lessonID, userID, storedPath)
	}()

	return upload, nil
}

func (s *VideoService) convertToHLS(uploadID, lessonID, userID int64, srcPath string) {
	outDir := filepath.Join(s.hlsDir, fmt.Sprintf("lesson_%d", lessonID))
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		s.markFailed(uploadID, lessonID, userID, err.Error())
		return
	}

	playlistPath := filepath.Join(outDir, "playlist.m3u8")

	cmd := exec.CommandContext(s.ctx, "ffmpeg",
		"-i", srcPath,
		"-profile:v", "baseline", "-level", "3.0",
		"-start_number", "0",
		"-hls_time", "10", "-hls_list_size", "0",
		"-force_key_frames", "expr:gte(t,n_forced*10)",
		"-sc_threshold", "0",
		"-vf", "scale=1280:720",
		"-c:v", "libx264", "-pix_fmt", "yuv420p",
		"-c:a", "aac", "-b:a", "128k",
		"-f", "hls", "-y",
		playlistPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		if s.ctx.Err() != nil {
			log.Printf("ffmpeg cancelled for lesson %d (shutdown)", lessonID)
		} else {
			log.Printf("ffmpeg failed for lesson %d: %v\n%s", lessonID, err, string(output))
		}
		s.markFailed(uploadID, lessonID, userID, err.Error())
		return
	}

	hlsURL := fmt.Sprintf("/hls/lesson_%d/playlist.m3u8", lessonID)

	bg := context.Background()
	if err := s.uploads.UpdateStatus(bg, uploadID, "ready", hlsURL, ""); err != nil {
		log.Printf("video: upload %d status update failed: %v", uploadID, err)
	}
	if err := s.courses.UpdateVideoStatus(bg, lessonID, "ready", hlsURL); err != nil {
		log.Printf("video: lesson %d status update failed: %v", lessonID, err)
	}
	if s.audit != nil {
		_ = s.audit.Log(bg, userID, "video.ready", "lesson", lessonID, map[string]any{
			"upload_id": uploadID,
			"hls_url":   hlsURL,
		})
	}
}

func (s *VideoService) markFailed(uploadID, lessonID, userID int64, errMsg string) {
	bg := context.Background()
	_ = s.uploads.UpdateStatus(bg, uploadID, "failed", "", errMsg)
	_ = s.courses.UpdateVideoStatus(bg, lessonID, "failed", "")
	if s.audit != nil {
		_ = s.audit.Log(bg, userID, "video.failed", "lesson", lessonID, map[string]any{
			"upload_id": uploadID,
			"error":     errMsg,
		})
	}
}

func (s *VideoService) GetStatus(ctx context.Context, lessonID int64) (*repo.VideoUpload, error) {
	return s.uploads.GetByLesson(ctx, lessonID)
}
