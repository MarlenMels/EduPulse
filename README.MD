# EduPulse

Educational platform project — backend (Go) + frontend (Vue 3).

## Project Structure

```
EduPulse/
├── backend/        # Go REST API
│   ├── cmd/server/ # entry point
│   ├── internal/   # handlers, services, repos, middleware, auth
│   ├── go.mod
│   └── edupulse.db # SQLite database (auto-created on first run)
└── frontend/       # Vue 3 + TypeScript SPA
    ├── src/
    │   ├── views/      # pages (Landing, Login, Register, Dashboard, etc.)
    │   ├── layouts/    # AuthenticatedLayout with sidebar
    │   ├── api/        # Axios client + all API functions
    │   ├── stores/     # Pinia auth store
    │   └── router/     # Vue Router with auth guards
    ├── package.json
    └── vite.config.ts
```

## Backend

Go + SQLite REST API with JWT authentication.

**What's implemented:**
- JWT auth with login & registration (`/auth/login`, `/auth/register`)
- Role-based access control — admin, manager, teacher, student, parent
- CRUD endpoints for branches, sessions, homework
- Homework status management (pending → approved / rejected)
- Audit logging for key actions
- Notifications (event-driven: homework submit → notification created)
- Auto-migration on startup (creates all tables if they don't exist)

**How to run:**
```bash
cd backend
go run ./cmd/server
# listening on :8080
```

## Frontend

Vue 3 + TypeScript + Tailwind CSS v4 + Pinia + Vue Router.

**What's implemented:**
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

**How to run:**
```bash
cd frontend
npm install
npm run dev
# opens on http://localhost:5173
```

The Vite dev server proxies `/api` → `localhost:8080` so the backend needs to be running.

## Note About Git

The frontend was originally in a separate repo, so it had its own `.git` folder.
That caused git to treat it as a submodule, which is why it couldn't be committed
and pushed together with the backend. I removed the nested `.git` from `frontend/`
so now everything lives in one repo and can be pushed normally.

The `.gitignore` at the root excludes:
- `backend/edupulse.db` (local database, regenerated on startup)
- `frontend/node_modules/` (install via `npm install`)
- `frontend/dist/` and `frontend/.vite/` (build artifacts)
