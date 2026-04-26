package api

import (
	"fmt"
	"time"

	"github.com/SocioProphet/cloudshell-fog/internal/placement"
	"github.com/SocioProphet/cloudshell-fog/internal/session"
)

// PlacementResponse is a richer placement payload than the current region-only
// response. It preserves the implementation's placement decision metadata in a
// stable JSON shape for future API alignment work.
type PlacementResponse struct {
	Region  string   `json:"region"`
	NodeID  string   `json:"node_id"`
	Tier    string   `json:"tier"`
	Reasons []string `json:"reasons,omitempty"`
}

// AttachResponse describes the PTY attach details returned after session creation.
type AttachResponse struct {
	WSURL     string `json:"ws_url"`
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
}

// CreateSessionResponseVNext is a richer response envelope for session creation.
type CreateSessionResponseVNext struct {
	SessionID string            `json:"session_id"`
	Attach    AttachResponse    `json:"attach"`
	Placement PlacementResponse `json:"placement"`
}

// SessionStatusResponseVNext is a richer response envelope for session status.
type SessionStatusResponseVNext struct {
	SessionID string            `json:"session_id"`
	Status    string            `json:"status"`
	Placement PlacementResponse `json:"placement"`
	CreatedAt string            `json:"created_at"`
	ExpiresAt string            `json:"expires_at"`
	ImageRef  string            `json:"image_ref"`
}

// BuildCreateSessionResponseVNext converts the current placement engine decision
// plus attach material into a richer response payload.
func BuildCreateSessionResponseVNext(sessionID, gatewayURL, token string, tokenExpiresAt time.Time, decision placement.Decision) CreateSessionResponseVNext {
	return CreateSessionResponseVNext{
		SessionID: sessionID,
		Attach: AttachResponse{
			WSURL:     fmt.Sprintf("%s/v1/sessions/%s/pty", gatewayURL, sessionID),
			Token:     token,
			ExpiresAt: tokenExpiresAt.UTC().Format(time.RFC3339),
		},
		Placement: PlacementResponse{
			Region:  decision.Region,
			NodeID:  decision.NodeID,
			Tier:    string(decision.Tier),
			Reasons: append([]string(nil), decision.Reasons...),
		},
	}
}

// BuildSessionStatusResponseVNext converts stored session placement metadata into
// a richer response payload for GET /v1/sessions/{id}.
func BuildSessionStatusResponseVNext(sess *session.Session) SessionStatusResponseVNext {
	return SessionStatusResponseVNext{
		SessionID: sess.ID,
		Status:    string(sess.Status),
		Placement: PlacementResponse{
			Region:  sess.Placement,
			NodeID:  sess.PlacementNodeID,
			Tier:    sess.PlacementTier,
			Reasons: append([]string(nil), sess.PlacementReasons...),
		},
		CreatedAt: sess.CreatedAt.UTC().Format(time.RFC3339),
		ExpiresAt: sess.ExpiresAt.UTC().Format(time.RFC3339),
		ImageRef:  sess.ImageRef,
	}
}
