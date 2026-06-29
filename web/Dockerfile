# --- Stage 1: build ----------------------------------------------------
FROM node:20-alpine AS build
WORKDIR /app

# Cache deps
COPY web/package.json web/pnpm-lock.yaml ./
RUN corepack enable && corepack prepare pnpm@10 --activate
RUN pnpm install --frozen-lockfile

COPY web/ ./
RUN pnpm build

# --- Stage 2: serve (static preview) ----------------------------------
# Use SvelteKit's `adapter-static` output served by `serve` for offline.
# If/when we switch to adapter-node this becomes `node ./build/index.js`.
FROM node:20-alpine
WORKDIR /app
RUN corepack enable && corepack prepare pnpm@10 --activate
COPY --from=build /app /app
EXPOSE 4173
HEALTHCHECK --interval=10s --timeout=3s --retries=5 \
  CMD wget -qO- http://localhost:4173/ >/dev/null 2>&1 || exit 1
USER node
CMD ["pnpm", "preview", "--host", "0.0.0.0", "--port", "4173"]</content>
