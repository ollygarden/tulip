# Tulip v26.08 — Comprehensive Upgrade & Deployment Guide

**Version:** v26.08.0  
**Release Date:** August 2026  
**Component Base Version:** v0.151.0 (OTel Collector core/contrib)  
**Previous LTS:** v26.05.1 (May 2026)

---

## Executive Summary

Tulip v26.08.0 is a **regular quarterly release** (not LTS) with one significant addition:
- **logdedupprocessor** (alpha) — log deduplication with time-window aggregation

**Upgrade Path:** v26.05.1 LTS → v26.08.0 is safe. All 27 components are actively maintained. Two components (debugexporter, spanprocessor) require attention for production use.

---

## What's New

### Added Components (1)

#### logdedupprocessor
- **Type:** Processor (logs)
- **Stability:** Alpha
- **Purpose:** Collapses identical logs over a time window, emits count attribute
- **Use case:** Reduce storage costs for identical repeating logs

See **Integration Guide** (below) for configuration examples.

---

## Component Health Audit

### Summary

| Health Level | Count | Assessment |
|---|---|---|
| **Production-Ready** | 18 | Stable/BETA, actively maintained |
| **Ready w/ Testing** | 7 | BETA with caution (load-test filestorage, auth extensions) |
| **Requires Attention** | 2 | debugexporter (dev-only), spanprocessor (single maintainer) |

### All Components (v0.151.0)

#### Extensions (7)

| Component | Stability | Maintainers | Status | Risk |
|-----------|-----------|-------------|--------|------|
| zpagesextension | STABLE | OTel Core Team | Production-ready | ✅ |
| pprofextension | BETA | MovieStoreGuy | Actively maintained | ⚠️ |
| basicauthextension | BETA | frzifus | **HIGH bus factor** | ⚠️ |
| bearertokenauthextension | BETA | frzifus | **HIGH bus factor** | ⚠️ |
| oauth2clientauthextension | BETA | pavankrish123 | Single owner | ⚠️ |
| oidcauthextension | BETA | asweet-confluent | Single owner | ⚠️ |
| filestorage | BETA | Multiple owners | Actively maintained, requires testing | ⚠️ |

#### Receivers (4)

| Component | Stability | Maintainers | Status | Risk |
|-----------|-----------|-------------|--------|------|
| nopreceiver | BETA | OTel Core Team | Production-ready | ✅ |
| otlpreceiver | STABLE | OTel Core Team | Production-ready | ✅ |
| hostmetricsreceiver | BETA | 3 owners | Actively maintained, load-test first | ⚠️ |
| filelogreceiver | BETA | 4 owners | Actively maintained, load-test first | ⚠️ |

#### Exporters (5)

| Component | Stability | Maintainers | Status | Risk |
|-----------|-----------|-------------|--------|------|
| debugexporter | ALPHA | OTel Core Team | **DEV-ONLY** (not for production) | 🔴 |
| nopexporter | BETA | OTel Core Team | Production-ready | ✅ |
| otlpexporter | STABLE | OTel Core Team | Production-ready | ✅ |
| otlphttpexporter | STABLE | OTel Core Team | Production-ready | ✅ |
| fileexporter | ALPHA | paulojmdias | Actively maintained, test first | ⚠️ |

#### Processors (8)

| Component | Stability | Maintainers | Status | Risk |
|-----------|-----------|-------------|--------|------|
| attributesprocessor | BETA | boostchicken | Actively maintained | ⚠️ |
| resourceprocessor | BETA | dmitryax | Actively maintained | ⚠️ |
| spanprocessor | ALPHA | boostchicken | Single owner, high risk | 🔴 |
| probabilisticsamplerprocessor | BETA/ALPHA | jmacd | Actively maintained | ⚠️ |
| filterprocessor | ALPHA | 4 owners | Production-tested | ✅ |
| transformprocessor | BETA | 4 owners | Actively maintained, well-supported | ✅ |
| redactionprocessor | BETA | 4 owners | Actively maintained | ✅ |
| logdedupprocessor | ALPHA | MikeGoldsmith | NEW, actively maintained | ⚠️ |

#### Connectors (1)

| Component | Stability | Maintainers | Status | Risk |
|-----------|-----------|-------------|--------|------|
| forwardconnector | BETA | OTel Core Team | Production-ready | ✅ |

#### Providers (3)

| Component | Stability | Maintainers | Status | Risk |
|-----------|-----------|-------------|--------|------|
| envprovider | STABLE | OTel Core Team | Production-ready | ✅ |
| fileprovider | STABLE | OTel Core Team | Production-ready | ✅ |
| yamlprovider | STABLE | OTel Core Team | Production-ready | ✅ |

---

## Critical Issues & Mitigations

### 🔴 Issue #1: debugexporter Alpha Status

**What:** debugexporter is explicitly for development/testing. Output format is unstable and changes without notice.

**Impact:** Do NOT use in production automation or telemetry parsing.

**Mitigation:**
```yaml
# ❌ WRONG (debug output changes format, breaks parsing)
exporters:
  debug:
    loglevel: info

service:
  pipelines:
    traces:
      exporters: [debug]  # For production monitoring

# ✅ RIGHT (stable format, production-safe)
exporters:
  otlp:
    endpoint: otlp.example.com:4317
  otlphttp:
    endpoint: http://otlp.example.com:4318

service:
  pipelines:
    traces:
      exporters: [otlp]  # Use this
```

**Action:** If using debugexporter in production, migrate to `otlpexporter` or `otlphttpexporter` before upgrading to v26.08.0.

---

### 🔴 Issue #2: spanprocessor Single-Maintainer Risk

**What:** spanprocessor is ALPHA with one maintainer (boostchicken). API subject to breaking changes on minor version upgrades.

**Impact:** Breaking changes possible on every OTel update without notice. Single point of failure for maintenance.

**Mitigation Option A: Use transformprocessor instead**
```yaml
# ❌ RISKY (ALPHA, single maintainer)
processors:
  span:
    name:
      action: insert
      key: custom_span_name
      value: "my-app"

# ✅ SAFER (BETA, 4 maintainers, stable API)
processors:
  transform:
    log_statements:
      - context: span
        statements:
          - set(attributes["custom_span_name"], "my-app")
```

**Mitigation Option B: Test every quarter**
- When upgrading OTel versions, run full regression tests for spanprocessor
- Monitor GitHub issues for breaking changes
- Have transformprocessor alternative ready to migrate to

**Recommendation:** Use `transformprocessor` for new deployments. Existing spanprocessor users: test quarterly, plan migration to transform.

---

### ⚠️ Issue #3: Authentication Extensions Bus Factor

**What:** Four authentication extensions have **single maintainers**:
- basicauthextension, bearertokenauthextension → @frzifus
- oauth2clientauthextension → @pavankrish123
- oidcauthextension → @asweet-confluent

**Impact:** If a maintainer burns out, disappears, or leaves the project, auth stops being maintained. OTel community is actively seeking co-owners.

**Mitigation:**
1. **Primary:** Implement authentication at reverse proxy layer (nginx/Envoy) as backup defense
   ```yaml
   # Reverse proxy handles auth; collector uses simpler/no auth
   # Reduces dependency on OTel auth extension stability
   ```

2. **Secondary:** Monitor maintainer activity weekly
   - Check GitHub commit history and issue response times
   - Watch for maintainer burnout signals
   - Be ready to fork/maintain locally if needed

3. **Community:** Contribute as secondary maintainer to one or more auth extensions

---

### ⚠️ Issue #4: filestorage (New in LTS) Requires Testing

**What:** filestorage extension (added in v26.05.0 LTS) provides persistent queue storage. BETA stability, multiple maintainers.

**Impact:** Critical for durable telemetry delivery. Must be tested at your scale.

**Testing checklist:**
- [ ] Queue persistence survives collector restart
- [ ] Disk usage does not exceed budget (set `queue_size` appropriately)
- [ ] No data loss under sustained high load
- [ ] Recovery from corrupted DB file (manually delete, restart)

**Configuration example:**
```yaml
extensions:
  file_storage:
    directory: /var/lib/otel-collector/queue
    timeout: 10s
    compaction:
      directory: /var/lib/otel-collector/compaction
      reindex_interval: 100

exporters:
  otlp:
    endpoint: otlp.example.com:4317
    sending_queue:
      storage: file_storage
      queue_size: 5000
      num_consumers: 10
```

---

### ⚠️ Issue #5: ALPHA Processors Need Quarterly Testing

These components have ALPHA stability and APIs may change:
- **logdedupprocessor** (NEW)
- **filterprocessor** (production-tested but ALPHA)
- **probabilisticsamplerprocessor** (BETA/ALPHA boundary)
- **fileexporter** (ALPHA)

**Action:** On each OTel version bump, run regression tests for these components. Monitor GitHub issues for breaking changes.

---

## Integration Guide

### Logdedup Processor Configuration

#### Basic Example: Deduplicate identical logs

```yaml
receivers:
  filelog:
    include_paths: [/var/log/app/*.log]

processors:
  logdedup:
    interval: 5s
    log_count_attribute: dedup_count

exporters:
  otlp:
    endpoint: otlp.example.com:4317
    sending_queue:
      storage: file_storage
      queue_size: 1000

service:
  pipelines:
    logs:
      receivers: [filelog]
      processors: [logdedup]
      exporters: [otlp]
```

#### Advanced: Conditional dedup (only ERROR logs)

```yaml
processors:
  logdedup:
    interval: 5s
    log_count_attribute: dedup_count
    condition:
      match_type: regexp
      regexp: '.*ERROR.*'  # Only dedupe ERROR-level logs
```

---

## Migration from v26.05.1 → v26.08.0

### No Breaking Changes

All existing pipelines remain compatible. The upgrade is backwards-compatible.

### Optional: Enable log deduplication

If you want to reduce storage costs for identical repeated logs:
1. Add `logdedup` processor to your log pipeline (see examples above)
2. Test with your actual logs first
3. Monitor dedup_count before scaling

### Required: Audit debugexporter usage

If you have any `exporters: [debug]` in production configs, migrate to `otlp` or `otlphttp` before upgrading.

### Recommended: Audit spanprocessor usage

If you use `spanprocessor`, consider migration to `transformprocessor` for lower maintenance risk. If keeping spanprocessor, add quarterly regression testing to your OTel upgrade checklist.

### Monitoring: Track auth extension maintainers

Add weekly check of GitHub issues and commit history for:
- basicauthextension
- bearertokenauthextension
- oauth2clientauthextension
- oidcauthextension

---

## Deployment Checklist

**Before upgrading to v26.08.0:**

- [ ] Remove debugexporter from all production configurations
- [ ] Audit spanprocessor usage; plan quarterly testing or migration to transformprocessor
- [ ] Load-test filestorage with your queue sizing
- [ ] Load-test hostmetricsreceiver and filelogreceiver at your scale
- [ ] Set up weekly monitoring for auth extension maintainer activity
- [ ] Review authentication at reverse proxy layer (backup for auth extension failure)
- [ ] Document which ALPHA components your deployment uses
- [ ] Schedule quarterly regression testing for ALPHA processors

**During upgrade:**

- [ ] Run full test suite (`make test`)
- [ ] Validate telemetry pipeline end-to-end
- [ ] Confirm debugexporter removal doesn't break monitoring
- [ ] Monitor collector startup logs for errors

**Post-upgrade:**

- [ ] Monitor component metrics and error rates
- [ ] Run logdedup smoke tests if using the new processor
- [ ] Confirm auth extensions are working (bearer token, OIDC, OAuth)

---

## Component Maintenance Status (as of Aug 10, 2026)

✅ **All 27 components have commits in the last 7 days** — actively maintained.

- **Core components** (10): Guaranteed maintained by OTel core team
- **Well-maintained contrib** (12): BETA/STABLE, multi-owner, low risk
- **Active but single-owner** (3): Alpha/BETA, maintained but bus factor risk
- **Requires attention** (2): debugexporter (dev-only), spanprocessor (single owner, ALPHA)

---

## Support & Issues

**For Tulip v26.08.0 issues:**
- Open issues in https://github.com/ollygarden/tulip/issues
- Reference this document and component health status

**For upstream component issues:**
- Check https://github.com/open-telemetry/opentelemetry-collector-contrib/issues
- Search by component name (e.g., "filterprocessor", "bearertokenauthextension")

**For OTel community engagement:**
- Become co-owner/maintainer of high-bus-factor components (auth extensions)
- Contribute fixes or documentation improvements
- Report security issues privately to OTel security team

---

## References

- **v26.08.0 Release Notes:** `docs/tulip-v26.08-release-notes.md`
- **LTS Precedent:** `docs/tulip-lts-may2026-upgrade-plan.md`
- **E-2660:** Evaluation of logdedup processor (integration testing, gotchas, maintainer analysis)
- **OTel Contrib Manifest:** https://github.com/open-telemetry/opentelemetry-collector-releases/tree/main/distributions/otelcol-contrib
