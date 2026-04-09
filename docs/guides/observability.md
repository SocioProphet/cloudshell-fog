# Observability

cloudshell-fog emits traces, metrics, and structured audit logs using [OpenTelemetry](https://opentelemetry.io). This guide explains what is emitted, how to collect it, and how to query it.

## What is emitted

### Traces

Every significant operation is wrapped in an OpenTelemetry span:

| Span | Description |
|---|---|
| `session.create` | Full session creation flow (auth → policy → placement → connector → audit) |
| `session.get` | Session status lookup |
| `session.delete` | Session termination |
| `pty.attach` | PTY WebSocket lifecycle (from upgrade to disconnect) |

Spans include attributes for `session_id`, `subject`, `placement`, and `profile`.

### Metrics

| Metric | Type | Description |
|---|---|---|
| `cloudshell_sessions_created_total` | Counter | Sessions created since startup |
| `cloudshell_sessions_active` | Gauge | Currently running sessions |
| `cloudshell_session_duration_seconds` | Histogram | Session duration at termination |
| `cloudshell_placement_decisions_total` | Counter | Placement decisions by tier (`fog`, `cloud`) |
| `cloudshell_policy_denials_total` | Counter | Policy denials by reason |
| `cloudshell_pty_attach_duration_seconds` | Histogram | PTY session duration |

### Audit logs

The audit emitter writes structured JSON events via `slog` to stdout. Each event has the shape:

```json
{
  "ts": "2026-01-01T00:00:00Z",
  "session_id": "550e8400-...",
  "subject": "user@example.com",
  "type": "session.created",
  "placement": "eu-west-1",
  "details": { "profile": "default", "image_ref": "..." }
}
```

Audit event types:

| Type | Trigger |
|---|---|
| `session.created` | New session successfully created |
| `session.attached` | PTY WebSocket connection established |
| `session.terminated` | Session deleted or TTL expired |
| `placement.decided` | Placement engine selected a node |
| `runtime.allocated` | Connector provisioned a runtime pod |
| `policy.denied` | Admission policy rejected a session request |

---

## Default exporter (stdout)

By default, the gateway uses the **OTel stdout exporter** — it writes OTLP-formatted JSON to stdout. This is zero-dependency and suitable for collection by log aggregators (Fluentd, Fluentbit, Loki).

Tail the logs in Kubernetes:

```bash
kubectl -n cloudshell-system logs -f deployment/cloudshell-gateway
```

---

## Sending to an OTel Collector

To export to a real backend (Jaeger, Tempo, Prometheus, Grafana), deploy an [OpenTelemetry Collector](https://opentelemetry.io/docs/collector/) and configure the gateway to send to it via the OTLP exporter.

### 1. Deploy the OTel Collector

A minimal collector config (`otel-collector-config.yaml`):

```yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318

exporters:
  jaeger:
    endpoint: jaeger-collector.monitoring:14250
    tls:
      insecure: true
  prometheus:
    endpoint: 0.0.0.0:8889

service:
  pipelines:
    traces:
      receivers: [otlp]
      exporters: [jaeger]
    metrics:
      receivers: [otlp]
      exporters: [prometheus]
```

### 2. Configure the gateway

Set these environment variables in `deploy/k8s/deployment.yaml`:

```yaml
- name: OTEL_EXPORTER_OTLP_ENDPOINT
  value: "http://otel-collector.monitoring:4318"
- name: OTEL_SERVICE_NAME
  value: "cloudshell-fog"
```

> **Note:** The gateway's OTel initialisation (`internal/otel`) currently uses stdout exporters. To switch to OTLP, update `internal/otel/otel.go` to use `go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp` and `otlpmetrichttp`. This is a planned improvement (see [CHANGELOG.md](../../CHANGELOG.md)).

---

## Prometheus scraping

If using Prometheus, expose the metrics port in the Service and add a `ServiceMonitor` or static scrape config:

```yaml
# prometheus-scrape.yaml
- job_name: cloudshell-fog
  static_configs:
    - targets: ['cloudshell-gateway.cloudshell-system:9090']
```

---

## Useful queries

### Active sessions (PromQL)

```promql
cloudshell_sessions_active
```

### Session creation rate (per minute)

```promql
rate(cloudshell_sessions_created_total[1m])
```

### Placement fog vs cloud ratio

```promql
sum by (tier) (rate(cloudshell_placement_decisions_total[5m]))
```

### Policy denial rate

```promql
rate(cloudshell_policy_denials_total[5m])
```

### P95 session duration

```promql
histogram_quantile(0.95, rate(cloudshell_session_duration_seconds_bucket[5m]))
```

---

## Log aggregation

Because audit logs are structured JSON on stdout, they work natively with:

- **Fluentbit** — use the `tail` input and the `json` parser
- **Loki** — use the Kubernetes log discovery with `json` pipeline stage
- **Elasticsearch** — use Filebeat or Fluentd with the JSON codec

Example Fluentbit filter to extract `session_id` as a field:

```ini
[FILTER]
    Name   parser
    Match  cloudshell.*
    Key_Name log
    Parser json
```

---

## Related

- [Troubleshooting](troubleshooting.md) — diagnosing issues using logs and traces
- [Architecture](../architecture.md) — how `internal/audit` and `internal/otel` fit into the gateway
