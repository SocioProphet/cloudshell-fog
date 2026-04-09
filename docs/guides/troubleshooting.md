# Troubleshooting

This guide covers common issues with cloudshell-fog and how to diagnose them.

## Quick diagnostics checklist

Before diving into specific issues, run through these checks:

```bash
# 1. Is the gateway running?
kubectl -n cloudshell-system get pods

# 2. Is it healthy?
curl -s https://shell.example.com/healthz

# 3. What do the recent logs say?
kubectl -n cloudshell-system logs deployment/cloudshell-gateway --tail=50

# 4. Are there any Kubernetes events?
kubectl -n cloudshell-system get events --sort-by=.lastTimestamp
```

---

## Authentication issues

### `401 Unauthorized` on all requests

**Symptoms:** Every request to `/v1/sessions` returns `401`.

**Causes and fixes:**

1. **OIDC is misconfigured** — Check that `OIDC_ISSUER_URL` and `OIDC_CLIENT_ID` are set and correct.

   ```bash
   kubectl -n cloudshell-system exec deploy/cloudshell-gateway -- env | grep OIDC
   ```

2. **Token `iss` does not match `OIDC_ISSUER_URL`** — Decode your token (e.g. at [jwt.io](https://jwt.io)) and compare the `iss` claim with the configured issuer URL. They must match exactly (including trailing slashes).

3. **Token is expired** — Access tokens are short-lived. Obtain a fresh token and retry.

4. **JWKS endpoint is unreachable from the pod** — Test connectivity from inside the cluster:

   ```bash
   kubectl -n cloudshell-system exec deploy/cloudshell-gateway -- \
     wget -qO- https://accounts.example.com/.well-known/openid-configuration
   ```

   If this fails, check NetworkPolicies and DNS resolution in the cluster.

5. **Using dev auth shim accidentally in production** — If neither `OIDC_ISSUER_URL` nor `OIDC_CLIENT_ID` is set, the gateway uses the dev shim. The dev shim accepts any token in local dev mode only; it does not cause `401` errors — but its presence means OIDC is not active. Check the startup logs:

   ```bash
   kubectl -n cloudshell-system logs deploy/cloudshell-gateway | grep "auth"
   ```

---

## Session creation fails

### `403 Forbidden` — policy denied

**Symptoms:** `POST /v1/sessions` returns `403` with a message like `policy denied: group not allowed` or `quota exceeded`.

**Causes and fixes:**

1. **Group not allowed for the requested profile** — The `allowed_groups` field in `config/policy.yaml` restricts access by OIDC group. Verify:
   - The requested profile name is correct.
   - The user's `groups` claim in the token includes a group listed in `allowed_groups`.
   - The group name matches exactly (case-sensitive).

2. **Per-user session quota exceeded** — The profile's `max_sessions` limit is reached. Terminate idle sessions:

   ```bash
   curl -X DELETE https://shell.example.com/v1/sessions/<session_id> \
     -H "Authorization: Bearer $TOKEN"
   ```

3. **TTL exceeds `max_ttl_seconds`** — The requested `ttl_seconds` is higher than the profile allows. Use a smaller value or increase the policy limit.

### `500 Internal Server Error` — connector failure

**Symptoms:** Session creation returns `500` and logs show a connector error.

**Causes and fixes:**

1. **`USE_K8S` not set** — The gateway is running with the stub connector. Set `USE_K8S=1` in the deployment.

2. **RBAC permissions** — The gateway service account lacks permission to create pods/namespaces. Check:

   ```bash
   kubectl -n cloudshell-system auth can-i create pods --as=system:serviceaccount:cloudshell-system:cloudshell-gateway
   ```

   If this returns `no`, re-apply the RBAC manifests:

   ```bash
   kubectl apply -f deploy/k8s/rbac.yaml
   ```

3. **Pod image pull failure** — The session pod image cannot be pulled. Check pod events in the session namespace:

   ```bash
   kubectl get pods --all-namespaces | grep cloudshell-session
   kubectl describe pod <session-pod> -n <session-namespace>
   ```

---

## PTY / WebSocket issues

### Terminal connects but no output appears

**Symptoms:** The xterm.js UI opens but the shell prompt never appears.

**Causes and fixes:**

1. **Runtime pod is not ready** — The session pod may still be initialising. Check its status:

   ```bash
   kubectl get pods --all-namespaces | grep <session_id>
   ```

2. **Short-lived token expired** — The PTY token is valid for 15 minutes. If there was a long delay between session creation and PTY attach, obtain a new session.

3. **WebSocket not reaching the gateway** — Ensure the ingress supports WebSocket and has long-enough proxy timeouts:

   ```yaml
   nginx.ingress.kubernetes.io/proxy-read-timeout: "3600"
   nginx.ingress.kubernetes.io/proxy-send-timeout: "3600"
   ```

### `403` when connecting to PTY WebSocket

**Symptoms:** The WebSocket upgrade returns `403`.

**Cause:** The `token` query parameter is invalid, expired, or belongs to a different session.

**Fix:** Obtain a fresh token by creating a new session (`POST /v1/sessions`) and use the `attach.token` from the response immediately.

### Terminal disconnects immediately

**Symptoms:** The terminal connects briefly then closes with exit code 0 or 1.

**Causes and fixes:**

1. **Stub connector (dev mode)** — The stub connector opens `/bin/sh` and terminates when the shell exits. Type a command to keep the shell alive.

2. **Session TTL expired** — The session's TTL has passed. Create a new session with a longer `ttl_seconds`.

3. **Runtime pod OOMKilled** — The session pod ran out of memory. Check pod status:

   ```bash
   kubectl describe pod <session-pod> -n <session-namespace>
   ```

   Increase the profile memory limit in `config/policy.yaml`.

---

## Placement issues

### All sessions land in the cloud fallback region

**Symptoms:** Every session's `placement` is the `CLOUD_FALLBACK_REGION` value.

**Cause:** No fog nodes are registered, or all registered fog nodes are unhealthy.

**Fix:** The fog node registry (`internal/placement`) is populated at runtime. Nodes must call the internal registration endpoint to register themselves. If you have no fog nodes, this behaviour is expected — the cloud fallback is working correctly.

---

## Kubernetes deployment issues

### Pods stuck in `Pending`

```bash
kubectl describe pod <pod-name> -n cloudshell-system
```

Common reasons: resource constraints (no nodes with sufficient CPU/memory), missing image pull secret, or PVC binding failure.

### `CrashLoopBackOff`

```bash
kubectl logs <pod-name> -n cloudshell-system --previous
```

Look for startup errors such as:
- `failed to read policy config` — `config/policy.yaml` is missing or malformed.
- `OIDC issuer unreachable` — the OIDC provider URL is not reachable from the pod.
- `invalid session token signing key` — `SESSION_TOKEN_SIGNING_KEY` is too short (must be 32+ bytes).

---

## Collecting debug information

When opening a bug report, include the following:

```bash
# Gateway version (git SHA)
kubectl -n cloudshell-system get deploy cloudshell-gateway \
  -o jsonpath='{.spec.template.spec.containers[0].image}'

# Recent logs
kubectl -n cloudshell-system logs deploy/cloudshell-gateway --tail=100

# Pod events
kubectl -n cloudshell-system get events --sort-by=.lastTimestamp

# Environment variables (redact secrets)
kubectl -n cloudshell-system exec deploy/cloudshell-gateway -- env | grep -v KEY | sort
```

---

## Still stuck?

- Search [existing issues](https://github.com/SocioProphet/cloudshell-fog/issues) — your problem may already have a solution.
- Open a [new bug report](https://github.com/SocioProphet/cloudshell-fog/issues/new?template=bug_report.yml) with the debug information above.
