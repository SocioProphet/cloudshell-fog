// Package placement implements the fog placement engine described in interfaces-v1 §3.
package placement

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

// TrustTier distinguishes attested fog nodes from cloud fallback nodes.
type TrustTier string

const (
	TrustTierFog   TrustTier = "fog"
	TrustTierCloud TrustTier = "cloud"
)

// NodeHealth represents the operational state of a placement node.
type NodeHealth string

const (
	NodeHealthHealthy   NodeHealth = "healthy"
	NodeHealthDegraded  NodeHealth = "degraded"
	NodeHealthUnhealthy NodeHealth = "unhealthy"
)

// Node represents a runtime placement candidate (fog node or cloud region).
type Node struct {
	ID           string
	Region       string
	Tier         TrustTier
	Health       NodeHealth
	CapacityFree float64 // fraction 0..1
	LatencyMS    int     // estimated one-way latency to user region in ms
	UpdatedAt    time.Time
}

// Decision is the output of the placement engine.
type Decision struct {
	NodeID  string
	Region  string
	Tier    TrustTier
	Reasons []string
}

// Registry holds the live set of known nodes.
type Registry struct {
	mu    sync.RWMutex
	nodes map[string]*Node
}

// NewRegistry returns an empty Registry.
func NewRegistry() *Registry {
	return &Registry{nodes: make(map[string]*Node)}
}

// Upsert adds or replaces a node record.
func (r *Registry) Upsert(n Node) {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := n
	r.nodes[n.ID] = &cp
}

// List returns a snapshot of all registered nodes.
func (r *Registry) List() []Node {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]Node, 0, len(r.nodes))
	for _, n := range r.nodes {
		result = append(result, *n)
	}
	return result
}

// Engine selects a placement node for a new session.
type Engine struct {
	registry      *Registry
	cloudFallback string // region name of last-resort cloud fallback
}

// NewEngine builds an Engine from a Registry and a mandatory cloud-fallback region name.
func NewEngine(registry *Registry, cloudFallback string) *Engine {
	return &Engine{registry: registry, cloudFallback: cloudFallback}
}

// Decide selects the best available node given an optional placement hint (region name).
//
// Selection order (per interfaces-v1 §3):
//  1. Healthy fog nodes in the hinted region, sorted by latency then free capacity.
//  2. Any other healthy fog node, sorted the same way.
//  3. Healthy cloud nodes, sorted by latency.
//  4. Configured cloud-fallback region when no registered node qualifies.
func (e *Engine) Decide(_ context.Context, hint string) (Decision, error) {
	nodes := e.registry.List()

	var fogNodes, cloudNodes []Node
	for _, n := range nodes {
		if n.Health != NodeHealthHealthy || n.CapacityFree <= 0.1 {
			continue
		}
		if n.Tier == TrustTierFog {
			fogNodes = append(fogNodes, n)
		} else {
			cloudNodes = append(cloudNodes, n)
		}
	}

	// Narrow fog candidates to hinted region when a hint is provided.
	if hint != "" {
		var hinted []Node
		for _, n := range fogNodes {
			if n.Region == hint {
				hinted = append(hinted, n)
			}
		}
		if len(hinted) > 0 {
			fogNodes = hinted
		}
	}

	byLatencyCapacity := func(ns []Node) {
		sort.Slice(ns, func(i, j int) bool {
			if ns[i].LatencyMS != ns[j].LatencyMS {
				return ns[i].LatencyMS < ns[j].LatencyMS
			}
			return ns[i].CapacityFree > ns[j].CapacityFree
		})
	}

	byLatencyCapacity(fogNodes)
	if len(fogNodes) > 0 {
		n := fogNodes[0]
		return Decision{
			NodeID:  n.ID,
			Region:  n.Region,
			Tier:    TrustTierFog,
			Reasons: []string{"fog-preferred", "healthy", "capacity-ok"},
		}, nil
	}

	byLatencyCapacity(cloudNodes)
	if len(cloudNodes) > 0 {
		n := cloudNodes[0]
		return Decision{
			NodeID:  n.ID,
			Region:  n.Region,
			Tier:    TrustTierCloud,
			Reasons: []string{"fog-unavailable", "cloud-fallback"},
		}, nil
	}

	if e.cloudFallback != "" {
		return Decision{
			NodeID:  "cloud-default",
			Region:  e.cloudFallback,
			Tier:    TrustTierCloud,
			Reasons: []string{"no-nodes-registered", "cloud-fallback-default"},
		}, nil
	}

	return Decision{}, fmt.Errorf("no placement available: no healthy nodes and no cloud fallback configured")
}
