package connector

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
)

// StubConnector is a minimal in-process connector for dev/test without a real k8s cluster.
// It spawns an echo loop that mimics a simple shell prompt.
type StubConnector struct {
	mu       sync.Mutex
	sessions map[string]*stubSession
}

type stubSession struct {
	stdinW  *io.PipeWriter
	stdoutW *io.PipeWriter
	stdoutR *io.PipeReader
}

// NewStubConnector returns an initialised StubConnector.
func NewStubConnector() *StubConnector {
	return &StubConnector{sessions: make(map[string]*stubSession)}
}

// Allocate creates a stub echo session and starts the echo goroutine.
func (s *StubConnector) Allocate(_ context.Context, sessionID string, profile Profile, placement string, imageRef string) (RuntimeRef, error) {
	stdinR, stdinW := io.Pipe()
	stdoutR, stdoutW := io.Pipe()

	sess := &stubSession{
		stdinW:  stdinW,
		stdoutW: stdoutW,
		stdoutR: stdoutR,
	}

	s.mu.Lock()
	s.sessions[sessionID] = sess
	s.mu.Unlock()

	// Echo goroutine: write a banner then echo stdin lines.
	go func() {
		banner := fmt.Sprintf("cloudshell stub (image=%s placement=%s)\r\n$ ", imageRef, placement)
		_, _ = stdoutW.Write([]byte(banner))
		buf := make([]byte, 4096)
		for {
			n, err := stdinR.Read(buf)
			if err != nil {
				break
			}
			chunk := string(buf[:n])
			_, _ = stdoutW.Write([]byte(chunk))
			if strings.ContainsAny(chunk, "\r\n") {
				_, _ = stdoutW.Write([]byte("$ "))
			}
		}
		stdoutW.Close()
	}()

	return RuntimeRef{
		ID:        sessionID,
		Namespace: "stub-" + sessionID,
		PodName:   "shell",
		NodeID:    placement,
	}, nil
}

// AttachPTY returns the pipe streams for the stub echo session.
func (s *StubConnector) AttachPTY(_ context.Context, ref RuntimeRef) (*PTYStreams, error) {
	s.mu.Lock()
	sess, ok := s.sessions[ref.ID]
	s.mu.Unlock()
	if !ok {
		return nil, fmt.Errorf("stub: session not found: %s", ref.ID)
	}
	return &PTYStreams{
		Stdin:  sess.stdinW,
		Stdout: sess.stdoutR,
		Resize: func(_, _ uint16) error { return nil },
	}, nil
}

// EnforceLimits is a no-op for the stub connector.
func (s *StubConnector) EnforceLimits(_ context.Context, _ RuntimeRef) error { return nil }

// Terminate closes the stub session's pipes.
func (s *StubConnector) Terminate(_ context.Context, ref RuntimeRef) error {
	s.mu.Lock()
	sess, ok := s.sessions[ref.ID]
	if ok {
		delete(s.sessions, ref.ID)
	}
	s.mu.Unlock()
	if ok {
		sess.stdinW.Close()
		sess.stdoutW.Close()
	}
	return nil
}
