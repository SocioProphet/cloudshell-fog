# EF-0009 — Session namespace bootstrap mechanism

## Purpose

The repository previously described per-session namespace isolation, but the checked-in NetworkPolicy example in `deploy/k8s/networkpolicy.yaml` was still anchored to `cloudshell-system` and behaved more like a template than a true per-session mechanism.

This note defines the intended bootstrap flow for each session namespace.

## Recommended flow

When the runtime connector provisions a new namespace:

1. create namespace `cloudshell-<sessionID>`
2. label namespace with at least:
   - `cloudshell.io/session-id=<sessionID>`
   - `cloudshell.io/managed=true`
3. apply the session namespace policy pack from:
   - `deploy/k8s/session-namespace/`
4. create the session pod inside that namespace

## Why the policy pack is namespace-scoped

The files under `deploy/k8s/session-namespace/` intentionally omit `metadata.namespace` so they can be applied into whichever namespace has just been created.

This is the missing bridge between:
- the architecture/docs promise of per-session isolation
- the operational reality of a runtime connector creating dynamic namespaces

## Implementation options

### Option A — connector applies resources directly
The Kubernetes connector creates the namespace and then applies the policy objects via the Kubernetes API.

### Option B — bootstrap job/controller
The connector creates the namespace with labels/annotations and a separate bootstrap controller or job applies the standard pack.

## Recommendation

For v0/v1 maturity:
- use **Option A** inside the Kubernetes connector because it is the smallest operational step
- move to Option B only if we later want stricter separation of duties
