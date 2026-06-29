# Minimaax Makefile — orchestrates the full stack.
# Conventions:
#   make up              → full stack (db + backend + frontend) on Docker
#   make dev             → DB in Docker, backend + frontend on host (live reload)
#   make test            → backend + frontend unit tests
#   make e2e             → Playwright against the full Docker stack
#   make seed            → creates seeded categories for a demo user

.PHONY: help up down logs restart ps db-shell backend-install web-install \
        install dev dev-backend dev-web build build-backend build-web \
        test test-backend test-web e2e e2e-install clean

help:
	@echo "Targets:"
	@echo "  make up              - Levanta DB + backend + frontend (Docker, fresh)"
	@echo "  make dev             - Levanta DB en Docker y backend/frontend en host (live reload)"
	@echo "  make down            - Baja todo y mantiene volúmenes"
	@echo "  make logs            - Logs en vivo de todos los servicios"
	@echo "  make restart         - Rebuild + up"
	@echo "  make ps              - Estado de contenedores"
	@echo "  make db-shell        - psql dentro del contenedor de Postgres"
	@echo "  make install         - Instala deps backend (go mod) + web (pnpm)"
	@echo "  make build           - Compila binarios locales"
	@echo "  make test            - Tests unitarios backend + frontend"
	@echo "  make e2e-install     - Instala binarios de Playwright (chromium)"
	@echo "  make e2e             - Corre Playwright contra el stack arriba"
	@echo "  make clean           - Limpia artefactos locales"

up:
	docker compose --env-file .env up -d --build
	@echo "Esperando a Postgres..."
	@for i in 1 2 3 4 5 6 7 8 9 10; do \
	  docker compose exec -T postgres pg_isready -U finanzas -d finanzas >/dev/null 2>&1 && break; \
	  sleep 2; \
	done
	@docker compose ps

dev: dev-db
	@echo "Backend en :8080 y Web en :5173"
	@make -j2 dev-backend dev-web

dev-db:
	docker compose up -d postgres
	@for i in 1 2 3 4 5; do \
	  docker compose exec -T postgres pg_isready -U finanzas -d finanzas >/dev/null 2>&1 && break; \
	  sleep 2; \
	done

down:
	docker compose down

logs:
	docker compose logs -f

restart:
	docker compose down
	@make up

ps:
	docker compose ps

db-shell:
	docker compose exec postgres psql -U finanzas -d finanzas

install: backend-install web-install

backend-install:
	cd backend && go mod download

web-install:
	cd web && corepack enable && pnpm install --frozen-lockfile

dev-backend:
	@if [ ! -f backend/.env ]; then cp backend/.env.example backend/.env && echo "Creado backend/.env"; fi
	@cd backend && DATABASE_URL=postgres://finanzas:finanzas_dev_password@localhost:5432/finanzas?sslmode=disable go run ./cmd/api

dev-web:
	@cd web && pnpm dev

build: build-backend build-web

build-backend:
	cd backend && go build -o ./bin/api ./cmd/api

build-web:
	cd web && pnpm build

test: test-backend test-web

test-backend:
	cd backend && go test ./... -race -count=1

test-web:
	cd web && pnpm test

e2e-install:
	cd web && pnpm exec playwright install --with-deps chromium

e2e:
	cd web && PLAYWRIGHT_BASE_URL=http://localhost:4173 pnpm exec playwright test

clean:
	rm -rf backend/bin web/.svelte-kit web/build web/test-results web/playwright-report
	@echo "Artifacts cleaned."