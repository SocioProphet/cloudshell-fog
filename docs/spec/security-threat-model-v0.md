# Security Threat Model v0 — cloudshell-fog

## 0. Purpose

Describe the minimum adversary model, trust assumptions, and control surfaces for cloudshell-fog.

## 1. Assets

Primary assets:
- subject identity and group claims
- session metadata and ownership
- runtime execution boundary
- session namespace and pod resources
- image provenance / signature / digest trust
- audit events
- deployment credentials and signing trust material

## 2. Adversaries

### A. External attacker with network access
Attempts to:
- abuse unauthenticated endpoints
- replay tokens
- exploit WebSocket / PTY path

### B. Authenticated but malicious tenant
Attempts to:
- access another subject's session
- escape runtime boundaries
- abuse namespace / cluster privileges

### C. Compromised build or image source
Attempts to:
- inject malicious runtime image
- exploit mutable tags / unsigned artifacts

### D. Compromised fog node or degraded environment
Attempts to:
- impersonate healthy runtime target
- exfiltrate data or subvert runtime isolation

## 3. Trust assumptions

Current baseline assumes:
- OIDC issuer is trustworthy
- policy engine configuration is correctly maintained
- Kubernetes control plane enforces RBAC and admission policy
- audit logs are emitted truthfully by the gateway process

These assumptions should be weakened over time with stronger attestation and external log shipping.

## 4. Primary controls

### Identity and token controls
- OIDC bearer token required for control API
- session-scoped short-lived PTY token for attach path
- attach token subject/session binding enforced in PTY handler

### Runtime controls
- isolated per-session namespace model (target posture)
- non-root, no privilege escalation, dropped capabilities, RuntimeDefault seccomp
- default-deny network posture for managed session pods

### Supply-chain controls
- digest-only production posture
- signed-image verification
- provenance / Chains posture
- SBOM generation and retention

### Observability and audit
- session created / attached / terminated
- placement decided
- runtime allocated
- policy denied
- PTY resize

## 5. Threats and mitigations

### T1. PTY token replay
Mitigations:
- short TTL
- audience restriction
- session binding
- subject binding check at attach

### T2. Cross-session attach
Mitigations:
- session ownership check in GET/DELETE and PTY attach
- namespace-per-session posture
- connector runtime ref scoped to session ID

### T3. Mutable-tag drift
Mitigations:
- require-image-digest policy
- production overlays / pinned references

### T4. Unsigned or untrusted image execution
Mitigations:
- signed-image verification policy
- keyless trust model or explicit public key model

### T5. Excessive runtime privilege
Mitigations:
- Kyverno runtime baseline
- hardened pod spec in connector implementation
- least privilege service account and namespace controls

### T6. Cluster credential ambiguity
Mitigations:
- explicit connector mode selection
- explicit kubeconfig vs in-cluster credential model
- fail-fast startup on missing k8s config

## 6. Known current weaknesses

- gateway connector selection is only partially wired in reviewed mainline
- signed-image policy still contains placeholder trust material in baseline example
- deployment examples still include mutable tags in some places
- per-session network policy application mechanism was previously under-specified

## 7. Priority security backlog

1. finish connector selection wiring
2. complete signing trust model
3. eliminate misleading mutable-tag production examples
4. make per-session namespace bootstrap real
5. strengthen state-machine semantics for degraded/disconnected flows
