# EduPulse — Beta Project Evaluation Lab

## Phase 1 — System Audit (1-page system sheet)

### Description (no jargon)
EduPulse is a small online school management platform. Admins/managers/teachers
create courses with video lessons and schedule live sessions. Students watch
lessons and submit homework; teachers grade it. Admins see usage statistics.

### User & problem
- **Users**: admin, manager, teacher, student, parent.
- **Problem**: small training centers track lessons, attendance, and homework
  in chat apps and spreadsheets. EduPulse centralizes content, scheduling,
  homework review, and basic analytics in one app.

### Core workflow
1. Admin/teacher registers and signs in.
2. Teacher creates a course and uploads lesson videos (server transcodes to HLS).
3. Teacher schedules a live session.
4. Students log in, watch lessons, and submit homework for a session.
5. Teacher reviews homework and updates status.
6. Admin views stats (online users by role).

### Technical audit

| Layer | Tech | Notes |
|---|---|---|
| Frontend | Vue 3 + Vite + TypeScript + Tailwind, Pinia, Axios | Hosted on Vercel: `edu-pulse-ashen.vercel.app` |
| Backend | Go 1.25, chi router, JWT auth, bcrypt, swaggo | Hosted on Railway: `edupulse-production-477b.up.railway.app` |
| Database | SQLite (`modernc.org/sqlite`) | Single-file DB on backend host |
| Media | FFmpeg → HLS transcoding worker | Stored under `videos/`, `hls/` on backend FS |
| Deployment | Vercel (frontend) + Railway (backend), GitHub `main` auto-deploys | `VITE_API_URL` env var points frontend → backend |

### API endpoints (no `/api` prefix on backend)

| Method | Path | Auth | Notes |
|---|---|---|---|
| GET | `/healthz` | — | health |
| POST | `/auth/register` | — | returns JWT |
| POST | `/auth/login` | — | returns JWT |
| GET | `/users/me` | bearer | profile |
| GET/POST | `/sessions`, `/sessions/{id}` | bearer + RBAC | live sessions |
| GET/POST | `/courses`, `/courses/{id}/lessons` | bearer + RBAC | courses + lessons |
| PUT | `/courses/{id}/lessons/{lessonId}` | bearer + RBAC | update lesson |
| POST/GET | `/lessons/{id}/video` | bearer + RBAC | HLS upload + status |
| POST/GET/PATCH | `/homework[/...]` | bearer + RBAC | submit/list/grade |
| GET | `/stats` | admin | usage |
| GET | `/audit-logs` | admin | audit |
| GET | `/notifications` | admin/manager | notifications |
| GET | `/swagger/*` | — | Swagger UI |
| GET | `/uploads/*`, `/videos/*`, `/hls/*` | — | static files |

### Data model (SQLite)
`users`, `sessions`, `homework_submissions`, `courses`, `lessons`,
`video_uploads`, `audit_logs`, `notifications`. Foreign keys on, indexes on
hot read paths (`sessions.start_time`, `homework.session_id`,
`lessons.course_id`, `audit_logs.created_at`, `notifications.status`,
`video_uploads.lesson_id`).

---

## Phase 2 — Failure Log

Tests run against production backend (`edupulse-production-477b.up.railway.app`).

| # | Test | Result | Severity | Bug? |
|---|---|---|---|---|
| T1 | Empty body POST /auth/login | 400 `invalid json body` | — | OK |
| T2 | Malformed JSON | 400 `invalid json body` | — | OK |
| T3 | Missing fields `{}` | 400 `email and password are required` | — | OK |
| T4 | Register with `email = "not-an-email"` | **201 Created** + JWT | **High** | YES: no email validation |
| T5 | Register with `password = "x"` (1 char) | **201 Created** + JWT | **High** | YES: no min length on server |
| T6 | GET /users/me without token | 401 `missing bearer token` | — | OK |
| T7 | Unknown route `/nonexistent` | 404 plain text `404 page not found` | Low | inconsistent format (not JSON) |
| T8 | Register with 50 KB password | **409** `bcrypt: password length exceeds 72 bytes` | **Medium** | YES: leaks internal lib name + wrong status code |
| T9 | Invalid role `superhacker` | 400 with allowed list | — | OK |
| T10 | GET on `/auth/login` (POST endpoint) | 405 | — | OK |
| T11 | 20 concurrent wrong-password logins | 20× 401, no throttle | **Medium** | YES: no rate limit / lockout |
| T12 | Duplicate email register | 409 `email already registered` | — | OK |
| T13 | SQL-injection-style email | 401 `invalid credentials` | — | OK (chi + parameterized queries) |
| T14 | Register with `email = "<script>alert(1)</script>"` | **201 Created** + JWT | **High** | YES: ties together with T4 — same root cause |

### Severity rationale
- **High**: lets bogus accounts onto the system, weakens the user-identity layer.
- **Medium**: enables internal-detail disclosure (T8) or password brute force (T11).
- **Low**: cosmetic API inconsistency.

### Assumed root causes
- T4/T5/T14: `Register` handler trims email and checks `IsValidRole` but never validates
  email format or password length. Frontend has these checks; backend trusts the client.
- T8: Service calls `bcrypt.GenerateFromPassword` which fails > 72 bytes; the error
  bubbles up unwrapped through `service.Register` → handler returns 409 with raw text.
- T11: No middleware tracks login attempts per IP/email. Each request just does a DB
  lookup + bcrypt compare — costly but unbounded.
- T7: chi's default 404 handler returns `http.NotFound` (text/plain), not the project's
  JSON error envelope.

---

## Phase 3 — Fix & Improve (before vs after)

### Fix #1 — Server-side email & password validation (`backend/internal/http/handlers/auth.go`)

**Before**
```go
req.Email = strings.TrimSpace(req.Email)
if req.Email == "" || req.Password == "" {
    writeError(w, http.StatusBadRequest, "email and password are required")
    return
}
if !auth.IsValidRole(req.Role) { ... }
```
- Accepts `"not-an-email"` → 201
- Accepts 1-char password → 201

**After**
```go
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
const minPasswordLen, maxPasswordLen = 6, 72

// in Register handler:
if !emailRegex.MatchString(req.Email) {
    writeError(w, http.StatusBadRequest, "invalid email format"); return
}
if len(req.Password) < minPasswordLen { ... } // 400 "at least 6 characters"
if len(req.Password) > maxPasswordLen { ... } // 400 "at most 72 characters"
```
- `"not-an-email"` → 400 `invalid email format`
- `"x"` → 400 `password must be at least 6 characters`
- 50 KB password → 400 `password must be at most 72 characters` (no bcrypt leak)

### Fix #2 — Login no longer leaks bcrypt internals (same file)

**Before**: 50 KB password on `/auth/login` would push bcrypt past its 72-byte limit
and produce an internal error in the log + a confusing response.

**After**: handler now rejects oversized passwords up-front with a generic
`401 invalid credentials`. The login response surface is identical to a real
wrong password — no oracle that reveals an account exists or that the limit was hit.

### UX improvement — distinguish network errors in `LoginView.vue`

**Before**
```ts
catch (e: any) {
  errorMsg.value = e.response?.data?.error || 'Login failed'
}
```
If the user has no internet or the backend is down, `e.response` is undefined and the
user sees a misleading `Login failed` — same message they'd see with wrong credentials.

**After**
```ts
if (e.response?.data?.error) {
  errorMsg.value = e.response.data.error            // 401 "invalid credentials"
} else if (e.code === 'ERR_NETWORK' || e.message === 'Network Error') {
  errorMsg.value = 'Cannot reach server. Check your connection and try again.'
} else {
  errorMsg.value = 'Login failed. Please try again.'
}
```
Now a flaky connection produces an actionable message instead of suggesting the
credentials were wrong.

### Verification (local)
```
F1 short password    → HTTP 400 "password must be at least 6 characters"
F2 invalid email     → HTTP 400 "invalid email format"
F3 50KB password     → HTTP 400 "password must be at most 72 characters"
F4 valid registration→ HTTP 201 (regression OK)
go build ./...       → exit 0
npm run type-check   → exit 0
```

---

## Phase 4 — System Thinking

### What breaks at scale?
- **SQLite is single-writer.** With more than a few concurrent
  `POST /sessions` / `/courses` / `/homework` requests, writes serialize and
  latency climbs. WAL mode would help but won't survive multi-instance Railway.
- **Single backend process owns the FFmpeg worker, the upload directory, and
  the DB.** Horizontal scaling is impossible without externalizing all three
  (object storage for media, Postgres for DB, queue for transcode jobs).
- **CORS `AllowedOrigins: ["*"]`** — fine for a beta but means any site can
  hit the API on behalf of a logged-in user via stolen JWT.

### Biggest bottleneck
Video transcoding. FFmpeg runs in-process on the same box that serves API
requests; a single 500 MB upload pegs CPU and the request loop slows down for
everyone. There is no queue and no rate limit on uploads.

### Weakest design part
Auth + identity. Until this lab, the backend would mint JWTs for `password = "x"`
and `email = "<script>"`. There is no email verification, no password reset,
no rate limiting on login, JWTs live for 24h and cannot be revoked. The frontend's
client-side checks created the illusion of safety.

### What would I rebuild?
1. **Move DB to Postgres on Railway** — instantly unblocks horizontal scaling
   and gets us real concurrency control + transactions across services.
2. **Move media to S3-compatible storage + a real job queue (Redis/RabbitMQ)**
   for transcoding. Backend stays stateless; transcoders scale independently.
3. **Auth hardening**: email verification, rate-limited login (per-IP and
   per-email), shorter access-token TTL with refresh tokens, server-side
   revocation list.
4. **Observability**: structured JSON logs + a `request_id` carried end-to-end,
   plus a `/metrics` endpoint for basic Prometheus counters.

---

## Phase 5 — Self-review (peer review substitute)

Three issues identified during this audit:

1. **Frontend trusts itself for validation.** The Register page enforces
   min-6 password and email format in the browser, but a `curl` request bypasses
   all of it. *(Fixed in Phase 3.)*
2. **No rate limit on `/auth/login`.** 20 parallel guesses returned 20× 401 with
   no slowdown. Brute force is essentially free.
3. **Errors are not normalized.** Most endpoints return `{"error":"..."}` but
   chi's default 404 returns plain text `404 page not found`. Frontend's
   `e.response?.data?.error` falls through to a generic message in that case.

Suggested improvement: add a chi `NotFound` handler that returns the project's
JSON error envelope, and add a simple `httprate.LimitByIP` (10 req/min) on the
`/auth` group. Both are < 30 lines and remove two real-world failure modes.
