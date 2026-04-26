# API Reference

This is the complete HTTP and WebSocket API reference for cloudshell-fog. All endpoints are served by the gateway at the configured `LISTEN_ADDR` (default `:8080`).

## Authentication

All session-management endpoints (`/v1/sessions*`) require an OIDC access token:

```
Authorization: Bearer <oidc-access-token>
```

The PTY WebSocket endpoint uses a short-lived session token passed as a query parameter (see [PTY attach](#get-v1sessionsidpty) below).

In development mode (no `OIDC_ISSUER_URL` set), any non-empty token is accepted.

---

## Health check

### `GET /healthz`

Returns gateway liveness status. Does not require authentication.

**Response `200 OK`**

```json
{ "status": "ok" }
```

---

## Session endpoints

### `POST /v1/sessions`

Create a new shell session.

**Request headers**

| Header | Value |
|---|---|
| `Authorization` | `Bearer <oidc-access-token>` |
| `Content-Type` | `application/json` |

**Request body**

| Field | Type | Required | Description |
|---|---|---|---|
| `profile` | string | Yes | Resource profile name (must exist in `config/policy.yaml`) |
| `ttl_seconds` | integer | Yes | Session lifetime in seconds (must be ≤ profile's `max_ttl_seconds`) |
| `placement_hint` | string | No | Preferred region or node ID (e.g. `eu-west-1`) |
| `image_ref` | string | No | OCI image for the session runtime (production: must be a pinned digest) |

**Example request**

```bash
curl -s -X POST https://shell.example.com/v1/sessions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "profile": "default",
    "ttl_seconds": 3600,
    "placement_hint": "eu-west-1",
    "image_ref": "ghcr.io/socioprophet/cloudshell-runtime@sha256:abc123..."
  }' | jq .
```

**Response `201 Created`**

```json
{
  "session_id": "550e8400-e29b-41d4-a716-446655440000",
  "attach": {
    "ws_url": "wss://shell.example.com/v1/sessions/550e8400-.../pty",
    "token": "<short-lived JWT>",
    "expires_at": "2026-01-01T00:15:00Z"
  },
  "placement": {
    "region": "eu-west-1",
    "node_id": "fog-node-1",
    "tier": "fog",
    "reasons": ["fog-preferred", "healthy", "capacity-ok"]
  }
}
```

| Field | Description |
|---|---|
| `session_id` | UUID identifying the session |
| `attach.ws_url` | Full WebSocket URL for PTY attach |
| `attach.token` | Short-lived HMAC JWT valid for 15 minutes, scoped to this session only |
| `attach.expires_at` | Token expiry time (RFC3339) |
| `placement.region` | Region selected by the placement engine |
| `placement.node_id` | Concrete runtime node selected by the placement engine |
| `placement.tier` | Placement trust tier such as `fog` or `cloud` |
| `placement.reasons` | Human-readable decision reasons emitted by the placement engine |

**Error responses**

| Status | Cause |
|---|---|
| `400 Bad Request` | Invalid or missing request body fields |
| `401 Unauthorized` | Missing, invalid, or expired OIDC token |
| `403 Forbidden` | Policy denied: group not allowed, quota exceeded, or TTL too long |
| `500 Internal Server Error` | Connector or placement failure (check gateway logs) |

---

### `GET /v1/sessions/{id}`

Get the status of an existing session.

**Request headers**

| Header | Value |
|---|---|
| `Authorization` | `Bearer <oidc-access-token>` |

**Example request**

```bash
curl -s https://shell.example.com/v1/sessions/550e8400-e29b-41d4-a716-446655440000 \
  -H "Authorization: Bearer $TOKEN" | jq .
```

**Response `200 OK`**

```json
{
  "session_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "running",
  "placement": {
    "region": "eu-west-1",
    "node_id": "fog-node-1",
    "tier": "fog",
    "reasons": ["fog-preferred", "healthy", "capacity-ok"]
  },
  "created_at": "2026-01-01T00:00:00Z",
  "expires_at": "2026-01-01T01:00:00Z",
  "image_ref": "ghcr.io/socioprophet/cloudshell-runtime@sha256:abc123..."
}
```

| Field | Description |
|---|---|
| `session_id` | UUID identifying the session |
| `status` | One of: `pending`, `running`, `terminated` |
| `placement.region` | Region where the runtime was placed |
| `placement.node_id` | Concrete runtime node selected for the session |
| `placement.tier` | Placement trust tier such as `fog` or `cloud` |
| `placement.reasons` | Stored decision reasons when available |
| `created_at` | Session creation timestamp (RFC3339) |
| `expires_at` | Session expiry timestamp (RFC3339) |
| `image_ref` | Runtime image reference |

**Error responses**

| Status | Cause |
|---|---|
| `401 Unauthorized` | Missing, invalid, or expired OIDC token |
| `404 Not Found` | Session does not exist or has already been terminated |

---

### `DELETE /v1/sessions/{id}`

Terminate a session and release its runtime resources.

**Request headers**

| Header | Value |
|---|---|
| `Authorization` | `Bearer <oidc-access-token>` |

**Example request**

```bash
curl -s -X DELETE \
  https://shell.example.com/v1/sessions/550e8400-e29b-41d4-a716-446655440000 \
  -H "Authorization: Bearer $TOKEN" | jq .
```

**Response `200 OK`**

```json
{ "terminated": true }
```

**Error responses**

| Status | Cause |
|---|---|
| `401 Unauthorized` | Missing, invalid, or expired OIDC token |
| `404 Not Found` | Session does not exist |

---

## PTY WebSocket

### `GET /v1/sessions/{id}/pty`

Upgrade to a WebSocket connection and attach an interactive PTY.

**Authentication:** Pass the short-lived `token` from the `POST /v1/sessions` response as a query parameter (not in the `Authorization` header):

```
wss://shell.example.com/v1/sessions/{id}/pty?token=<short-lived-JWT>
```

This token is intentionally separate from the OIDC token — it is short-lived (15 min), session-scoped, and cannot be used to create new sessions. This allows the browser to pass it safely in a WebSocket URL.

**WebSocket subprotocol:** none required.

### Frame schema

All messages are JSON-encoded text frames.

**Client → Server frames**

| Frame | Description |
|---|---|
| `{"type":"stdin","data_b64":"<base64>"}` | Terminal input; `data_b64` is the base64-encoded stdin bytes |
| `{"type":"resize","cols":220,"rows":50}` | Terminal resize; sends `SIGWINCH` to the shell process |

**Server → Client frames**

| Frame | Description |
|---|---|
| `{"type":"stdout","data_b64":"<base64>"}` | Terminal output; `data_b64` is the base64-encoded stdout/stderr bytes |
| `{"type":"exit","code":0}` | Shell process exited; `code` is the exit status |

### JavaScript example

```javascript
// After creating a session, attach the PTY
const { ws_url, token } = session.attach;
const ws = new WebSocket(`${ws_url}?token=${token}`);

ws.onmessage = (event) => {
  const frame = JSON.parse(event.data);
  if (frame.type === 'stdout') {
    const text = atob(frame.data_b64);
    terminal.write(text);            // xterm.js
  } else if (frame.type === 'exit') {
    console.log('Shell exited with code', frame.code);
    ws.close();
  }
};

// Send stdin
terminal.onData((data) => {
  ws.send(JSON.stringify({
    type: 'stdin',
    data_b64: btoa(data),
  }));
});

// Send resize
terminal.onResize(({ cols, rows }) => {
  ws.send(JSON.stringify({ type: 'resize', cols, rows }));
});
```

---

## Pagination and rate limits

The current API does not implement pagination or rate limiting. These are planned for a future release.

---

## Related

- [Configuration Reference](configuration.md) — environment variables and policy schema
- [Getting Started](../guides/getting-started.md) — end-to-end walkthrough
- [Troubleshooting](../guides/troubleshooting.md) — diagnosing API errors
