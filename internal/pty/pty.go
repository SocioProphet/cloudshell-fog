// Package pty implements the WebSocket ↔ PTY bridge described in interfaces-v1 §1.3.
package pty

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"github.com/SocioProphet/cloudshell-fog/internal/audit"
	"github.com/SocioProphet/cloudshell-fog/internal/auth"
	"github.com/SocioProphet/cloudshell-fog/internal/connector"
	"github.com/SocioProphet/cloudshell-fog/internal/session"
)

var upgrader = websocket.Upgrader{
	HandshakeTimeout: 10 * time.Second,
	// CheckOrigin is permissive here; production deployments should restrict to known origins.
	CheckOrigin: func(_ *http.Request) bool { return true },
}

// msgType is the discriminator field in WebSocket message frames.
type msgType string

const (
	msgResize msgType = "resize"
	msgStdin  msgType = "stdin"
	msgStdout msgType = "stdout"
	msgExit   msgType = "exit"
)

// frame is the JSON message schema defined in interfaces-v1 §1.3.
type frame struct {
	Type    msgType `json:"type"`
	Cols    uint16  `json:"cols,omitempty"`
	Rows    uint16  `json:"rows,omitempty"`
	DataB64 string  `json:"data_b64,omitempty"`
	Code    int     `json:"code,omitempty"`
}

// Handler handles WebSocket PTY attach requests at /v1/sessions/{id}/pty.
type Handler struct {
	minter    *auth.SessionTokenMinter
	store     session.Store
	connector connector.Connector
	emitter   audit.Emitter
}

// NewHandler creates a Handler with the given dependencies.
func NewHandler(
	minter *auth.SessionTokenMinter,
	store session.Store,
	conn connector.Connector,
	emitter audit.Emitter,
) *Handler {
	return &Handler{
		minter:    minter,
		store:     store,
		connector: conn,
		emitter:   emitter,
	}
}

// ServeHTTP upgrades to WebSocket and bridges PTY I/O for the session.
// Auth: validates the session token from the ?token= query parameter.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["id"]
	rawToken := r.URL.Query().Get("token")

	// Validate the short-lived session token (not the OIDC token).
	claims, err := h.minter.Validate(rawToken)
	if err != nil {
		http.Error(w, "unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}
	if claims.SessionID != sessionID {
		http.Error(w, "token session mismatch", http.StatusForbidden)
		return
	}

	sess, err := h.store.Get(r.Context(), sessionID)
	if err != nil {
		http.Error(w, "session not found", http.StatusNotFound)
		return
	}
	if sess.Subject != claims.Subject {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if sess.Status != session.StatusRunning {
		http.Error(w, fmt.Sprintf("session not running (status=%s)", sess.Status), http.StatusConflict)
		return
	}

	ref := connector.RuntimeRef{
		ID:        sess.ID,
		Namespace: "cloudshell-" + sess.ID,
		PodName:   "shell",
		NodeID:    sess.Placement,
	}

	streams, err := h.connector.AttachPTY(r.Context(), ref)
	if err != nil {
		http.Error(w, "attach failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// Upgrade writes its own error response on failure.
		streams.Stdin.Close()
		return
	}
	defer ws.Close()
	defer streams.Stdin.Close()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Close the WebSocket when context is cancelled (e.g. PTY exit).
	go func() {
		<-ctx.Done()
		ws.Close()
	}()

	h.emitter.Emit(ctx, audit.Event{
		TS:        time.Now(),
		SessionID: sessionID,
		Subject:   claims.Subject,
		Type:      audit.EventSessionAttached,
		Placement: sess.Placement,
	})

	// Goroutine: PTY stdout → WebSocket frames.
	go func() {
		defer cancel()
		buf := make([]byte, 4096)
		for {
			n, err := streams.Stdout.Read(buf)
			if n > 0 {
				f := frame{
					Type:    msgStdout,
					DataB64: base64.StdEncoding.EncodeToString(buf[:n]),
				}
				if werr := ws.WriteJSON(f); werr != nil {
					return
				}
			}
			if err != nil {
				exitCode := 0
				if err != io.EOF {
					exitCode = 1
				}
				_ = ws.WriteJSON(frame{Type: msgExit, Code: exitCode})
				return
			}
		}
	}()

	// Main loop: WebSocket frames → PTY stdin / resize.
	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			return
		}
		var f frame
		if err := json.Unmarshal(msg, &f); err != nil {
			continue
		}
		switch f.Type {
		case msgStdin:
			data, err := base64.StdEncoding.DecodeString(f.DataB64)
			if err != nil {
				continue
			}
			if _, err := streams.Stdin.Write(data); err != nil {
				return
			}
		case msgResize:
			_ = streams.Resize(f.Cols, f.Rows)
			h.emitter.Emit(ctx, audit.Event{
				TS:        time.Now(),
				SessionID: sessionID,
				Subject:   claims.Subject,
				Type:      audit.EventPTYResize,
				Placement: sess.Placement,
				Details:   map[string]any{"cols": f.Cols, "rows": f.Rows},
			})
		}
	}
}
