# Post-Merge Validation Checklist v0 — cloudshell-fog

Use this checklist after merging the current PR train.

## A. Gateway / connector path

- [ ] `CONNECTOR_MODE` defaults to `stub`
- [ ] `CONNECTOR_MODE=stub` starts cleanly
- [ ] `CONNECTOR_MODE=k8s` without valid config fails fast at startup
- [ ] `CONNECTOR_MODE=k8s` with valid kubeconfig or in-cluster config selects Kubernetes connector

## B. Runtime image resolution

- [ ] explicit API `image_ref` overrides env/config defaults
- [ ] `RUNTIME_IMAGE_REF` environment variable is respected when explicit request field is absent
- [ ] canonical code default is used when neither is provided

## C. PTY contract

- [ ] PTY attach still requires valid session-scoped token
- [ ] `stdin/stdout/resize/exit` frame behavior remains unchanged

## D. Deployment / GitOps

- [ ] Argo Applications exist for main k8s deployment, policy resources, and Tekton resources
- [ ] production overlay guidance is visible and points to digest usage
- [ ] no production doc path implies `:latest` is acceptable

## E. Session namespace isolation

- [ ] session namespace policy pack exists
- [ ] namespace bootstrap mechanism is documented
- [ ] per-session namespace policy application path is chosen (connector-applied or controller-applied)

## F. Supply-chain trust

- [ ] signed-image trust model doc is present
- [ ] keyless verification policy variant is present
- [ ] placeholder-key policy is clearly distinguished from production recommendation

## G. Security / ADR memory

- [ ] threat model present
- [ ] runtime isolation doc present
- [ ] ADR set present

## H. Follow-on engineering queue

After merge, the next engineering wave should focus on:
1. replacing placeholder trust material with real production trust configuration
2. reconciling dynamic namespace bootstrap in connector code
3. tightening API response shape vs richer placement semantics
4. deciding whether to physically rewrite base deployment manifests away from dev-oriented examples
