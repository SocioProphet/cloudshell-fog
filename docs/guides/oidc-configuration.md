# OIDC Configuration

This guide explains how to configure cloudshell-fog to validate access tokens from an OpenID Connect (OIDC) provider. The gateway supports any standards-compliant provider — Keycloak, Dex, Okta, Auth0, Google, and others.

## How OIDC validation works

When `OIDC_ISSUER_URL` and `OIDC_CLIENT_ID` are set, the gateway:

1. Fetches the provider's JWKS endpoint from `{OIDC_ISSUER_URL}/.well-known/openid-configuration` at startup.
2. On every session request, validates the `Authorization: Bearer <token>` against the JWKS.
3. Extracts the `sub` (subject) claim as the user identity.
4. Extracts the `groups` claim (if present) for policy group-based access control.

When neither variable is set, the gateway falls back to the **dev auth shim**, which accepts any token and injects a fixed `subject=dev-user` — suitable for local development only.

## Environment variables

| Variable | Description |
|---|---|
| `OIDC_ISSUER_URL` | The issuer URL of your OIDC provider (e.g. `https://accounts.example.com`). The gateway appends `/.well-known/openid-configuration` to discover JWKS. |
| `OIDC_CLIENT_ID` | The OAuth 2.0 client ID registered with your provider. Used to validate the `aud` claim. |

---

## Keycloak

### 1. Create a realm and client

1. Log in to the Keycloak admin console.
2. Create a realm (e.g. `cloudshell`).
3. Under **Clients**, create a new client:
   - **Client ID**: `cloudshell-gateway`
   - **Client Protocol**: `openid-connect`
   - **Access Type**: `confidential` (for backend flows) or `public` (for SPA flows)
4. Under **Client Scopes**, ensure the `groups` scope is included in the default scopes (required for group-based policy).

### 2. Configure the gateway

```bash
OIDC_ISSUER_URL=https://keycloak.example.com/realms/cloudshell
OIDC_CLIENT_ID=cloudshell-gateway
```

### 3. Map groups to the token

In Keycloak, add a **Group Membership** mapper to the client:

- **Mapper Type**: Group Membership
- **Token Claim Name**: `groups`
- **Full group path**: off (use short names like `admins`, not `/admins`)

This populates the `groups` claim in the access token, which the policy engine reads.

---

## Dex

[Dex](https://dexidp.io) is a popular self-hosted OIDC provider that federates to upstream identity sources (LDAP, GitHub, SAML, etc.).

### 1. Register a client in dex config

```yaml
# config.yaml
staticClients:
  - id: cloudshell-gateway
    secret: <client-secret>
    name: cloudshell-fog
    redirectURIs:
      - https://shell.example.com/callback
```

### 2. Configure the gateway

```bash
OIDC_ISSUER_URL=https://dex.example.com
OIDC_CLIENT_ID=cloudshell-gateway
```

### 3. Groups claim

Dex propagates groups from the upstream connector. For the GitHub connector, groups correspond to GitHub organisation teams. For LDAP, they correspond to LDAP groups. Ensure the `groups` scope is requested in the authorization request.

---

## Okta

### 1. Create an application

1. In the Okta developer console, go to **Applications → Create App Integration**.
2. Choose **OIDC – OpenID Connect** and **Web Application**.
3. Set the **Sign-in redirect URI** to your shell URL.

### 2. Add a `groups` claim

1. Go to **Security → API → Authorization Servers → default**.
2. Under **Claims**, add a claim:
   - **Name**: `groups`
   - **Include in token type**: Access Token
   - **Value type**: Groups
   - **Filter**: Matches regex `.*`

### 3. Configure the gateway

```bash
OIDC_ISSUER_URL=https://your-org.okta.com/oauth2/default
OIDC_CLIENT_ID=<your-client-id>
```

---

## Kubernetes Secret injection

For production deployments, inject the OIDC configuration from a Kubernetes Secret rather than hard-coding values in the Deployment manifest:

```bash
kubectl -n cloudshell-system create secret generic cloudshell-oidc \
  --from-literal=oidc-issuer-url=https://accounts.example.com \
  --from-literal=oidc-client-id=cloudshell-gateway
```

Reference the secret in `deploy/k8s/deployment.yaml`:

```yaml
env:
  - name: OIDC_ISSUER_URL
    valueFrom:
      secretKeyRef:
        name: cloudshell-oidc
        key: oidc-issuer-url
  - name: OIDC_CLIENT_ID
    valueFrom:
      secretKeyRef:
        name: cloudshell-oidc
        key: oidc-client-id
```

---

## Verifying OIDC is working

After starting the gateway with OIDC configured:

1. Obtain an access token from your provider (e.g. via `curl`, Postman, or a test client).
2. Make a request to the sessions endpoint:

   ```bash
   curl -s -X POST https://shell.example.com/v1/sessions \
     -H "Authorization: Bearer $ACCESS_TOKEN" \
     -H "Content-Type: application/json" \
     -d '{"profile":"default","ttl_seconds":600}' | jq .
   ```

3. If you see `401 Unauthorized`, check the gateway logs:

   ```bash
   kubectl -n cloudshell-system logs deployment/cloudshell-gateway | grep "auth"
   ```

   Common causes:
   - `iss` claim does not match `OIDC_ISSUER_URL`
   - `aud` claim does not contain `OIDC_CLIENT_ID`
   - Token is expired or the clock is skewed
   - JWKS endpoint is not reachable from the gateway pod

For additional troubleshooting steps see [Troubleshooting](troubleshooting.md).

---

## Security notes

- The gateway validates every session-management request independently — tokens are never stored or cached beyond the JWKS key material.
- Session tokens (minted after successful OIDC validation) are short-lived HMAC-signed JWTs scoped to a single session. They cannot be used to create new sessions.
- The `groups` claim is trusted as-is from the OIDC token. Ensure your OIDC provider is the authoritative source for group membership and that group names in `config/policy.yaml` match those in the token exactly.
