# EduPulse (Go + SQLite) — minimal backend for university assignment

A small but *real* REST API backend in Go for an educational center platform **EduPulse**.

Key features implemented (per assignment):
- REST API (8 endpoints)
- JWT auth + RBAC middleware (admin/manager/teacher/student/parent)
- Event-driven simulation: `homework_submitted` published on submit, consumed by worker -> creates `notifications` row + system audit log
- Audit logging: `create_session`, `submit_homework` (plus a couple extra)
- H3 geo-indexing: compute + store `h3_index` on branches/sessions; filter sessions by H3; aggregate sessions by `h3_index/day`
- SQLite with SQL migration + auto-migration on startup
- Modular layout: handlers/services/repos/middleware/events/db/auth

## Project layout
