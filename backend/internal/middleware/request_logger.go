package middleware

import (
	"log"
	"net/http"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := chimw.GetReqID(r.Context())
		start := time.Now()
		log.Printf("request_start request_id=%s method=%s path=%s remote=%s", requestID, r.Method, r.URL.Path, r.RemoteAddr)

		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(recorder, r)

		duration := time.Since(start)
		if recorder.status >= http.StatusBadRequest {
			log.Printf("request_failure request_id=%s method=%s path=%s status=%d duration_ms=%d", requestID, r.Method, r.URL.Path, recorder.status, duration.Milliseconds())
			return
		}
		log.Printf("request_success request_id=%s method=%s path=%s status=%d duration_ms=%d", requestID, r.Method, r.URL.Path, recorder.status, duration.Milliseconds())
	})
}
