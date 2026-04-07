# ── Stage 1: Build the Go gateway binary ──────────────────────────────────────
FROM golang:1.22-bookworm AS go-build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /gateway ./cmd/gateway

# ── Stage 2: Build the web UI ─────────────────────────────────────────────────
FROM node:20-bookworm-slim AS frontend-build
WORKDIR /web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# ── Stage 3: Minimal runtime image ────────────────────────────────────────────
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app

COPY --from=go-build /gateway ./gateway
COPY --from=frontend-build /web/public ./web/public
COPY config/ ./config/

LABEL org.opencontainers.image.title="cloudshell-fog" \
      org.opencontainers.image.description="Fog-optimised cloud shell gateway with OIDC, placement engine, and WebSocket PTY" \
      org.opencontainers.image.url="https://github.com/SocioProphet/cloudshell-fog" \
      org.opencontainers.image.source="https://github.com/SocioProphet/cloudshell-fog" \
      org.opencontainers.image.documentation="https://github.com/SocioProphet/cloudshell-fog/blob/main/README.md" \
      org.opencontainers.image.licenses="Apache-2.0" \
      org.opencontainers.image.vendor="SocioProphet"

EXPOSE 8080
ENTRYPOINT ["/app/gateway"]
