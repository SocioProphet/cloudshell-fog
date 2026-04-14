package api

import (
	"testing"
	"time"

	"github.com/SocioProphet/cloudshell-fog/internal/placement"
)

func TestBuildCreateSessionResponseVNext(t *testing.T) {
	expiresAt := time.Date(2026, 4, 13, 4, 0, 0, 0, time.UTC)
	decision := placement.Decision{
		NodeID:  "fog-node-1",
		Region:  "us-east-1",
		Tier:    placement.TrustTierFog,
		Reasons: []string{"fog-preferred", "healthy", "capacity-ok"},
	}

	resp := BuildCreateSessionResponseVNext("sess-123", "wss://shell.example.com", "tok-abc", expiresAt, decision)

	if resp.SessionID != "sess-123" {
		t.Fatalf("expected session id, got %q", resp.SessionID)
	}
	if resp.Attach.WSURL != "wss://shell.example.com/v1/sessions/sess-123/pty" {
		t.Fatalf("unexpected ws url: %q", resp.Attach.WSURL)
	}
	if resp.Attach.Token != "tok-abc" {
		t.Fatalf("unexpected token: %q", resp.Attach.Token)
	}
	if resp.Placement.Region != "us-east-1" {
		t.Fatalf("unexpected region: %q", resp.Placement.Region)
	}
	if resp.Placement.NodeID != "fog-node-1" {
		t.Fatalf("unexpected node id: %q", resp.Placement.NodeID)
	}
	if resp.Placement.Tier != string(placement.TrustTierFog) {
		t.Fatalf("unexpected tier: %q", resp.Placement.Tier)
	}
	if len(resp.Placement.Reasons) != 3 {
		t.Fatalf("unexpected reasons: %#v", resp.Placement.Reasons)
	}
}
