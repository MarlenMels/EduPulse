package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"edupulse/internal/auth"
	dbpkg "edupulse/internal/db"
	"edupulse/internal/middleware"
	"edupulse/internal/repo"
	"edupulse/internal/service"
)

func newTestCourseHandler(t *testing.T) *CourseHandler {
	t.Helper()

	cfg := &dbpkg.Config{Path: filepath.Join(t.TempDir(), "audit.db")}
	database, err := dbpkg.Open(cfg)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	if err := dbpkg.Migrate(database, cfg.Driver); err != nil {
		t.Fatalf("migrate db: %v", err)
	}
	if err := dbpkg.Seed(database); err != nil {
		t.Fatalf("seed db: %v", err)
	}

	courseSvc := service.NewCourseService(repo.NewCourseRepo(database))
	for i := 0; i < 3; i++ {
		_, err := courseSvc.CreateWithLessons(context.Background(), repo.Course{
			Title:       []string{"Algebra Basics", "World History", "Go Foundations"}[i],
			Description: []string{"Numbers and equations", "Events across regions", "Backend programming and APIs"}[i],
		}, []repo.Lesson{
			{Title: "Intro", Description: "Warm-up", SortOrder: 1},
		})
		if err != nil {
			t.Fatalf("seed course %d: %v", i, err)
		}
	}

	return NewCourseHandler(courseSvc)
}

func newTestAuthHandler(t *testing.T) *AuthHandler {
	t.Helper()

	cfg := &dbpkg.Config{Path: filepath.Join(t.TempDir(), "auth.db")}
	database, err := dbpkg.Open(cfg)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	if err := dbpkg.Migrate(database, cfg.Driver); err != nil {
		t.Fatalf("migrate db: %v", err)
	}
	if err := dbpkg.Seed(database); err != nil {
		t.Fatalf("seed db: %v", err)
	}

	return NewAuthHandler(auth.NewService(repo.NewUserRepo(database), "test-secret"))
}

func performJSONRequest(handler http.HandlerFunc, method, target string, body any) *httptest.ResponseRecorder {
	var reader *strings.Reader
	if body == nil {
		reader = strings.NewReader("")
	} else {
		payload, _ := json.Marshal(body)
		reader = strings.NewReader(string(payload))
	}

	req := httptest.NewRequest(method, target, reader)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func TestCourseSearchFilter(t *testing.T) {
	handler := newTestCourseHandler(t)

	rec := performJSONRequest(handler.List, http.MethodGet, "/courses?search=history&limit=10", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Items []struct {
			Title string `json:"title"`
		} `json:"items"`
		Count int `json:"count"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Count != 1 || resp.Items[0].Title != "World History" {
		t.Fatalf("unexpected search result: %+v", resp)
	}
}

func TestAuditBreakCases(t *testing.T) {
	authHandler := newTestAuthHandler(t)
	courseHandler := newTestCourseHandler(t)

	longTitle := strings.Repeat("x", 121)
	longSearch := strings.Repeat("s", 81)

	tests := []struct {
		name   string
		call   func() *httptest.ResponseRecorder
		status int
	}{
		{
			name: "empty input",
			call: func() *httptest.ResponseRecorder {
				return performJSONRequest(authHandler.Register, http.MethodPost, "/auth/register", map[string]any{})
			},
			status: http.StatusBadRequest,
		},
		{
			name: "invalid format",
			call: func() *httptest.ResponseRecorder {
				return performJSONRequest(authHandler.Register, http.MethodPost, "/auth/register", map[string]any{
					"email":    "bad-email",
					"password": "validpass",
					"role":     "student",
				})
			},
			status: http.StatusBadRequest,
		},
		{
			name: "long input",
			call: func() *httptest.ResponseRecorder {
				return performJSONRequest(courseHandler.Create, http.MethodPost, "/courses", map[string]any{
					"title": longTitle,
				})
			},
			status: http.StatusBadRequest,
		},
		{
			name: "invalid query format",
			call: func() *httptest.ResponseRecorder {
				req := httptest.NewRequest(http.MethodGet, "/courses?limit=abc&search="+longSearch, nil)
				rec := httptest.NewRecorder()
				courseHandler.List(rec, req)
				return rec
			},
			status: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		rec := tc.call()
		if rec.Code != tc.status {
			t.Fatalf("%s: expected %d, got %d: %s", tc.name, tc.status, rec.Code, rec.Body.String())
		}
		t.Logf("%s -> %d %s", tc.name, rec.Code, strings.TrimSpace(rec.Body.String()))
	}
}

func TestRapidCourseListRequests(t *testing.T) {
	handler := newTestCourseHandler(t)

	const requests = 25
	var wg sync.WaitGroup
	errs := make(chan string, requests)
	start := time.Now()

	for i := 0; i < requests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodGet, "/courses?limit=20", nil)
			rec := httptest.NewRecorder()
			handler.List(rec, req)
			if rec.Code != http.StatusOK {
				errs <- rec.Body.String()
			}
		}()
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Fatalf("rapid request failed: %s", err)
	}

	t.Logf("rapid requests=%d duration_ms=%d", requests, time.Since(start).Milliseconds())
}

func TestPerformanceSamples(t *testing.T) {
	handler := newTestCourseHandler(t)

	const normalRuns = 100
	normalStart := time.Now()
	for i := 0; i < normalRuns; i++ {
		normal := httptest.NewRecorder()
		handler.List(normal, httptest.NewRequest(http.MethodGet, "/courses?limit=20", nil))
		if normal.Code != http.StatusOK {
			t.Fatalf("normal request failed: %s", normal.Body.String())
		}
	}
	normalDuration := time.Since(normalStart)

	heavyPayload := map[string]any{
		"title":       "Heavy Course",
		"description": strings.Repeat("Detailed lesson plan. ", 20),
		"lessons":     make([]map[string]any, 0, 30),
	}
	for i := 0; i < 30; i++ {
		heavyPayload["lessons"] = append(heavyPayload["lessons"].([]map[string]any), map[string]any{
			"title":       "Lesson " + time.Now().Add(time.Duration(i)*time.Second).Format("150405"),
			"description": strings.Repeat("content ", 20),
			"sort_order":  i + 1,
		})
	}

	const heavyRuns = 10
	heavyStart := time.Now()
	for i := 0; i < heavyRuns; i++ {
		payload := heavyPayload
		payload["title"] = "Heavy Course " + strings.Repeat("X", i+1)
		heavy := performJSONRequest(handler.Create, http.MethodPost, "/courses", payload)
		if heavy.Code != http.StatusCreated {
			t.Fatalf("heavy request failed: %d %s", heavy.Code, heavy.Body.String())
		}
	}
	heavyDuration := time.Since(heavyStart)

	t.Logf(
		"performance normal_total_ms=%d normal_avg_us=%d heavy_total_ms=%d heavy_avg_us=%d",
		normalDuration.Milliseconds(),
		normalDuration.Microseconds()/normalRuns,
		heavyDuration.Milliseconds(),
		heavyDuration.Microseconds()/heavyRuns,
	)
}

func TestRequestLoggerOutput(t *testing.T) {
	var buf bytes.Buffer
	prevWriter := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(prevWriter)

	handler := middleware.RequestLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/fail" {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))

	okReq := httptest.NewRequest(http.MethodGet, "/ok", nil)
	okRec := httptest.NewRecorder()
	handler.ServeHTTP(okRec, okReq)

	failReq := httptest.NewRequest(http.MethodPost, "/fail", nil)
	failRec := httptest.NewRecorder()
	handler.ServeHTTP(failRec, failReq)

	logOutput := buf.String()
	for _, token := range []string{"request_start", "request_success", "request_failure"} {
		if !strings.Contains(logOutput, token) {
			t.Fatalf("missing %s in log output: %s", token, logOutput)
		}
	}
	t.Log(strings.TrimSpace(logOutput))
}
