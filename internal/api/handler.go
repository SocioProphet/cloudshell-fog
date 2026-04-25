// Package api implements the HTTP session management API (interfaces-v1 §1.2).
package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/SocioProphet/cloudshell-fog/internal/audit"
	"github.com/SocioProphet/cloudshell-fog/internal/auth"
	"github.com/SocioProphet/cloudshell-fog/internal/connector"
	"github.com/SocioProphet/cloudshell-fog/internal/placement"
	"github.com/SocioProphet/cloudshell-fog/internal/policy"
	"github.com/SocioProphet/cloudshell-fog/internal/session"
)

// Handler holds all dependencies for the session HTTP API.
type Handler struct {
	store      session.Store
	policy     *policy.Engine
	placement  *placement.Engine
	connector  connector.Connector
	minter     *auth.SessionTokenMinter
	emitter    audit.Emitter
	gatewayURL string // base URL used to construct ws_url (e.g. "wss://shell.example.com")
}

// NewHandler constructs an API Handler.
func NewHandler(
	store session.Store,
	pol *policy.Engine,
	place *placement.Engine,
	conn connector.Connector,
	minter *auth.SessionTokenMinter,
	emitter audit.Emitter,
	gatewayURL string,
) *Handler {
	return &Handler{
		store:      store,
		policy:     pol,
		placement:  place,
		connector:  conn,
		minter:     minter,
		emitter:    emitter,
		gatewayURL: gatewayURL,
	}
}

// createSessionRequest is the body for POST /v1/sessions.
type createSessionRequest struct {
	Profile       string `json:"profile"`
	TTLSeconds    int    `json:"ttl_seconds"`
	PlacementHint string `json:"placement_hint,omitempty"`
	ImageRef      string `json:"image_ref,omitempty"`
}

// CreateSession handles POST /v1/sessions.
func (h *Handler) CreateSession(w http.ResponseWriter, r *http.Request) {
	subject := auth.SubjectFromContext(r.Context())
	groups := auth.GroupsFromContext(r.Context())

	var req createSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	if req.Profile == "" {
		writeError(w, http.StatusBadRequest, "profile is required")
		return
	}
	if req.TTLSeconds <= 0 {
		req.TTLSeconds = 3600
	}

	// Count active sessions for quota check.
	existing, err := h.store.ListBySubject(r.Context(), subject)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list sessions: "+err.Error())
		return
	}
	active := 0
	for _, s := range existing {
		if s.Status == session.StatusRunning || s.Status == session.StatusPending {
			active++
		}
	}

	prof, err := h.policy.CheckAdmission(req.Profile, groups, req.TTLSeconds, active)
	if err != nil {
		h.emitter.Emit(r.Context(), audit.Event{
			TS:      time.Now(),
			Subject: subject,
			Type:    audit.EventPolicyDenied,
			Details: map[string]any{"rule": err.Error()},
		})
		writeError(w, http.StatusForbidden, "policy denied: "+err.Error())
		return
	}

	decision, err := h.placement.Decide(r.Context(), req.PlacementHint)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "placement failed: "+err.Error())
		return
	}

	if req.ImageRef == "" {
		req.ImageRef = "ghcr.io/socioprophet/cloudshell-runtime:latest"
	}

	sessionID := uuid.NewString()
	now := time.Now()
	expiresAt := now.Add(time.Duration(req.TTLSeconds) * time.Second)

	sess := &session.Session{
		ID:           sessionID,
		Subject:      subject,
		Status:       session.StatusPending,
		Profile:      req.Profile,
		Placement:    decision.Region,
		ImageRef:     req.ImageRef,
		TTLSeconds:   req.TTLSeconds,
		CreatedAt:    now,
		ExpiresAt:    expiresAt,
		LastActiveAt: now,
	}
	if err := h.store.Create(r.Context(), sess); err != nil {
		writeError(w, http.StatusInternalServerError, "create session record: "+err.Error())
		return
	}

	connProfile := connector.Profile{
		CPU:     prof.CPU,
		Memory:  prof.Memory,
		Storage: prof.Storage,
	}
	runtimeRef, err := h.connector.Allocate(r.Context(), sessionID, connProfile, decision.NodeID, req.ImageRef)
	if err != nil {
		sess.Status = session.StatusTerminated
		_ = h.store.Update(r.Context(), sess)
		writeError(w, http.StatusInternalServerError, "allocate runtime: "+err.Error())
		return
	}

	sess.Status = session.StatusRunning
	sess.RuntimeRef = runtimeRef.Namespace + "/" + runtimeRef.PodName
	if err := h.store.Update(r.Context(), sess); err != nil {
		writeError(w, http.StatusInternalServerError, "update session record: "+err.Error())
		return
	}

	// Emit all required audit events.
	h.emitter.Emit(r.Context(), audit.Event{
		TS: now, SessionID: sessionID, Subject: subject,
		Type: audit.EventSessionCreated, Placement: decision.Region,
	})
	h.emitter.Emit(r.Context(), audit.Event{
		TS: now, SessionID: sessionID, Subject: subject,
		Type: audit.EventPlacementDecided, Placement: decision.Region,
		Details: map[string]any{
			"node_id": decision.NodeID,
			"tier":    string(decision.Tier),
			"reasons": decision.Reasons,
		},
	})
	h.emitter.Emit(r.Context(), audit.Event{
		TS: now, SessionID: sessionID, Subject: subject,
		Type: audit.EventRuntimeAllocated, Placement: decision.Region,
		Details: map[string]any{"image_ref": req.ImageRef},
	})

	token, tokenExpiresAt, err := h.minter.Mint(sessionID, subject)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "mint session token: "+err.Error())
		return
	}

	resp := BuildCreateSessionResponseVNext(sessionID, h.gatewayURL, token, tokenExpiresAt, decision)
	writeJSON(w, http.StatusCreated, resp)
}

// GetSession handles GET /v1/sessions/{id}.
func (h *Handler) GetSession(w http.ResponseWriter, r *http.Request) {
	subject := auth.SubjectFromContext(r.Context())
	id := mux.Vars(r)["id"]

	sess, err := h.store.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, session.ErrNotFound) {
			writeError(w, http.StatusNotFound, "session not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if sess.Subject != subject {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status":     string(sess.Status),
		"placement":  sess.Placement,
		"created_at": sess.CreatedAt.UTC().Format(time.RFC3339),
		"expires_at": sess.ExpiresAt.UTC().Format(time.RFC3339),
		"image_ref":  sess.ImageRef,
	})
}

// DeleteSession handles DELETE /v1/sessions/{id}.
func (h *Handler) DeleteSession(w http.ResponseWriter, r *http.Request) {
	subject := auth.SubjectFromContext(r.Context())
	id := mux.Vars(r)["id"]

	sess, err := h.store.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, session.ErrNotFound) {
			writeError(w, http.StatusNotFound, "session not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if sess.Subject != subject {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	ref := connector.RuntimeRef{
		ID:        sess.ID,
		Namespace: "cloudshell-" + sess.ID,
		PodName:   "shell",
		NodeID:    sess.Placement,
	}
	if err := h.connector.Terminate(r.Context(), ref); err != nil {
		writeError(w, http.StatusInternalServerError, "terminate runtime: "+err.Error())
		return
	}

	sess.Status = session.StatusTerminated
	_ = h.store.Update(r.Context(), sess)

	h.emitter.Emit(r.Context(), audit.Event{
		TS:        time.Now(),
		SessionID: id,
		Subject:   subject,
		Type:      audit.EventSessionTerminated,
		Placement: sess.Placement,
	})
	writeJSON(w, http.StatusOK, map[string]any{"terminated": true})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]any{"error": msg})
}
