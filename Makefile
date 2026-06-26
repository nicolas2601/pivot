.PHONY: help up down logs db-shell backend-install web-install install dev dev-backend dev-web

help:
	@echo "Targets:"
	@echo "  make up              - Levanta Postgres en Docker"
	@echo "  make down            - Baja Postgres"
	@echo "  make logs            - Logs de Postgres"
	@echo "  make db-shell        - psql dentro del contenedor"
	@echo "  make install         - Instala deps de backend + frontend"
	@echo "  make dev             - Levanta todo (backend + frontend + db)"
	@echo "  make dev-backend     - Solo backend"
	@echo "  make dev-web         - Solo frontend"

up:
	docker compose up -d
	@echo "Esperando a que Postgres esté listo..."
	@sleep 3
	@docker compose exec postgres pg_isready -U finanzas || echo "Postgres aún no listo, intenta de nuevo en unos segundos"

down:
	docker compose down

logs:
	docker compose logs -f postgres

db-shell:
	docker compose exec postgres psql -U finanzas -d finanzas

install: backend-install web-install

backend-install:
	cd backend && go mod download

web-install:
	cd web && pnpm install

dev:
	@echo "Levantando DB..."
	@make up
	@echo "Backend en :8080 y Web en :5173"
	@make -j2 dev-backend dev-web

dev-backend:
	cd backend && make run

dev-web:
	cd web && make dev