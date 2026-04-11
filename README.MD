# EduPulse

Educational platform project — backend (Go) + frontend (Vue 3).

## Project Structure

```
EduPulse/
├── backend/
│   ├── cmd/server/       # entry point
│   ├── internal/
│   │   ├── auth/         # JWT, password hashing, roles
│   │   ├── db/           # SQLite open, migrations, seed
│   │   ├── events/       # in-memory event bus
│   │   ├── http/         # router (chi), handlers
│   │   ├── middleware/    # JWT auth, RBAC
│   │   ├── repo/         # database repositories
│   │   ├── service/      # business logic
│   │   └── workers/      # async event consumers
│   ├── uploads/          # static images
│   ├── videos/           # static video files
│   ├── hls/              # HLS video streaming
│   ├── go.mod
│   └── edupulse.db       # SQLite (auto-created on first run)
└── frontend/             # Vue 3 + TypeScript SPA
    ├── src/
    │   ├── views/        # pages (Landing, Login, Register, Dashboard, etc.)
    │   ├── layouts/      # AuthenticatedLayout with sidebar
    │   ├── api/          # Axios client + all API functions
    │   ├── stores/       # Pinia auth store
    │   └── router/       # Vue Router with auth guards
    ├── package.json
    └── vite.config.ts
```

---

## Backend

Go + chi + SQLite REST API with JWT authentication and RBAC.

### How to run

```bash
cd backend
go run ./cmd/server
```

Server starts on `http://localhost:8080`. The SQLite database (`edupulse.db`) and all tables are created automatically on first launch.

### Environment variables

| Variable | Default | Description |
|---|---|---|
| `EDUPULSE_ADDR` | `:8080` | Listen address |
| `EDUPULSE_DB_PATH` | `./edupulse.db` | Path to SQLite file |
| `EDUPULSE_JWT_SECRET` | `dev-secret-change-me` | JWT signing secret |

### Seed data

On first run (empty DB) the server creates test users:

| Email | Password | Role |
|---|---|---|
| `admin@edupulse.local` | `adminpass` | admin |
| `manager@edupulse.local` | `managerpass` | manager |
| `teacher@edupulse.local` | `teacherpass` | teacher |
| `student@edupulse.local` | `studentpass` | student |
| `parent@edupulse.local` | `parentpass` | parent |

### API Endpoints

#### Public (no auth required)

| Method | Path | Description |
|---|---|---|
| `GET` | `/healthz` | Health check, returns `ok` |
| `POST` | `/auth/register` | Register a new user |
| `POST` | `/auth/login` | Login, returns JWT token |
| `GET` | `/uploads/*` | Static image files |
| `GET` | `/videos/*` | Static video files |
| `GET` | `/hls/*` | HLS video streaming |

#### Protected (JWT required)

| Method | Path | Roles | Description |
|---|---|---|---|
| `GET` | `/users/me` | any | Current user profile |
| `POST` | `/branches` | admin, manager | Create a branch |
| `GET` | `/branches` | any | List branches |
| `GET` | `/branches/{id}` | any | Get branch by ID |
| `POST` | `/sessions` | manager, teacher | Create a session |
| `GET` | `/sessions` | any | List sessions |
| `GET` | `/sessions/{id}` | any | Get session by ID |
| `POST` | `/homework/submit` | student | Submit homework |
| `GET` | `/homework` | teacher, manager, admin | List all homework |
| `GET` | `/homework/mine` | student | List own homework |
| `PATCH` | `/homework/{id}/status` | teacher | Update homework status |
| `GET` | `/courses` | any | List courses with lessons and reviews |
| `GET` | `/audit-logs` | admin | Audit log |
| `GET` | `/notifications` | admin, manager | Notifications |

### Auth examples

**Register:**
```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"secret","role":"student"}'
```

Response: `{"token":"eyJ...","role":"student"}`

**Login:**
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@edupulse.local","password":"adminpass"}'
```

Response: `{"token":"eyJ...","role":"admin"}`

**Use token in requests:**
```bash
curl http://localhost:8080/users/me \
  -H "Authorization: Bearer <token>"
```

### Features

- **JWT Authentication** — register & login, 24h token expiry
- **Role-Based Access Control** — 5 roles: admin, manager, teacher, student, parent
- **Branches** — CRUD for physical locations (lat/lng)
- **Sessions** — teaching sessions linked to branches and teachers
- **Homework** — students submit, teachers approve/reject
- **Courses** — courses with lessons and reviews
- **Audit logging** — tracks key actions (create branch, create session, etc.)
- **Notifications** — event-driven (homework submit triggers notification)
- **Static file serving** — images (`/uploads`), videos (`/videos`), HLS streaming (`/hls`)
- **CORS** — enabled for all origins
- **Auto-migration** — tables created on startup, no manual SQL needed

### Tech stack

- **Go** — standard library + chi router
- **SQLite** — via `modernc.org/sqlite` (pure Go, no CGO)
- **JWT** — `golang-jwt/jwt/v5`
- **Passwords** — bcrypt via `golang.org/x/crypto`
- **CORS** — `go-chi/cors`

---

## Frontend

Vue 3 + TypeScript + Tailwind CSS v4 + Pinia + Vue Router.

### How to run

```bash
cd frontend
npm install
npm run dev
```

Opens on `http://localhost:5173`. The Vite dev server proxies `/api` → `localhost:8080`, so the backend needs to be running.

### Features

- Landing page with animated network background
- Login / Register with form validation
- Google & Apple sign-in buttons (UI only, OAuth not wired yet)
- Dashboard with stats (branches count, sessions count, notifications)
- Branches list + create modal
- Sessions list + create modal
- Homework list, submit (student), approve/reject (teacher)
- Notifications & Audit logs pages (role-gated)
- Profile page with logout
- Responsive sidebar layout (desktop + mobile)
- Dark theme (#121212 bg, cyan accent, Nunito font)
