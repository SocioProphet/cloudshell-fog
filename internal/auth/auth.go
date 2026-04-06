// Package auth provides OIDC token validation and short-lived session token minting.
package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	gooidc "github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v5"
)

// contextKey is an unexported type for context keys defined in this package.
type contextKey string

// ContextKeySubject and ContextKeyGroups are context keys set by the auth middleware.
const (
	ContextKeySubject contextKey = "subject"
	ContextKeyGroups  contextKey = "groups"
)

// Claims holds data extracted from a validated OIDC access token.
type Claims struct {
	Subject string
	Groups  []string
}

// OIDCValidator validates OIDC access tokens using the provider's JWKS.
type OIDCValidator struct {
	verifier *gooidc.IDTokenVerifier
}

// NewOIDCValidator creates a validator by fetching OIDC discovery metadata from issuerURL.
func NewOIDCValidator(ctx context.Context, issuerURL, clientID string) (*OIDCValidator, error) {
	provider, err := gooidc.NewProvider(ctx, issuerURL)
	if err != nil {
		return nil, fmt.Errorf("create oidc provider for %q: %w", issuerURL, err)
	}
	verifier := provider.Verifier(&gooidc.Config{ClientID: clientID})
	return &OIDCValidator{verifier: verifier}, nil
}

// Validate verifies rawToken and returns the extracted Claims.
func (v *OIDCValidator) Validate(ctx context.Context, rawToken string) (*Claims, error) {
	token, err := v.verifier.Verify(ctx, rawToken)
	if err != nil {
		return nil, fmt.Errorf("verify token: %w", err)
	}

	var extra struct {
		Groups []string `json:"groups"`
		Roles  []string `json:"roles"`
	}
	if err := token.Claims(&extra); err != nil {
		return nil, fmt.Errorf("extract extra claims: %w", err)
	}

	groups := append(extra.Groups, extra.Roles...)
	return &Claims{Subject: token.Subject, Groups: groups}, nil
}

// Middleware returns an HTTP middleware that validates Bearer OIDC tokens.
// On success it injects subject and groups into the request context.
func Middleware(validator *OIDCValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw, err := extractBearer(r)
			if err != nil {
				http.Error(w, "unauthorized: "+err.Error(), http.StatusUnauthorized)
				return
			}
			claims, err := validator.Validate(r.Context(), raw)
			if err != nil {
				http.Error(w, "unauthorized: "+err.Error(), http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), ContextKeySubject, claims.Subject)
			ctx = context.WithValue(ctx, ContextKeyGroups, claims.Groups)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractBearer(r *http.Request) (string, error) {
	h := r.Header.Get("Authorization")
	if h == "" {
		return "", errors.New("missing Authorization header")
	}
	parts := strings.SplitN(h, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return "", errors.New("invalid Authorization header; expected 'Bearer <token>'")
	}
	return parts[1], nil
}

// SubjectFromContext extracts the OIDC subject from ctx.
func SubjectFromContext(ctx context.Context) string {
	v, _ := ctx.Value(ContextKeySubject).(string)
	return v
}

// GroupsFromContext extracts the groups/roles from ctx.
func GroupsFromContext(ctx context.Context) []string {
	v, _ := ctx.Value(ContextKeyGroups).([]string)
	return v
}

// SessionTokenClaims are the JWT claims for a short-lived, session-bound PTY token.
type SessionTokenClaims struct {
	SessionID string `json:"session_id"`
	jwt.RegisteredClaims
}

// SessionTokenMinter mints and validates short-lived session JWTs.
// These tokens are scoped to a single session and cannot be used to create new sessions.
type SessionTokenMinter struct {
	signingKey []byte
	ttl        time.Duration
	issuer     string
}

// NewSessionTokenMinter returns a minter using HMAC-SHA256.
func NewSessionTokenMinter(signingKey []byte, ttl time.Duration, issuer string) *SessionTokenMinter {
	return &SessionTokenMinter{signingKey: signingKey, ttl: ttl, issuer: issuer}
}

// Mint issues a signed session token for (sessionID, subject).
// Returns the signed token string and its expiry time.
func (m *SessionTokenMinter) Mint(sessionID, subject string) (string, time.Time, error) {
	expiresAt := time.Now().Add(m.ttl)
	claims := SessionTokenClaims{
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   subject,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			Audience:  jwt.ClaimStrings{"cloudshell-pty"},
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.signingKey)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign session token: %w", err)
	}
	return signed, expiresAt, nil
}

// Validate parses and verifies a session token, returning its claims on success.
func (m *SessionTokenMinter) Validate(raw string) (*SessionTokenClaims, error) {
	token, err := jwt.ParseWithClaims(raw, &SessionTokenClaims{},
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return m.signingKey, nil
		},
		jwt.WithAudience("cloudshell-pty"),
		jwt.WithIssuer(m.issuer),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return nil, fmt.Errorf("parse session token: %w", err)
	}
	claims, ok := token.Claims.(*SessionTokenClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid session token")
	}
	return claims, nil
}
