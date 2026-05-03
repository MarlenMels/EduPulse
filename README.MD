# EduPulse — Education Center Management Platform

Web application for managing an education center: class scheduling, homework submissions, course catalog, notifications, and audit logging.

## Problem

Education centers manage schedules, homework, and reporting manually — through spreadsheets, messengers, and paper. This leads to lost data, status confusion, and lack of oversight. EduPulse brings everything into one place with role-based access control.

## Features

- **Authentication** — register and login via email/password, JWT tokens (24h expiry)
- **5 Roles** — admin, manager, teacher, student, parent with RBAC
- **Class Sessions** — create and view sessions linked to teachers
- **Homework** — students submit, teachers review and update status (approved/rejected)
- **Course Catalog** — courses with lessons and reviews
- **Notifications** — automatic events triggered on homework submission (event bus)
- **Audit Logging** — tracks all key actions in the system
- **Media Files** — serves images, videos, and HLS streaming
- **Swagger UI** — interactive API documentation at `/swagger/index.html`
- **Auto-Migration** — database and tables are created automatically on startup

## Installation & Setup

### Requirements

- Go 1.25+
- Node.js 18+

### Backend

```bash
cd backend
go run ./cmd/server
```

Server starts at `http://localhost:8080`. The SQLite database is created automatically.

**Environment variables:**

| Variable | Default | Description |
|---|---|---|
| `EDUPULSE_ADDR` | `:8080` | Server address |
| `EDUPULSE_DB_URL` | _(empty)_ | Postgres connection string. When set, the backend uses Postgres instead of SQLite. Example: `postgres://user:pass@host:5432/dbname?sslmode=require` |
| `EDUPULSE_DB_PATH` | `./edupulse.db` | Path to SQLite file. Used only when `EDUPULSE_DB_URL` is empty. |
| `EDUPULSE_JWT_SECRET` | `dev-secret-change-me` | JWT signing secret |

**Test users** (created on first run):

| Email | Password | Role |
|---|---|---|
| `admin@edupulse.local` | `adminpass` | admin |
| `manager@edupulse.local` | `managerpass` | manager |
| `teacher@edupulse.local` | `teacherpass` | teacher |
| `student@edupulse.local` | `studentpass` | student |
| `parent@edupulse.local` | `parentpass` | parent |

**Swagger UI:** after starting the server, open `http://localhost:8080/swagger/index.html`

### Frontend

```bash
cd frontend
npm install
npm run dev
```

Opens at `http://localhost:5173`. The Vite dev server proxies API requests to the backend (`localhost:8080`).

## Production Deployment

The backend runs on Railway, the frontend on Vercel.
[text](https://edu-pulse-ashen.vercel.app/)

### Persisting data on Railway (SQLite)

Railway containers have an **ephemeral filesystem** — every redeploy wipes the SQLite file, all uploaded videos, and HLS chunks. Use a Railway Volume so the data survives.

1. Railway → service → **Settings → Volumes → New Volume**, mount path: `/data`
2. Add environment variables on the same service:
   - `EDUPULSE_DB_PATH=/data/edupulse.db`
3. Redeploy.

> The `videos/`, `hls/`, and `uploads/` directories also currently live on the local filesystem. Until they are moved to object storage (e.g. S3), put them on the same Volume by changing the working directory or symlinking inside `/data`.

### Switching to Postgres (recommended for production)

For more than one instance, real backups, and concurrent writes, use Postgres.

1. Railway → project → **+ New → Database → Postgres**. Railway exposes `DATABASE_URL`.
2. On the backend service set:
   - `EDUPULSE_DB_URL` = the Postgres URL (copy it from the Postgres service variables)
3. Redeploy. On startup the backend detects the URL, runs the Postgres migrations from `internal/db/migrations_pg/`, and seeds the same default users.

The backend supports either driver — when `EDUPULSE_DB_URL` is empty it falls back to SQLite (`EDUPULSE_DB_PATH`). Local development can stay on SQLite without any code change.

## Tech Stack

| Layer | Technologies |
|---|---|
| **Backend** | Go, Chi Router, SQLite (modernc.org/sqlite) or Postgres (pgx/v5), JWT, bcrypt, Swaggo, File Upload API |
| **Frontend** | Vue 3, TypeScript, Tailwind CSS v4, Pinia, Vue Router, Axios, Vercel Blob Upload |
| **Infrastructure** | Vite, HLS Streaming, Event Bus, Railway (Backend), Vercel (Frontend) |

## Project Structure

```
EduPulse/
├── backend/
│   ├── cmd/server/        # entry point
│   ├── docs/              # Swagger (auto-generated)
│   ├── internal/
│   │   ├── auth/          # JWT, password hashing, roles
│   │   ├── db/            # SQLite, migrations, seed (SQLite + PostgreSQL)
│   │   ├── events/        # event bus
│   │   ├── http/          # router (chi), handlers (homework, sessions, uploads)
│   │   ├── middleware/     # JWT auth, RBAC
│   │   ├── repo/          # repositories (homework with attachments)
│   │   ├── service/       # business logic
│   │   └── workers/       # event consumers
│   ├── uploads/           # images
│   ├── videos/            # video files
│   └── hls/               # HLS streaming
└── frontend/
    └── src/
        ├── views/         # pages (HomeworkView with file upload, SessionDetailView)
        ├── layouts/       # layout with sidebar
        ├── api/           # Axios client, Vercel Blob Upload
        ├── stores/        # Pinia (auth store)
        ├── router/        # Vue Router + guards
        └── components/    # VideoUploader
```

## Student IDs

230103109 - Madi Zhazykenov
230109011 - Yernar Bekbolatuly
230103024 - Marlen Mels
230103369 - Alisher Kudaibergen