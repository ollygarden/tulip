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
- **Production-ready defaults** — the default configuration includes OOM protection, retry logic.
- **Extended support window** — security patches and critical fixes backported without requiring a full version upgrade.

## Overview

This document defines the component composition and configuration changes for the Tulip LTS release, targeting May 2026. The LTS follows a **stability-first** approach: only stable/beta components with active maintainers are included.

**Version bump:** v0.145.0 → v0.151.0 (all components)

The target version v0.151.0 was selected as the latest stable upstream release at the time of this LTS cut, verified against the official otelcol-contrib distribution manifest:
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

### Removed: healthcheckextension

**Reason:** The healthcheck extension v1 and v2 share code between them using a feature flag ([#42256](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/42256)). Including v1 would indirectly support v2, which is not the objective for an LTS release where component behavior must be fully predictable and stable.

---

### Added: redactionprocessor

**Reason:** Allows redacting sensitive data from telemetry attributes before export. Essential for enterprise environments with data privacy requirements (PII, HIPAA, GDPR). This is a contrib component with stable maturity.

---

### Added: filestorage extension

**Reason:** Provides persistent file-based storage for components that need durable state across collector restarts (e.g., exporter sending queues, receiver checkpoints). Critical for reliable telemetry delivery in production.

---

### Added: hostmetricsreceiver

**Reason:** Collects host-level metrics (CPU, memory, disk, network, filesystem, processes) from the machine running the collector. Essential for infrastructure monitoring use cases where the collector also serves as a host metrics agent.

---

### Added: filelogreceiver

**Reason:** Reads and parses log data from files on the host. Enables log collection pipelines where applications write to local files. Widely used in production for collecting application logs, system logs, and container logs.

---

## Final LTS Component Manifest (28 components)

### Extensions (7)

| Component | Source | Stability | Note |
|-----------|--------|-----------|------|
| zpagesextension | core | stable | |
| pprofextension | contrib | beta | |
| basicauthextension | contrib | beta | |
| bearertokenauthextension | contrib | beta | |
| oauth2clientauthextension | contrib | beta | |
| oidcauthextension | contrib | beta | |
| filestorage | contrib | beta | NEW |

### Receivers (4)

| Component | Source | Stability | Note |
|-----------|--------|-----------|------|
| otlpreceiver | core | stable | |
| nopreceiver | core | stable | |
| hostmetricsreceiver | contrib | beta | NEW |
| filelogreceiver | contrib | beta | NEW |

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
| attributesprocessor | contrib | stable | |
| resourceprocessor | contrib | stable | |
| spanprocessor | contrib | stable | |
| probabilisticsamplerprocessor | contrib | stable | |
| filterprocessor | contrib | stable | |
| transformprocessor | contrib | stable | |
| redactionprocessor | contrib | stable | NEW |

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

