package httpapi

import (
	"net/http"

	"edupulse/internal/auth"
	"edupulse/internal/http/handlers"
	"edupulse/internal/middleware"
	"edupulse/internal/service"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

type Deps struct {
	JWTSecret string

	AuthSvc             *auth.Service
	UserSvc             *service.UserService
	BranchSvc           *service.BranchService
	BranchReadSvc       *service.BranchReadService
	SessionSvc          *service.SessionService
	SessionReadSvc      *service.SessionReadService
	HomeworkSvc         *service.HomeworkService
	HomeworkManageSvc   *service.HomeworkManageService
	AuditSvc            *service.AuditService
	AnalyticsSvc        *service.AnalyticsService
	NotificationSvc     *service.NotificationService
}

type Server struct {
	deps Deps
	r    chi.Router
}

func NewServer(d Deps) *Server {
	s := &Server{deps: d}
	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Recoverer)

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Public
	authH := handlers.NewAuthHandler(d.AuthSvc)
	r.Route("/auth", func(r chi.Router) {
		r.Post("/login", authH.Login)
	})

	// Protected
	authMW := middleware.AuthJWT(d.JWTSecret)
	r.Group(func(r chi.Router) {
		r.Use(authMW)

		userH := handlers.NewUserHandler(d.UserSvc)

		branchH := handlers.NewBranchHandler(d.BranchSvc, d.BranchReadSvc)
		sessionH := handlers.NewSessionHandler(d.SessionSvc, d.SessionReadSvc)
		hwH := handlers.NewHomeworkHandler(d.HomeworkSvc, d.HomeworkManageSvc)

		auditH := handlers.NewAuditHandler(d.AuditSvc)
		analyticsH := handlers.NewAnalyticsHandler(d.AnalyticsSvc)
		notifH := handlers.NewNotificationHandler(d.NotificationSvc)

		// Profile
		r.Get("/users/me", userH.Me)

		// Branches
		r.Post("/branches", middleware.RBAC(auth.RoleAdmin, auth.RoleManager)(http.HandlerFunc(branchH.Create)).ServeHTTP)
		r.Get("/branches", branchH.List)
		r.Get("/branches/{id}", branchH.Get)

		// Sessions
		r.Post("/sessions", middleware.RBAC(auth.RoleManager, auth.RoleTeacher)(http.HandlerFunc(sessionH.Create)).ServeHTTP)
		r.Get("/sessions", sessionH.List)
		r.Get("/sessions/{id}", sessionH.Get)
		r.Get("/sessions/nearby", sessionH.Nearby)

		// Homework
		r.Post("/homework/submit", middleware.RBAC(auth.RoleStudent)(http.HandlerFunc(hwH.Submit)).ServeHTTP)
		r.Get("/homework", middleware.RBAC(auth.RoleTeacher, auth.RoleManager, auth.RoleAdmin)(http.HandlerFunc(hwH.List)).ServeHTTP)
		r.Get("/homework/mine", middleware.RBAC(auth.RoleStudent)(http.HandlerFunc(hwH.Mine)).ServeHTTP)
		r.Patch("/homework/{id}/status", middleware.RBAC(auth.RoleTeacher)(http.HandlerFunc(hwH.UpdateStatus)).ServeHTTP)

		// Audit / analytics / notifications
		r.Get("/audit-logs", middleware.RBAC(auth.RoleAdmin)(http.HandlerFunc(auditH.List)).ServeHTTP)
		r.Get("/analytics/sessions-by-h3", middleware.RBAC(auth.RoleAdmin, auth.RoleManager)(http.HandlerFunc(analyticsH.ListSessionsByH3)).ServeHTTP)
		r.Get("/notifications", middleware.RBAC(auth.RoleAdmin, auth.RoleManager)(http.HandlerFunc(notifH.List)).ServeHTTP)
	})

	s.r = r
	return s
}

func (s *Server) Router() http.Handler { return s.r }