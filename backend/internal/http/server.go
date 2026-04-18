package httpapi

import (
	"net/http"
	"strings"

	"edupulse/internal/auth"
	"edupulse/internal/http/handlers"
	"edupulse/internal/middleware"
	"edupulse/internal/repo"
	"edupulse/internal/service"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

type Deps struct {
	JWTSecret string

	AuthSvc           *auth.Service
	UserSvc           *service.UserService
	SessionSvc        *service.SessionService
	SessionReadSvc    *service.SessionReadService
	HomeworkSvc       *service.HomeworkService
	HomeworkManageSvc *service.HomeworkManageService
	AuditSvc          *service.AuditService
	NotificationSvc   *service.NotificationService
	CourseSvc         *service.CourseService
	StatsSvc          *service.StatsService
	VideoSvc          *service.VideoService
	UserRepo          *repo.UserRepo
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
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	}))

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Swagger UI — auto-authorize after login/register
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.PersistAuthorization(true),
		httpSwagger.AfterScript(`
			// Intercept fetch to auto-apply token from login/register responses
			const origFetch = window.fetch;
			window.fetch = async function() {
				const res = await origFetch.apply(this, arguments);
				const url = arguments[0] || '';
				if (typeof url === 'string' && (url.includes('/auth/login') || url.includes('/auth/register'))) {
					const clone = res.clone();
					try {
						const body = await clone.json();
						if (body.token) {
							const token = 'Bearer ' + body.token;
							window.ui.preauthorizeApiKey('BearerAuth', token);
							console.log('Swagger: token auto-applied');
						}
					} catch(e) {}
				}
				return res;
			};
		`),
	))

	// Static files
	r.Handle("/uploads/*", http.StripPrefix("/uploads", http.FileServer(http.Dir("./uploads"))))
	r.Handle("/videos/*", http.StripPrefix("/videos", http.FileServer(http.Dir("./videos"))))
	r.Handle("/hls/*", http.StripPrefix("/hls", hlsFileHandler("./hls")))

	// Public
	authH := handlers.NewAuthHandler(d.AuthSvc)
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", authH.Register)
		r.Post("/login", authH.Login)
	})

	// Protected
	authMW := middleware.AuthJWT(d.JWTSecret)
	r.Group(func(r chi.Router) {
		r.Use(authMW)
		r.Use(middleware.LastSeen(d.UserRepo))

		userH := handlers.NewUserHandler(d.UserSvc)
		sessionH := handlers.NewSessionHandler(d.SessionSvc, d.SessionReadSvc)
		hwH := handlers.NewHomeworkHandler(d.HomeworkSvc, d.HomeworkManageSvc)
		auditH := handlers.NewAuditHandler(d.AuditSvc)
		notifH := handlers.NewNotificationHandler(d.NotificationSvc)
		courseH := handlers.NewCourseHandler(d.CourseSvc)

		// Profile
		r.Get("/users/me", userH.Me)

		// Sessions
		r.Post("/sessions", middleware.RBAC(auth.RoleAdmin, auth.RoleManager, auth.RoleTeacher)(http.HandlerFunc(sessionH.Create)).ServeHTTP)
		r.Get("/sessions", sessionH.List)
		r.Get("/sessions/{id}", sessionH.Get)

		// Homework
		r.Post("/homework/submit", middleware.RBAC(auth.RoleAdmin, auth.RoleStudent)(http.HandlerFunc(hwH.Submit)).ServeHTTP)
		r.Get("/homework", middleware.RBAC(auth.RoleAdmin, auth.RoleTeacher, auth.RoleManager)(http.HandlerFunc(hwH.List)).ServeHTTP)
		r.Get("/homework/mine", middleware.RBAC(auth.RoleAdmin, auth.RoleStudent)(http.HandlerFunc(hwH.Mine)).ServeHTTP)
		r.Patch("/homework/{id}/status", middleware.RBAC(auth.RoleAdmin, auth.RoleTeacher)(http.HandlerFunc(hwH.UpdateStatus)).ServeHTTP)

		// Courses
		r.Post("/courses", middleware.RBAC(auth.RoleAdmin, auth.RoleManager, auth.RoleTeacher)(http.HandlerFunc(courseH.Create)).ServeHTTP)
		r.Get("/courses", courseH.List)
		r.Post("/courses/{id}/lessons", middleware.RBAC(auth.RoleAdmin, auth.RoleManager, auth.RoleTeacher)(http.HandlerFunc(courseH.AddLesson)).ServeHTTP)
		r.Put("/courses/{id}/lessons/{lessonId}", middleware.RBAC(auth.RoleAdmin, auth.RoleManager, auth.RoleTeacher)(http.HandlerFunc(courseH.UpdateLesson)).ServeHTTP)

		// Lesson videos (HLS)
		videoH := handlers.NewVideoHandler(d.VideoSvc)
		r.Post("/lessons/{id}/video", middleware.RBAC(auth.RoleAdmin, auth.RoleManager, auth.RoleTeacher)(http.HandlerFunc(videoH.Upload)).ServeHTTP)
		r.Get("/lessons/{id}/video", videoH.Status)

		// Stats (admin only)
		statsH := handlers.NewStatsHandler(d.StatsSvc)
		r.Get("/stats", middleware.RBAC(auth.RoleAdmin)(http.HandlerFunc(statsH.Get)).ServeHTTP)

		// Audit / notifications
		r.Get("/audit-logs", middleware.RBAC(auth.RoleAdmin)(http.HandlerFunc(auditH.List)).ServeHTTP)
		r.Get("/notifications", middleware.RBAC(auth.RoleAdmin, auth.RoleManager)(http.HandlerFunc(notifH.List)).ServeHTTP)
	})

	s.r = r
	return s
}

func (s *Server) Router() http.Handler { return s.r }

// hlsFileHandler serves HLS files from dir and sets the correct Content-Type
// for .m3u8 playlists and .ts segments before delegating to the file server.
func hlsFileHandler(dir string) http.Handler {
	fs := http.FileServer(http.Dir(dir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, ".m3u8"):
			w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
		case strings.HasSuffix(r.URL.Path, ".ts"):
			w.Header().Set("Content-Type", "video/mp2t")
		}
		w.Header().Set("Cache-Control", "no-cache")
		fs.ServeHTTP(w, r)
	})
}
