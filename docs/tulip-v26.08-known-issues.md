# Tulip v26.08.0 — Known Issues & Performance Considerations

**Version:** v0.158.0 (OTel Collector)  
**Release Date:** August 10, 2026  
**Based on upstream audit:** August 10, 2026

---

## Overview

Tulip v26.08.0 uses OpenTelemetry Collector v0.158.0 (latest upstream release, 7 versions ahead of v0.151.0). This document lists known issues that may affect production deployments and recommended mitigations.

**Summary:**
- **Critical Issues:** 4 (filelogreceiver CPU, data loss; filterprocessor panic)
- **High Issues:** 7 (memory leaks, performance degradation)
- **Security Advisories:** 2 (prometheus receiver, azure auth extension)
- **Breaking Changes:** 1 (drainprocessor config migration required)

---

## 🚨 CRITICAL Issues (Production Impact)

### Issue #1: Filelog Receiver CPU Consumption (#27404)

**Severity:** CRITICAL (Performance)  
**Component:** filelogreceiver  
**Status:** OPEN (10+ months unresolved)  
**Impact:** Linear CPU growth with file count; ~110% CPU for 1,000 monitored files

**Symptoms:**
- High CPU usage when monitoring hundreds of log files
- CPU increases linearly with file count
- Affects deployments watching many files (e.g., container logs, multi-tenant systems)

**Mitigation:**
1. **Load-test before production:** Monitor CPU on your expected file count
2. **Shard collectors:** If monitoring 1000+ files, run multiple collectors with filelogreceiver sharded per-pod
3. **Monitor metrics:** Track `receiver_filelogreceiver_*` metrics for CPU correlation
4. **Alternative:** Use host-level log collection (syslog, rsyslog) + centralized parsing instead of per-file receiver

**Workaround:** Currently no fix available in v0.158.0. Expected in v0.159.0 or later.

---

### Issue #2: Filelog Receiver Data Loss (#35137)

**Severity:** CRITICAL (Data Loss)  
**Component:** filelogreceiver  
**Status:** OPEN  
**Impact:** Losing ~70% of log lines in high-throughput scenarios

**Symptoms:**
- Significant log loss (70%+) when logs are written rapidly
- Logs appear in files but not in telemetry output
- Affects high-volume log sources

**Mitigation:**
1. **Test throughput limits:** Validate filelogreceiver at your log volume before production
2. **Add redundancy:** Dual-write logs to multiple collectors
3. **Monitoring:** Add metrics alerting for log ingestion gaps
4. **Conservative tuning:** Lower batch sizes, increase retry intervals

**Workaround:** Wait for upstream fix or use alternative log collection method.

---

### Issue #3: Filter Processor Panic (#44705)

**Severity:** CRITICAL (Crash)  
**Component:** filterprocessor  
**Status:** OPEN  
**Impact:** Collector crash/panic when writing debug logs after dropping datapoints

**Symptoms:**
- Collector panic when filterprocessor drops data and debug logging is enabled
- Random crashes during filter operations
- Affects deployments using filterprocessor + debug logging

**Mitigation:**
1. **Disable debug logging:** Set collector log level to `info` (not `debug`)
2. **Test filtering logic:** Validate your filter configs don't cause panics on sample data
3. **Monitoring:** Alert on collector restarts; correlate with filter config changes
4. **Alternative:** Use transformprocessor instead of filterprocessor if possible

**Workaround:** Avoid filterprocessor + debug logging combination. Upgrade to v0.159.0+ when available.

---

### Issue #4: Filelog Receiver Memory Failure (#29330)

**Severity:** CRITICAL (Memory)  
**Component:** filelogreceiver  
**Status:** OPEN  
**Impact:** High memory usage, no recovery after restart

**Symptoms:**
- Memory usage grows unchecked
- OOM kills/crashes
- Collector doesn't recover even after restart

**Mitigation:**
1. **Memory limits:** Set strict memory limits (`GOMAXPROCS`, container limits)
2. **Monitor memory:** Track memory growth; alert on threshold
3. **Rotate logs aggressively:** Keep log files small to reduce receiver memory
4. **Graceful reload:** Restart collectors periodically (rolling restart)

**Workaround:** Implement operational mitigations above; upstream fix expected in future.

---

## ⚠️ HIGH-Priority Issues (Performance/Stability)

### Issue #5: Prometheus Receiver Memory Leak (#31591)

**Severity:** HIGH (Memory leak)  
**Component:** prometheusreceiver  
**Status:** OPEN  
**Impact:** Memory leaks accumulate over days/weeks

**Symptoms:**
- Memory usage grows steadily over time
- Collector becomes slower as memory fills
- Eventually triggers OOM

**Mitigation:**
1. **Monitor memory trends:** Alert on steady memory growth
2. **Implement periodic restarts:** Restart collectors every 7-14 days (rolling)
3. **Metrics cardinality:** Limit prometheus targets and label cardinality
4. **Memory profiling:** Use pprof extension to identify leak source

**Workaround:** Scheduled restarts + monitoring. Upstream fix in progress.

---

### Issue #6: Tail Sampling Processor Memory Leak (#32551)

**Severity:** HIGH (Memory leak)  
**Component:** tailsamplingprocessor  
**Status:** OPEN  
**Impact:** Memory accumulates over days/weeks

**Symptoms:**
- Steady memory growth
- Affects deployments using trace tail sampling
- Similar to prometheus receiver leak pattern

**Mitigation:**
1. **Monitor memory:** Alert on growing memory trend
2. **Periodic restarts:** Every 7-14 days to clear leak
3. **Tune sampling:** Reduce sample rate or decision latency to minimize state
4. **Tracing:** Enable memory profiling to diagnose

**Workaround:** Scheduled restarts. Upstream fix expected.

---

### Issue #7: OTLP Receiver Buffering (#13050)

**Severity:** HIGH (Latency + Memory)  
**Component:** otlpreceiver  
**Status:** OPEN  
**Impact:** Buffers requests 25+ seconds; memory accumulation

**Symptoms:**
- High latency: requests wait 25+ seconds for processing
- Memory usage under spike loads
- Affects real-time monitoring SLAs

**Mitigation:**
1. **Tune buffer sizes:** Reduce `max_recv_msg_size_mib` and related settings
2. **Add horizontal scaling:** Deploy more receivers to distribute load
3. **Rate limit clients:** Implement client-side rate limiting or backpressure
4. **Monitor latency:** Track receiver processing time

**Workaround:** Configuration tuning + scaling. Upstream fix pending.

---

### Issue #8: Host Metrics Receiver High CPU (#8789, #33340)

**Severity:** HIGH (Performance)  
**Component:** hostmetricsreceiver  
**Status:** OPEN  
**Impact:** High CPU consumption in process metrics collection

**Symptoms:**
- Collector CPU spikes when collecting process metrics
- Worse on systems with many processes
- Windows multi-processor-group systems especially affected

**Mitigation:**
1. **Scrape interval tuning:** Increase collection interval (e.g., 60s instead of 30s)
2. **Selective metrics:** Collect only needed metrics (disable process metrics if not used)
3. **Shard collectors:** Distribute metrics collection across multiple collectors
4. **Platform-specific:** Windows users should test multi-processor-group compatibility

**Workaround:** Configuration tuning. Upstream fix in progress.

---

### Issue #9: Filter Processor Logic Issues (#30176, #22152)

**Severity:** HIGH (Data Integrity)  
**Component:** filterprocessor  
**Status:** OPEN  
**Impact:** Incorrect filtering; non-matching data passes through; K8s attributes not filtered

**Symptoms:**
- Filtered data appears in output (should be dropped)
- K8s-generated attributes bypass filtering
- Filter conditions evaluated incorrectly

**Mitigation:**
1. **Test thoroughly:** Validate filtering logic against sample data
2. **Alternative:** Use transformprocessor or custom OTTL logic instead
3. **Monitoring:** Validate output data matches filter criteria
4. **Simple filters:** Use simplest possible filter conditions

**Workaround:** Avoid complex filter scenarios. Use transformprocessor for complex logic.

---

## 🔴 SECURITY Issues (v0.158.0 Affected)

### Security Advisory #1: Prometheus Receiver (GHSA-8hxm-mxr2-qf9h)

**Severity:** HIGH  
**Affected:** prometheusreceiver v0.158.0  
**Type:** Authentication bypass / exposure

**Impact:**
- Sensitive prometheus metrics may be exposed
- Authentication bypassed under certain conditions

**Mitigation:**
1. **Limit access:** Firewall prometheus receiver port (default 8888)
2. **No auth credentials in metrics:** Never include secrets in prometheus targets
3. **Monitor access logs:** Alert on unexpected receiver port access
4. **Upgrade plan:** Plan migration to v0.159.0+ when available

**Workaround:** Network isolation. Avoid prometheus receiver in untrusted networks.

---

### Security Advisory #2: Azure Auth Extension (GHSA-pjv4-3c63-699f)

**Severity:** HIGH  
**Affected:** azureauthextension v0.158.0  
**Type:** Token refresh bypass

**Impact:**
- Azure token validation bypassed under certain conditions
- Potential unauthorized access to Azure resources

**Mitigation:**
1. **Don't use Azure auth for critical paths:** If possible, use simpler auth (bearer token, OAuth)
2. **Network isolation:** Restrict Azure auth extension to trusted networks only
3. **Audit logs:** Monitor Azure resource access for anomalies
4. **Upgrade plan:** Migrate to v0.159.0+ when available

**Workaround:** Use alternative authentication methods (bearer token, OIDC).

---

## 🟡 BREAKING CHANGES

### Drain Processor Configuration Migration

**Affected:** drainprocessor  
**Change Type:** Configuration breaking change

**Old config (v0.151.0):**
```yaml
processors:
  drain:
    extract_parameters: true
    params_attribute: parameters
```

**New config (v0.158.0):**
```yaml
processors:
  drain:
    drain:
      algorithm: drain
      tau: 0.5
    masking_rules:
      - pattern: '\d+'           # Mask numbers
        replacement: '<number>'
    emit_wildcards: true         # Emit learned templates
```

**Migration:**
1. **Update configs** before upgrade
2. **Test pipeline** with new drain config
3. **Validate templates** are generated correctly
4. **Update documentation** for your ops team

**Reference:** See `docs/tulip-v26.08-release-notes.md` for full drain configuration examples.

---

## Deployment Safety Checklist

Before deploying Tulip v26.08.0 to production:

- [ ] **filelogreceiver load test:** Validate CPU on your file count (if using filelogreceiver)
- [ ] **Filterprocessor validation:** Test filter logic on representative data (if using filterprocessor)
- [ ] **Security review:** Confirm no prometheus or azure auth exposed to untrusted networks
- [ ] **Drain config migration:** Update drain processor configs to v0.158.0 format
- [ ] **Memory monitoring:** Set up alerts for memory growth (prometheus rx, tail sampling)
- [ ] **Crash recovery:** Configure automatic restart on panic
- [ ] **Rollback plan:** Keep v0.151.0 container images available for quick rollback
- [ ] **Test schedule:** Plan weekly production monitoring of logs, metrics, and errors
- [ ] **Capacity buffer:** Allocate 20-30% extra CPU/memory for safety margin
- [ ] **Upgrade strategy:** Rolling restart of collectors during low-traffic window

---

## Monitoring & Alerting

### Metrics to Monitor

```yaml
# CPU and memory per component
process_runtime_go_goroutines           # Goroutine count (memory proxy)
process_runtime_go_mem_heap_alloc_bytes # Heap allocation
process_cpu_time_seconds_total          # CPU usage

# Receiver health
receiver_filelogreceiver_files_open     # Open files
receiver_filelogreceiver_files_read     # Files successfully read
receiver_accepted_log_records           # Ingested logs
receiver_refused_log_records            # Dropped logs

# Processor health
processor_dropped_spans                 # Dropped by filters
processor_batch_size_triggered_sends    # Batching activity
processor_tail_sampling_trace_removed   # Trace drops
```

### Alert Rules (Prometheus)

```yaml
# File receiver CPU spike
- alert: FilereceiverHighCPU
  expr: rate(process_cpu_time_seconds_total[5m]) > 0.8
  for: 10m
  labels:
    severity: warning

# Memory leak detector
- alert: MemoryTrendingUp
  expr: rate(process_runtime_go_mem_heap_alloc_bytes[1h]) > 0
  for: 1h
  labels:
    severity: warning

# Filter processor drop rate
- alert: HighFilterDropRate
  expr: rate(processor_dropped_spans[5m]) > 100
  for: 5m
  labels:
    severity: warning
```

---

## Support & Escalation

**For Tulip-specific issues:**
- File issue: https://github.com/ollygarden/tulip/issues
- Reference this document and specific issue number

**For upstream OTel component issues:**
- Search: https://github.com/open-telemetry/opentelemetry-collector-contrib/issues
- Report missing: https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/new

**For security advisories:**
- Report privately: https://github.com/open-telemetry/opentelemetry-collector/security/advisories

---

## Timeline to Resolution

Based on upstream release cadence (biweekly):

| Issue | Current | Expected Fix |
|-------|---------|--------------|
| filelogreceiver CPU (#27404) | OPEN 10+ months | v0.159.0+ (Aug 18+) or later |
| filelogreceiver data loss (#35137) | OPEN | v0.159.0+ |
| filterprocessor panic (#44705) | OPEN | v0.159.0+ |
| prometheus memory leak (#31591) | OPEN | TBD |
| tail sampling memory leak (#32551) | OPEN | TBD |
| otlpreceiver buffering (#13050) | OPEN | TBD |
| Drain config breaking change | FIXED (v0.158.0) | N/A |

**Recommendation:** Monitor upstream releases. Plan upgrade to v0.159.0 when available (~Aug 18).

---

## Questions?

This document is part of Tulip v26.08.0 release documentation. For questions or concerns:
1. Review this document thoroughly
2. Check upstream issue links for latest status
3. Test in staging environment before production
4. File issue with specific error messages and reproduction steps
