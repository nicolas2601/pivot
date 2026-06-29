# Minimaax

App personal de finanzas: gastos, presupuestos, metas, viajes en grupo. Backend Go + Gin, frontend SvelteKit + Svelte 5, Postgres 16, Docker.

## Stack

| Capa        | Tech                                                                  |
|-------------|-----------------------------------------------------------------------|
| Backend     | Go 1.23 · Gin · GORM · golang-jwt/jwt · bcrypt · PostgreSQL             |
| Frontend    | SvelteKit 2 · Svelte 5 (runes) · Tailwind v4 · TanStack Query · Zod   |
| PWA         | vite-plugin-pwa · workbox (autoUpdate SW, runtime cache)             |
| CI          | GitHub Actions (lint + tests + build + Playwright E2E)               |
| Container   | Docker + docker-compose (db · backend · frontend)                    |

## Quick start

```bash
# 1) Clonar y copiar .env
git clone <repo>
cp .env.example .env

# 2) Levantar full stack con Docker
make up
# → Postgres en :5432
# → Backend   en :8080
# → Frontend  en :4173 (SPA + SW precache)

# 3) o bien: DB en Docker + procesos locales (live-reload)
make dev
# → Postgres en :5432
# → Backend en :8080 (con hot reload via go run)
# → Frontend en :5173 (Vite dev server)
```

## Tests

```bash
make test            # backend (go test -race) + frontend (vitest)
make e2e-install     # solo la primera vez: playwright + chromium
make e2e             # Playwright contra el stack arriba
```

## Estructura

```
.
├── backend/         Go API (Gin, GORM, JWT, bcrypt)
│   ├── cmd/api/     entrypoint
│   ├── internal/    paquetes (auth, accounts, transactions, etc.)
│   ├── migrations/  SQL versionado
│   └── Dockerfile   multi-stage build (alpine, sin CGO)
├── web/             SvelteKit + Svelte 5
│   ├── src/
│   │   ├── routes/      file-based routing (auth/, app, reports, accounts, ...)
│   │   ├── lib/         api/, stores/, components/, schemas/, motion/
│   │   ├── app.html     shell con theme-color + manifest
│   │   └── app.css      tokens Tailwind + motion tokens
│   ├── e2e/         Playwright specs (auth, accounts, reports, smoke)
│   ├── static/      manifest.json, favicon
│   └── Dockerfile
├── docs/            documentación local
├── docker-compose.yml
├── Makefile         orquestador (up/dev/test/e2e/clean)
└── .github/workflows/ci.yml
```

## Variables de entorno

| Var                       | Ejemplo                                                          |
|---------------------------|------------------------------------------------------------------|
| `POSTGRES_USER/PASSWORD/DB` | `finanzas` / `finanzas_dev_password` / `finanzas`               |
| `DATABASE_URL`            | `postgres://finanzas:finanzas_dev_password@db:5432/finanzas?sslmode=disable` |
| `JWT_SECRET`              | ≥64 chars aleatorios en prod                                    |
| `PUBLIC_API_URL`          | URL absoluta para el frontend → API                              |

## QA handoff

Ver [docs/qa-handoff.md](docs/qa-handoff.md).