// Package audit provides structured audit event emission.
package audit

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"
)

// EventType identifies the kind of audit event.
type EventType string

const (
	EventSessionCreated    EventType = "session.created"
	EventSessionAttached   EventType = "session.attached"
	EventSessionTerminated EventType = "session.terminated"
	EventPlacementDecided  EventType = "placement.decided"
	EventRuntimeAllocated  EventType = "runtime.allocated"
	EventPolicyDenied      EventType = "policy.denied"
	EventPTYResize         EventType = "pty.resize"
)

// Event is the canonical audit log record (interfaces-v0 §5 / interfaces-v1 §1.5).
type Event struct {
	TS        time.Time      `json:"ts"`
	SessionID string         `json:"session_id,omitempty"`
	Subject   string         `json:"subject"`
	Type      EventType      `json:"type"`
	Placement string         `json:"placement,omitempty"`
	Details   map[string]any `json:"details,omitempty"`
}

// Emitter is the interface for emitting audit events.
type Emitter interface {
	Emit(ctx context.Context, ev Event)
}

// LogEmitter emits audit events as structured JSON log lines via slog.
type LogEmitter struct {
	logger *slog.Logger
}

// NewLogEmitter returns a LogEmitter backed by the provided slog.Logger.
func NewLogEmitter(logger *slog.Logger) *LogEmitter {
	return &LogEmitter{logger: logger}
}

// Emit serialises the event to JSON and writes it as an info log record.
func (e *LogEmitter) Emit(ctx context.Context, ev Event) {
	b, _ := json.Marshal(ev)
	e.logger.InfoContext(ctx, "audit", "event", string(b))
}
