# QA handoff — Minimaax v1.0

## Estado actual

Backend Go (Gin) + frontend SvelteKit + Postgres. Tests:
- Backend: 88 tests (`go test -race`).
- Frontend: 137 tests (vitest).
- E2E: Playwright con 4 specs contra el stack completo (auth, accounts,
  reports, smoke).

## Pre-requisitos QA

1. Docker + docker compose v2.
2. Node 20 + pnpm 10 (corenpack).
3. Go 1.23 (solo si necesitas correr backend nativo).

## Setup en 60 segundos

```bash
# 1) clonar
cd /path/to/repo
cp .env.example .env       # ajustar JWT_SECRET en prod

# 2) levantar todo
make up
# Esperado en 60s:
# - postgres en localhost:5432 (con tests de pg_isready)
# - backend  en localhost:8080
# - frontend en localhost:4173 (SPA + service worker)

# 3) smoke test
curl -sf http://localhost:8080/api/v1/health
# → 200
open http://localhost:4173
```

## Smoke checklist manual

- [ ] `/auth/register` — crear cuenta; redirige a dashboard.
- [ ] `/accounts` — crear cuenta, ver saldo, eliminar.
- [ ] `/categories` — sembrar categorías por defecto, ver lista, filtrar
  por tipo (gasto / ingreso).
- [ ] `/transactions` — crear gasto, ingreso, transferencia.
- [ ] `/recurring` — crear regla diaria/semanal/mensual; verificar que
  se aplique hoy si coincide.
- [ ] `/goals` — crear meta, hacer depósito, ver progreso.
- [ ] `/travel` — crear grupo, añadir gastos, ver liquidaciones.
- [ ] `/reports` — cambiar período (1/3/6/12 meses), ver charts.
- [ ] `/budgets` — crear presupuesto, ver vs actual.
- [ ] Toast: ver success en mutaciones exitosas, error en fallos.
- [ ] Página 404: navegar a `/foo-bar`.
- [ ] PWA: devtools → Application → Manifest muestra icono + name.
- [ ] PWA: Application → Service Workers → sw.js registrado.

## Tests automatizados

```bash
make test           # backend + frontend unit
make e2e-install    # solo la primera vez
make e2e            # Playwright contra el stack arriba
```

## Bugs conocidos / edge cases

- **Timezone**: la API trata fechas como UTC; el frontend usa
  `America/Bogota`. En el límite del día se puede ver "14 de ene" para
  un evento 2026-01-15T00:00:00Z. Workaround: el frontend pide siempre
  T12:00:00Z para evitar esto.
- **Saldo total**: la card del dashboard muestra `opening_balance`. No
  hay agregación de transacciones porque la API no expone el campo
  derivado `current_balance`. TODO: agregar campo derivado o endpoint
  `/accounts/with-balance`.
- **PWA**: el icono 192/512 son placeholders. Reemplazar con PNG reales
  antes de release. La build usa `--public` dist vacío porque
  adapter-auto no detecta plataforma; cambiar a adapter-static si se
  quiere publicar.

## Performance budgets

- LCP dashboard ≤ 2.5s en 3G (medir con WebPageTest).
- JS bundle ≤ 250 KiB gzipped (vite visualizer).
- Backend p95 latency ≤ 200ms en endpoints `/api/v1/*`.

## Security

- Auth: bcrypt(12), JWT(15m) + refresh(30d), HTTP-only cookies SameSite=Lax.
- Validación de input: Zod en frontend, struct tags + middleware en backend.
- Headers: `GIN_MODE=release` desactiva debug output.

## Roadmap para v2

- Settings page (cambiar moneda principal).
- Profile (avatar, display name).
- Onboarding 4-step tour.
- Mobile gestures (swipe to delete, pull-to-refresh).
- Real-time sync (WebSockets or Server-Sent Events).
- Export a CSV / PDF.
- Apple/Google sign-in.

## Contacto / escalation

- Repo: https://github.com/<org>/minimaax
- Slack: #finanzas-app
- Pager: on-call rotation visible en PagerDuty.