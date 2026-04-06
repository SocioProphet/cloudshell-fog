// Package policy enforces per-subject quotas and profile admission.
package policy

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Profile describes the resource envelope for a session.
type Profile struct {
	Name          string   `yaml:"name"`
	CPU           string   `yaml:"cpu"`
	Memory        string   `yaml:"memory"`
	Storage       string   `yaml:"storage"`
	MaxTTLSeconds int      `yaml:"max_ttl_seconds"`
	MaxSessions   int      `yaml:"max_sessions"`
	AllowedGroups []string `yaml:"allowed_groups"`
}

// Config holds the set of named profiles loaded from YAML.
type Config struct {
	Profiles []Profile `yaml:"profiles"`
}

// Engine evaluates admission requests against loaded profiles.
type Engine struct {
	profiles map[string]Profile
}

// LoadConfig reads and parses a YAML policy file.
func LoadConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open policy config %q: %w", path, err)
	}
	defer f.Close()
	var cfg Config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("decode policy config: %w", err)
	}
	return &cfg, nil
}

// NewEngine builds an Engine from the parsed config.
func NewEngine(cfg *Config) *Engine {
	m := make(map[string]Profile, len(cfg.Profiles))
	for _, p := range cfg.Profiles {
		m[p.Name] = p
	}
	return &Engine{profiles: m}
}

// Sentinel errors returned by CheckAdmission.
var (
	ErrProfileNotFound  = errors.New("profile not found")
	ErrProfileForbidden = errors.New("profile not allowed for subject groups")
	ErrTTLExceeded      = errors.New("requested TTL exceeds profile maximum")
	ErrQuotaExceeded    = errors.New("session quota exceeded")
)

// GetProfile looks up a profile by name.
func (e *Engine) GetProfile(name string) (Profile, error) {
	p, ok := e.profiles[name]
	if !ok {
		return Profile{}, fmt.Errorf("%w: %s", ErrProfileNotFound, name)
	}
	return p, nil
}

// CheckAdmission verifies that subject (with groups) may open a session using
// profileName with the requested TTL, given activeSessions already open.
// Returns the resolved Profile on success, or a sentinel error on denial.
func (e *Engine) CheckAdmission(profileName string, groups []string, ttlSeconds int, activeSessions int) (Profile, error) {
	p, ok := e.profiles[profileName]
	if !ok {
		return Profile{}, fmt.Errorf("%w: %s", ErrProfileNotFound, profileName)
	}

	if len(p.AllowedGroups) > 0 {
		allowed := false
	outer:
		for _, ag := range p.AllowedGroups {
			for _, g := range groups {
				if ag == g {
					allowed = true
					break outer
				}
			}
		}
		if !allowed {
			return Profile{}, fmt.Errorf("%w: profile=%s", ErrProfileForbidden, profileName)
		}
	}

	if p.MaxTTLSeconds > 0 && ttlSeconds > p.MaxTTLSeconds {
		return Profile{}, fmt.Errorf("%w: requested=%d max=%d", ErrTTLExceeded, ttlSeconds, p.MaxTTLSeconds)
	}
	if p.MaxSessions > 0 && activeSessions >= p.MaxSessions {
		return Profile{}, fmt.Errorf("%w: active=%d max=%d", ErrQuotaExceeded, activeSessions, p.MaxSessions)
	}
	return p, nil
}
