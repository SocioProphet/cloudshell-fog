# EF-0005 — `internal/api/handler.go` hookup patch

## Purpose

The branch already contains:
- `internal/api/placement_response.go`
- `internal/api/placement_response_test.go`

The remaining code change is to use the richer response helper in `CreateSession` when returning the session creation payload.

## Current block

```go
writeJSON(w, http.StatusCreated, map[string]any{
	"session_id": sessionID,
	"attach": map[string]any{
		"ws_url":     fmt.Sprintf("%s/v1/sessions/%s/pty", h.gatewayURL, sessionID),
		"token":      token,
		"expires_at": tokenExpiresAt.UTC().Format(time.RFC3339),
	},
	"placement": decision.Region,
})
```

## Replace with

```go
writeJSON(w, http.StatusCreated, BuildCreateSessionResponseVNext(
	sessionID,
	h.gatewayURL,
	token,
	tokenExpiresAt,
	decision,
))
```

## Effect

This preserves current attach semantics while enriching `placement` from a plain region string to a structured object:
- `region`
- `node_id`
- `tier`
- `reasons`

## Follow-on

After this hook-up lands, the OpenAPI schema should either:
- be updated to the richer placement object shape, or
- versioned explicitly if backward compatibility requires both forms.
