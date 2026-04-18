package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"edupulse/internal/auth"
	"edupulse/internal/db"
	"edupulse/internal/events"
	httpapi "edupulse/internal/http"
	"edupulse/internal/repo"
	"edupulse/internal/service"
	"edupulse/internal/workers"

	_ "edupulse/docs"
)

// @title           EduPulse API
// @version         1.0
// @description     Backend API for the EduPulse educational platform.

// @host      localhost:8080
// @BasePath  /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter token as: Bearer {your_token}

func main() {
	addr := env("EDUPULSE_ADDR", ":8080")
	dbPath := env("EDUPULSE_DB_PATH", "./edupulse.db")
	jwtSecret := env("EDUPULSE_JWT_SECRET", "dev-secret-change-me")

	database, err := db.OpenSQLite(dbPath)
	if err != nil {
		log.Fatalf("db open: %v", err)
	}
	defer database.Close()

	if err := db.Migrate(database); err != nil {
		log.Fatalf("db migrate: %v", err)
	}
	if err := db.Seed(database); err != nil {
		log.Fatalf("db seed: %v", err)
	}

	// Repos
	userRepo := repo.NewUserRepo(database)
	sessionRepo := repo.NewSessionRepo(database)
	hwRepo := repo.NewHomeworkRepo(database)
	auditRepo := repo.NewAuditRepo(database)
	notifRepo := repo.NewNotificationRepo(database)
	courseRepo := repo.NewCourseRepo(database)
	videoRepo := repo.NewVideoRepo(database)

	// Event bus + worker
	bus := events.NewBus(256)

	// Services
	auditSvc := service.NewAuditService(auditRepo)
	notifSvc := service.NewNotificationService(notifRepo)
	authSvc := auth.NewService(userRepo, jwtSecret)
	sessionSvc := service.NewSessionService(sessionRepo, auditSvc)
	hwSvc := service.NewHomeworkService(hwRepo, auditSvc, bus)

	userSvc := service.NewUserService(userRepo)
	sessionReadSvc := service.NewSessionReadService(sessionRepo)
	hwManageSvc := service.NewHomeworkManageService(hwRepo, auditSvc, bus)
	courseSvc := service.NewCourseService(courseRepo)
	statsSvc := service.NewStatsService(userRepo)
	videoSvc := service.NewVideoService(videoRepo, courseRepo, auditSvc, "./videos", "./hls", 500*1024*1024)
	if err := service.CheckFFmpeg(); err != nil {
		log.Printf("warning: ffmpeg not found in PATH; video conversion will be unavailable: %v", err)
	}

	consumer := workers.NewHomeworkConsumer(notifSvc, auditSvc)
	bus.StartWorker(context.Background(), consumer)

	api := httpapi.NewServer(httpapi.Deps{
		JWTSecret:         jwtSecret,
		AuthSvc:           authSvc,
		UserSvc:           userSvc,
		SessionSvc:        sessionSvc,
		SessionReadSvc:    sessionReadSvc,
		HomeworkSvc:       hwSvc,
		HomeworkManageSvc: hwManageSvc,
		AuditSvc:          auditSvc,
		NotificationSvc:   notifSvc,
		CourseSvc:         courseSvc,
		StatsSvc:          statsSvc,
		VideoSvc:          videoSvc,
		UserRepo:          userRepo,
	})

	srv := &http.Server{
		Addr:              addr,
		Handler:           api.Router(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Minute,
		WriteTimeout:      10 * time.Minute,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Printf("EduPulse backend listening on %s (db=%s)", addr, dbPath)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_ = srv.Shutdown(ctx)
	videoSvc.Shutdown()
	bus.Stop()
	log.Println("shutdown complete")
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
