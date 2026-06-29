# --- Stage 1: build ----------------------------------------------------
FROM golang:1.23-alpine AS build
WORKDIR /src

# Cache deps first
COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate

# --- Stage 2: runtime --------------------------------------------------
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata curl
WORKDIR /app
COPY --from=build /out/api /app/api
COPY --from=build /out/migrate /app/migrate
COPY backend/migrations /app/migrations
EXPOSE 8080
HEALTHCHECK --interval=10s --timeout=3s --retries=5 \
  CMD wget -qO- http://localhost:8080/api/v1/health >/dev/null 2>&1 || exit 1
USER nobody
ENTRYPOINT ["/app/api"]</content>
