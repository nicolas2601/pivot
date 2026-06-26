# Finanzas Personales

App personal de finanzas tipo PWA. Spec en `docs/superpowers/specs/`.

## Setup local

Requisitos: Go 1.23+, Node 20+, pnpm 9+, Docker.

```bash
make install
make up
make dev
```

- Frontend: http://localhost:5173
- Backend: http://localhost:8080
- Postgres: localhost:5432

## Estructura

- `backend/` — Go API (Gin + GORM)
- `web/` — SvelteKit PWA
- `docs/` — Specs y planes