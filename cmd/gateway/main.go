// Command gateway is the HTTP/WebSocket server for the fog-optimised cloud shell.
package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"

	"github.com/SocioProphet/cloudshell-fog/internal/api"
	"github.com/SocioProphet/cloudshell-fog/internal/audit"
	"github.com/SocioProphet/cloudshell-fog/internal/auth"
	"github.com/SocioProphet/cloudshell-fog/internal/connector"
	fogotel "github.com/SocioProphet/cloudshell-fog/internal/otel"
	"github.com/SocioProphet/cloudshell-fog/internal/placement"
	"github.com/SocioProphet/cloudshell-fog/internal/policy"
	"github.com/SocioProphet/cloudshell-fog/internal/pty"
	"github.com/SocioProphet/cloudshell-fog/internal/session"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// ── OpenTelemetry ──────────────────────────────────────────────────────────
	otelProviders, err := fogotel.Setup(ctx, "cloudshell-gateway")
	if err != nil {
		return fmt.Errorf("setup otel: %w", err)
	}
	defer otelProviders.Shutdown(context.Background()) //nolint:errcheck

	// ── Audit ─────────────────────────────────────────────────────────────────
	emitter := audit.NewLogEmitter(logger)

	// ── Session store ─────────────────────────────────────────────────────────
	store := session.NewInMemoryStore()

	// ── Policy ────────────────────────────────────────────────────────────────
	policyCfgPath := envOrDefault("POLICY_CONFIG", "config/policy.yaml")
	policyCfg, err := policy.LoadConfig(policyCfgPath)
	if err != nil {
		return fmt.Errorf("load policy config: %w", err)
	}
	policyEngine := policy.NewEngine(policyCfg)

	// ── Placement ─────────────────────────────────────────────────────────────
	cloudFallback := envOrDefault("CLOUD_FALLBACK_REGION", "us-east-1")
	registry := placement.NewRegistry()
	// Seed a default cloud-fallback node so the engine always has at least one candidate.
	registry.Upsert(placement.Node{
		ID:           "cloud-default",
		Region:       cloudFallback,
		Tier:         placement.TrustTierCloud,
		Health:       placement.NodeHealthHealthy,
		CapacityFree: 1.0,
		LatencyMS:    50,
		UpdatedAt:    time.Now(),
	})
	placementEngine := placement.NewEngine(registry, cloudFallback)

	// ── Runtime connector ─────────────────────────────────────────────────────
	conn, connMode, err := buildConnector(logger)
	if err != nil {
		return fmt.Errorf("build connector: %w", err)
	}
	_ = connMode // already logged inside buildConnector

	// ── Session TTL sweeper ───────────────────────────────────────────────────
	sweeper := session.NewSweeper(store, func(sCtx context.Context, s *session.Session) {
		ref := connector.RuntimeRef{
			ID:        s.ID,
			Namespace: "cloudshell-" + s.ID,
			PodName:   "shell",
			NodeID:    s.Placement,
		}
		if err := conn.Terminate(sCtx, ref); err != nil {
			logger.WarnContext(sCtx, "terminate expired session", "session_id", s.ID, "err", err)
		}
		emitter.Emit(sCtx, audit.Event{
			TS:        time.Now(),
			SessionID: s.ID,
			Subject:   s.Subject,
			Type:      audit.EventSessionTerminated,
			Placement: s.Placement,
			Details:   map[string]any{"reason": "ttl-expired"},
		})
	}, 30*time.Second)
	go sweeper.Run(ctx)

	// ── Session token minter ──────────────────────────────────────────────────
	signingKey, err := resolveSigningKey()
	if err != nil {
		return fmt.Errorf("resolve signing key: %w", err)
	}
	minter := auth.NewSessionTokenMinter(signingKey, 15*time.Minute, "cloudshell-gateway")

	// ── HTTP router ───────────────────────────────────────────────────────────
	gatewayURL := envOrDefault("GATEWAY_URL", "ws://localhost:8080")
	apiHandler := api.NewHandler(store, policyEngine, placementEngine, conn, minter, emitter, gatewayURL)
	ptyHandler := pty.NewHandler(minter, store, conn, emitter)

	r := mux.NewRouter()

	// Auth middleware: real OIDC when configured, dev shim otherwise.
	withAuth, err := buildAuthMiddleware(ctx, logger)
	if err != nil {
		return fmt.Errorf("build auth middleware: %w", err)
	}

	// Session management endpoints (OIDC-protected).
	r.Handle("/v1/sessions", withAuth(http.HandlerFunc(apiHandler.CreateSession))).Methods(http.MethodPost)
	r.Handle("/v1/sessions/{id}", withAuth(http.HandlerFunc(apiHandler.GetSession))).Methods(http.MethodGet)
	r.Handle("/v1/sessions/{id}", withAuth(http.HandlerFunc(apiHandler.DeleteSession))).Methods(http.MethodDelete)

	// PTY WebSocket (authenticated via short-lived session token in query param).
	r.HandleFunc("/v1/sessions/{id}/pty", ptyHandler.ServeHTTP).Methods(http.MethodGet)

	// Static web UI.
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("web/public")))

	addr := envOrDefault("LISTEN_ADDR", ":8080")
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 0, // disabled for long-lived WebSocket connections
		IdleTimeout:  60 * time.Second,
	}
	logger.Info("starting cloudshell gateway", "addr", addr)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "err", err)
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down")

	shutCtx, shutCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutCancel()
	return srv.Shutdown(shutCtx)
}

// buildAuthMiddleware returns a real OIDC middleware when OIDC_ISSUER_URL and
// OIDC_CLIENT_ID are set, otherwise a dev shim that injects a fixed subject.
func buildAuthMiddleware(ctx context.Context, logger *slog.Logger) (func(http.Handler) http.Handler, error) {
	issuer := os.Getenv("OIDC_ISSUER_URL")
	clientID := os.Getenv("OIDC_CLIENT_ID")
	if issuer != "" && clientID != "" {
		validator, err := auth.NewOIDCValidator(ctx, issuer, clientID)
		if err != nil {
			return nil, fmt.Errorf("create oidc validator: %w", err)
		}
		logger.Info("OIDC auth enabled", "issuer", issuer)
		return auth.Middleware(validator), nil
	}
	logger.Warn("OIDC not configured — using dev auth shim (dev-user / developers)")
	return devAuthMiddleware, nil
}

// devAuthMiddleware injects a fixed subject and group for local development.
func devAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), auth.ContextKeySubject, "dev-user")
		ctx = context.WithValue(ctx, auth.ContextKeyGroups, []string{"developers"})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func resolveSigningKey() ([]byte, error) {
	if envKey := os.Getenv("SESSION_TOKEN_SIGNING_KEY"); envKey != "" {
		return []byte(envKey), nil
	}
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("generate random signing key: %w", err)
	}
	return key, nil
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
