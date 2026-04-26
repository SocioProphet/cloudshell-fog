// Package session manages cloud-shell session state.
package session

import (
	"context"
	"errors"
	"sync"
	"time"
)

// Status represents the lifecycle state of a session.
type Status string

const (
	StatusPending    Status = "pending"
	StatusRunning    Status = "running"
	StatusTerminated Status = "terminated"
	StatusExpired    Status = "expired"
)

// Session holds all metadata for a single shell session.
type Session struct {
	ID               string
	Subject          string
	Status           Status
	Profile          string
	Placement        string // legacy region-only placement field
	PlacementNodeID  string
	PlacementTier    string
	PlacementReasons []string
	RuntimeRef       string // "<namespace>/<pod>"
	ImageRef         string
	TTLSeconds       int
	CreatedAt        time.Time
	ExpiresAt        time.Time
	LastActiveAt     time.Time
}

// ErrNotFound is returned when a session cannot be found.
var ErrNotFound = errors.New("session not found")

// Store is the interface for session persistence.
type Store interface {
	Create(ctx context.Context, s *Session) error
	Get(ctx context.Context, id string) (*Session, error)
	Update(ctx context.Context, s *Session) error
	Delete(ctx context.Context, id string) error
	ListBySubject(ctx context.Context, subject string) ([]*Session, error)
}

// InMemoryStore is a thread-safe in-memory implementation of Store.
type InMemoryStore struct {
	mu       sync.RWMutex
	sessions map[string]*Session
}

// NewInMemoryStore returns an initialised InMemoryStore.
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{sessions: make(map[string]*Session)}
}

// Create stores a new session, overwriting any existing entry with the same ID.
func (s *InMemoryStore) Create(_ context.Context, sess *Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := *sess
	s.sessions[sess.ID] = &cp
	return nil
}

// Get returns a copy of the session with the given ID.
func (s *InMemoryStore) Get(_ context.Context, id string) (*Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sess, ok := s.sessions[id]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *sess
	return &cp, nil
}

// Update replaces the stored session; returns ErrNotFound if absent.
func (s *InMemoryStore) Update(_ context.Context, sess *Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.sessions[sess.ID]; !ok {
		return ErrNotFound
	}
	cp := *sess
	s.sessions[sess.ID] = &cp
	return nil
}

// Delete removes a session; a missing session is silently ignored.
func (s *InMemoryStore) Delete(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, id)
	return nil
}

// ListBySubject returns copies of all sessions owned by subject.
func (s *InMemoryStore) ListBySubject(_ context.Context, subject string) ([]*Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*Session
	for _, sess := range s.sessions {
		if sess.Subject == subject {
			cp := *sess
			result = append(result, &cp)
		}
	}
	return result, nil
}

// Sweeper periodically expires sessions whose TTL has elapsed.
type Sweeper struct {
	store    *InMemoryStore
	onExpire func(ctx context.Context, sess *Session)
	interval time.Duration
}

// NewSweeper returns a Sweeper that calls onExpire for each TTL-expired session.
func NewSweeper(store *InMemoryStore, onExpire func(context.Context, *Session), interval time.Duration) *Sweeper {
	return &Sweeper{store: store, onExpire: onExpire, interval: interval}
}

// Run starts the sweep loop; it exits when ctx is cancelled.
func (sw *Sweeper) Run(ctx context.Context) {
	ticker := time.NewTicker(sw.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			sw.sweep(ctx)
		}
	}
}

func (sw *Sweeper) sweep(ctx context.Context) {
	sw.store.mu.Lock()
	var expired []*Session
	now := time.Now()
	for id, sess := range sw.store.sessions {
		if (sess.Status == StatusRunning || sess.Status == StatusPending) && now.After(sess.ExpiresAt) {
			cp := *sess
			expired = append(expired, &cp)
			delete(sw.store.sessions, id)
		}
	}
	sw.store.mu.Unlock()
	for _, sess := range expired {
		sw.onExpire(ctx, sess)
	}
}
