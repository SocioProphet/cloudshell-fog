// Package connector defines the runtime connector interface and implementations.
package connector

import (
	"context"
	"io"
)

// Profile describes the resource limits to apply to a new runtime.
type Profile struct {
	CPU     string // e.g. "500m"
	Memory  string // e.g. "512Mi"
	Storage string // e.g. "1Gi"
}

// RuntimeRef is an opaque reference to an allocated runtime.
type RuntimeRef struct {
	ID        string // session ID
	Namespace string // k8s namespace (or stub equivalent)
	PodName   string // k8s pod name
	NodeID    string // placement node ID
}

// PTYStreams holds the I/O channels for an attached PTY session.
type PTYStreams struct {
	Stdin  io.WriteCloser
	Stdout io.ReadCloser
	Resize func(cols, rows uint16) error
}

// Connector is the internal runtime connector contract described in interfaces-v1 §1.4.
type Connector interface {
	// Allocate provisions a new runtime for sessionID with the given constraints.
	Allocate(ctx context.Context, sessionID string, profile Profile, placement string, imageRef string) (RuntimeRef, error)
	// AttachPTY opens interactive PTY streams to an allocated runtime.
	AttachPTY(ctx context.Context, ref RuntimeRef) (*PTYStreams, error)
	// EnforceLimits verifies the runtime is within its resource envelope.
	EnforceLimits(ctx context.Context, ref RuntimeRef) error
	// Terminate deallocates the runtime and all associated resources.
	Terminate(ctx context.Context, ref RuntimeRef) error
}
