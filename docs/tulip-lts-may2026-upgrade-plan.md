# Tulip LTS May 2026 — Upgrade Plan

## What is an LTS release?

LTS (Long-Term Support) is a release model where a specific version receives extended maintenance, security patches, and bug fixes over a longer period than regular releases. Unlike the upstream OpenTelemetry Collector, which ships new versions roughly every two weeks, an LTS release provides a **stable, validated baseline** that production environments can depend on without the risk of frequent breaking changes.

### Why ship an LTS?

The upstream OpenTelemetry Collector moves fast — new releases every ~2 weeks, components changing stability levels, deprecations, and breaking changes. This velocity is great for innovation but creates challenges for production deployments:

- **Upgrade fatigue:** Keeping up with biweekly releases is unsustainable for teams running the collector at scale.
- **Stability risk:** Not every upstream release is equally battle-tested. Some introduce regressions that get fixed in the next release.
- **Component churn:** Components get deprecated (e.g., batchprocessor), replaced, or have their APIs changed. Teams need time to plan migrations.
- **Support burden:** Supporting arbitrary collector versions is impractical. An LTS gives a defined, tested target.

The Tulip LTS provides:

- **A curated, validated component set** — every component is reviewed for stability, active maintenance, and known issues before inclusion.
- **A predictable upgrade path** — instead of chasing every upstream release, teams upgrade LTS-to-LTS with clear migration documentation.
- **Production-ready defaults** — the default configuration includes OOM protection, retry logic, and durable queuing out of the box.
- **Extended support window** — security patches and critical fixes backported without requiring a full version upgrade.

## Overview

This document defines the component composition and configuration changes for the Tulip LTS release, targeting May 2026. The LTS follows a **stability-first** approach: only stable/beta components with active maintainers are included.

**Version bump:** v0.145.0 → v0.150.0 (all components)

The target version v0.150.0 was selected as the latest stable upstream release at the time of this LTS cut, verified against the official otelcol-contrib distribution manifest:
- https://github.com/open-telemetry/opentelemetry-collector-releases/blob/main/distributions/otelcol-contrib/manifest.yaml

---

## Component Changes

### Removed: batchprocessor

**Reason:** The batchprocessor has been formally deprecated in the OpenTelemetry Collector (PR [#15046](https://github.com/open-telemetry/opentelemetry-collector/pull/15046), April 2026). It also has a known data loss bug ([#12443](https://github.com/open-telemetry/opentelemetry-collector/issues/12443)) where data is silently dropped when a downstream exporter rejects and has no queue/retry configured.

**Replacement:** Exporter-level batching via `sending_queue` + `batch` configuration on each exporter. This is the officially recommended path and provides stronger delivery guarantees because:

- Data is durably enqueued in the exporter's persistent sending queue before acknowledgment
- Batching and queueing are consolidated within the exporter
- Data can be written to disk and recovered after a Collector restart
- No silent data loss — failed sends are retried with backoff

#### Migration: before and after

**Before (batchprocessor in pipeline):**

```yaml
processors:
  batch:
    send_batch_size: 8192
    timeout: 200ms

exporters:
  otlp:
    endpoint: otlp.example.com:4317

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlp]
```

**After (exporter-level batching):**

```yaml
exporters:
  otlp:
    endpoint: otlp.example.com:4317
    sending_queue:
      enabled: true
      queue_size: 1000
    batch:
      flush_timeout: 200ms
      min_size: 8192
    retry_on_failure:
      enabled: true
      initial_interval: 5s
      max_interval: 30s
      max_elapsed_time: 300s

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: []
      exporters: [otlp]
```

#### References

- Deprecation PR: https://github.com/open-telemetry/opentelemetry-collector/pull/15046
- Data loss bug: https://github.com/open-telemetry/opentelemetry-collector/issues/12443
- Resolution discussion: https://github.com/open-telemetry/opentelemetry-collector/issues/15047
- Docs removal: https://github.com/open-telemetry/opentelemetry-collector/issues/13766

---

### Added: memorylimiterprocessor

**Reason:** Prevents the collector from being OOM-killed in production. Without it, memory spikes from telemetry bursts or slow downstream exporters cause unbounded memory growth, leading to pod eviction (Kubernetes) or host destabilization.

This is a core component (not contrib), stable maturity, and recommended by every major OTel production deployment guide.

**Configuration example:**

```yaml
processors:
  memory_limiter:
    check_interval: 1s
    limit_mib: 1500
    spike_limit_mib: 512

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [memory_limiter]
      exporters: [otlp]
```

**Note:** `memory_limiter` should be the first processor in every pipeline to ensure backpressure is applied before any processing work is done.

---

## Final LTS Component Manifest (25 components)

### Extensions (7)

| Component | Source | Stability |
|-----------|--------|-----------|
| zpagesextension | core | stable |
| healthcheckextension | contrib | beta |
| pprofextension | contrib | beta |
| basicauthextension | contrib | beta |
| bearertokenauthextension | contrib | beta |
| oauth2clientauthextension | contrib | beta |
| oidcauthextension | contrib | beta |

### Receivers (2)

| Component | Source | Stability |
|-----------|--------|-----------|
| otlpreceiver | core | stable |
| nopreceiver | core | stable |

### Exporters (5)

| Component | Source | Stability |
|-----------|--------|-----------|
| debugexporter | core | stable |
| nopexporter | core | stable |
| otlpexporter | core | stable |
| otlphttpexporter | core | stable |
| fileexporter | contrib | beta |

### Processors (7)

| Component | Source | Stability | Note |
|-----------|--------|-----------|------|
| memorylimiterprocessor | core | stable | NEW |
| attributesprocessor | contrib | stable | |
| resourceprocessor | contrib | stable | |
| spanprocessor | contrib | stable | |
| probabilisticsamplerprocessor | contrib | stable | |
| filterprocessor | contrib | stable | |
| transformprocessor | contrib | stable | |

### Connectors (1)

| Component | Source | Stability |
|-----------|--------|-----------|
| forwardconnector | core | stable |

### Providers (3)

| Component | Source | Stability |
|-----------|--------|-----------|
| envprovider | core | stable |
| fileprovider | core | stable |
| yamlprovider | core | stable |

