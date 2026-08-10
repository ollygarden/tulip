# Tulip v26.08 — Release Notes

**Version:** v26.08.0  
**Release Date:** August 2026  
**Component Base Version:** v0.151.0

This is a regular quarterly release (not LTS). It follows the standard quarterly cadence: v26.05.0 (LTS) → v26.08.0 → v26.11.0 → v27.02.0 → ...

---

## Component Changes from v26.05.1

### Added: drainprocessor

**Type:** Processor (logs)  
**Stability:** Alpha  
**Source:** opentelemetry-collector-contrib v0.151.0  
**Codeowners:** MikeGoldsmith, atoulme, martinjt  

**Purpose:** Applies the Drain log-clustering algorithm to derive log message templates from raw log bodies. Groups similar logs by learning patterns (wildcards `<*>`), attaching the learned template as a `log.record.template` attribute. Essential for log reduction, grouping, and anomaly detection.

**When to use:** 
- Reduce high-cardinality log streams by clustering similar messages
- Enable aggregation and alerting on log templates instead of raw bodies
- Feed pipeline data to logdedup processor for per-template deduplication

**Configuration example:**
```yaml
processors:
  drain:
    drain:
      algorithm: drain
      tau: 0.5  # Cluster similarity threshold (0-1)
```

**Integration note:** Works best as a pipeline stage *before* logdedup processor. See E-2660 for full evaluation and chaining guidance.

---

### Added: logdedupprocessor

**Type:** Processor (logs)  
**Stability:** Alpha  
**Source:** opentelemetry-collector-contrib v0.151.0  
**Codeowners:** MikeGoldsmith (emeritus: djaglowski)  

**Purpose:** Collapses identical logs over a time window (`interval`), emitting a single record with a count attribute (configurable name, e.g., `dedup_count`). Reduces cardinality of identical log streams. Supports conditional deduplication (OTTL conditions), per-tenant metadata keys, and field inclusion/exclusion for flexible dedup key derivation.

**When to use:**
- Reduce storage and costs for high-volume identical logs (e.g., health checks, periodic warnings)
- Count duplicate log occurrences in aggregate views
- Track grouped logs over time windows (e.g., "3 identical connection errors in 5s")

**Configuration example:**
```yaml
processors:
  logdedup:
    loglevel: info
    interval: 5s
    log_count_attribute: dedup_count
    include_fields:
      - attributes.log.record.template  # IMPORTANT: when chaining after drain
    exclude_fields: []
    condition:
      match_type: regexp
      regexp: '.*ERROR.*'  # optional: only dedup ERROR level logs
```

**Integration note:** By default, dedup key is derived from `body + resource attributes + severity + log attributes`. When chaining after drainprocessor, **explicitly configure `include_fields: [attributes.log.record.template]`** to dedupe on the template instead of raw body. See E-2660 for full evaluation and gotchas.

---

## Full Component Manifest (29 components)

### Extensions (7)

| Component | Stability | Version |
|-----------|-----------|---------|
| zpagesextension | stable | v0.151.0 |
| pprofextension | beta | v0.151.0 |
| basicauthextension | beta | v0.151.0 |
| bearertokenauthextension | beta | v0.151.0 |
| oauth2clientauthextension | beta | v0.151.0 |
| oidcauthextension | beta | v0.151.0 |
| filestorage | beta | v0.151.0 |

### Receivers (4)

| Component | Stability | Version |
|-----------|-----------|---------|
| nopreceiver | stable | v0.151.0 |
| otlpreceiver | stable | v0.151.0 |
| hostmetricsreceiver | beta | v0.151.0 |
| filelogreceiver | beta | v0.151.0 |

### Exporters (5)

| Component | Stability | Version |
|-----------|-----------|---------|
| debugexporter | stable | v0.151.0 |
| nopexporter | stable | v0.151.0 |
| otlpexporter | stable | v0.151.0 |
| otlphttpexporter | stable | v0.151.0 |
| fileexporter | beta | v0.151.0 |

### Processors (9)

| Component | Stability | Version | Note |
|-----------|-----------|---------|------|
| attributesprocessor | stable | v0.151.0 | |
| resourceprocessor | stable | v0.151.0 | |
| spanprocessor | stable | v0.151.0 | |
| probabilisticsamplerprocessor | stable | v0.151.0 | |
| filterprocessor | stable | v0.151.0 | |
| transformprocessor | stable | v0.151.0 | |
| redactionprocessor | stable | v0.151.0 | |
| drainprocessor | alpha | v0.151.0 | NEW |
| logdedupprocessor | alpha | v0.151.0 | NEW |

### Connectors (1)

| Component | Stability | Version |
|-----------|-----------|---------|
| forwardconnector | stable | v0.151.0 |

### Providers (3)

| Component | Stability | Version |
|-----------|-----------|---------|
| envprovider | stable | v1.57.0 |
| fileprovider | stable | v1.57.0 |
| yamlprovider | stable | v1.57.0 |

---

## Migration Guide

No breaking changes from v26.05.1 to v26.08.0. All existing pipelines remain compatible.

### Optional: Enable log dedup + clustering

To start using the new processors, add them to a log pipeline:

```yaml
receivers:
  filelog:
    include_paths: [/var/log/*.log]

processors:
  drain:
    drain:
      algorithm: drain
      tau: 0.5
  logdedup:
    interval: 5s
    log_count_attribute: dedup_count
    include_fields:
      - attributes.log.record.template

exporters:
  otlp:
    endpoint: localhost:4317

service:
  pipelines:
    logs:
      receivers: [filelog]
      processors: [drain, logdedup]
      exporters: [otlp]
```

---

## References

- **Evaluation & Integration Testing:** E-2660
- **Drain Processor Upstream:** [opentelemetry-collector-contrib](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/drainprocessor)
- **LogDedup Processor Upstream:** [opentelemetry-collector-contrib](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/logdedupprocessor)
