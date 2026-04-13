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
| `EDUPULSE_DB_PATH` | `./edupulse.db` | Path to SQLite file |
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

## Tech Stack

| Layer | Technologies |
|---|---|
| **Backend** | Go, Chi Router, SQLite (modernc.org/sqlite), JWT, bcrypt, Swaggo |
| **Frontend** | Vue 3, TypeScript, Tailwind CSS v4, Pinia, Vue Router, Axios |
| **Infrastructure** | Vite, HLS Streaming, Event Bus |

## Project Structure

```
EduPulse/
├── backend/
│   ├── cmd/server/        # entry point
│   ├── docs/              # Swagger (auto-generated)
│   ├── internal/
│   │   ├── auth/          # JWT, password hashing, roles
│   │   ├── db/            # SQLite, migrations, seed
│   │   ├── events/        # event bus
│   │   ├── http/          # router (chi), handlers
│   │   ├── middleware/     # JWT auth, RBAC
│   │   ├── repo/          # repositories
│   │   ├── service/       # business logic
│   │   └── workers/       # event consumers
│   ├── uploads/           # images
│   ├── videos/            # video files
│   └── hls/               # HLS streaming
└── frontend/
    └── src/
        ├── views/         # pages
        ├── layouts/       # layout with sidebar
        ├── api/           # Axios client
        ├── stores/        # Pinia (auth store)
        └── router/        # Vue Router + guards
```
