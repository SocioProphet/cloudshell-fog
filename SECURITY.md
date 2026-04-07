# Security Policy

## Supported Versions

| Version | Supported |
|---------|-----------|
| `main`  | ✅ Yes     |

We currently support only the `main` branch. Once tagged releases are published, this table will be updated.

## Reporting a Vulnerability

**Please do not open a public GitHub issue for security vulnerabilities.**

To report a security issue, email the maintainers at the address listed in the repository's GitHub profile, or use [GitHub's private vulnerability reporting](https://docs.github.com/en/code-security/security-advisories/guidance-on-reporting-and-writing/privately-reporting-a-security-vulnerability) feature:

1. Navigate to the **Security** tab of this repository.
2. Click **Report a vulnerability**.
3. Fill in the details and submit.

We will acknowledge receipt within **3 business days** and aim to provide a fix or mitigation within **14 days** for critical issues.

## Security Design Principles

cloudshell-fog is designed with the following security invariants (see [`docs/spec/interfaces-v1.md`](docs/spec/interfaces-v1.md) §2):

- **No long-lived secrets in the browser or gateway.** Session tokens are short-lived JWTs bound to a single session.
- **OIDC-only authentication.** The gateway validates access tokens from a configured OIDC provider; it does not issue its own identity tokens.
- **Pinned container digests in production.** Mutable image tags are rejected by the admission policy.
- **Supply-chain integrity.** Production builds require a cosign signature, a provenance attestation (Tekton Chains / in-toto), and an SPDX/CycloneDX SBOM.
- **Network isolation.** Each session pod runs in its own namespace with default-deny NetworkPolicies; only DNS and HTTPS egress are permitted.
- **Least-privilege RBAC.** The gateway service account has the minimum permissions required to manage session namespaces and pods.

## Scope

Reports are welcome for any component in this repository, including:

- The Go gateway (`cmd/gateway`, `internal/`)
- The TypeScript browser UI (`web/`)
- Kubernetes / Argo CD / Tekton deployment manifests (`deploy/`)
- The policy configuration (`config/`)

Out-of-scope: third-party dependencies (please report those upstream), and issues in a fork or derived project not maintained here.
