# OllyGarden Tracker - Tulip Components Open Issues & Stability Status

> Generated: 2026-04-30
> Based on: `distributions/tulip/manifest.yaml` v0.150.0

## Purpose

This document serves as OllyGarden's **baseline tracker** for known upstream issues
across all components included in the Tulip distribution.

**Goals:**
- Provide a clear snapshot of the current state of each component at the time of a release
- Help OllyGarden organize and proactively assist customers with bugs we know exist in each component
- At the start of each new release, compare this baseline against what has been resolved upstream and what remains open
- Track stability status changes over time to inform LTS support commitments

**Usage:** Before each new Tulip release, generate a fresh version of this tracker and diff
it against the previous one. Resolved issues should be highlighted in release notes.
New issues should be evaluated for customer impact and potential workarounds.

---

## Stability Summary

| Component | Type | Stability | Repo |
|---|---|---|---|
| zpagesextension | Extension | Beta | core |
| healthcheckextension | Extension | Alpha | contrib |
| pprofextension | Extension | Beta | contrib |
| basicauthextension | Extension | Beta | contrib |
| bearertokenauthextension | Extension | Beta | contrib |
| oauth2clientauthextension | Extension | Beta | contrib |
| oidcauthextension | Extension | Beta | contrib |
| nopreceiver | Receiver | Beta (traces/metrics/logs), Alpha (profiles) | core |
| otlpreceiver | Receiver | Stable (traces/metrics/logs), Alpha (profiles) | core |
| debugexporter | Exporter | Alpha | core |
| nopexporter | Exporter | Beta (traces/metrics/logs), Alpha (profiles) | core |
| otlpexporter | Exporter | Stable (traces/metrics/logs), Alpha (profiles) | core |
| otlphttpexporter | Exporter | Stable (traces/metrics/logs), Alpha (profiles) | core |
| fileexporter | Exporter | Alpha (traces/metrics/logs), Development (profiles) | contrib |
| memorylimiterprocessor | Processor | Beta (traces/metrics/logs), Alpha (profiles) | core |
| attributesprocessor | Processor | Beta (traces/metrics/logs) | contrib |
| resourceprocessor | Processor | Beta (traces/metrics/logs), Development (profiles) | contrib |
| spanprocessor | Processor | Alpha (traces) | contrib |
| probabilisticsamplerprocessor | Processor | Beta (traces), Alpha (logs) | contrib |
| filterprocessor | Processor | Alpha (traces/metrics/logs), Development (profiles) | contrib |
| transformprocessor | Processor | Beta (traces/metrics/logs), Development (profiles) | contrib |
| forwardconnector | Connector | Beta (traces/metrics/logs), Alpha (profiles) | core |

---

## Extensions

### zpagesextension
**Stability:** Beta
**Repo:** `open-telemetry/opentelemetry-collector`

| Issue | Link |
|---|---|
| Include description, fromVersion, toVersion, and reference URL in featuregate String() | [#7625](https://github.com/open-telemetry/opentelemetry-collector/issues/7625) |
| Reduce transitive dependencies of components regarding internal telemetry providers | [#13842](https://github.com/open-telemetry/opentelemetry-collector/issues/13842) |

**Summary:** Mature component with no significant open bugs. Issue #7625 is about general feature gate improvements. #13842 specifically mentions zpagesextension as a target for dependency refactoring. Note: zpages does not work when `service::telemetry::traces::level` is set to `none`, as it creates a No-Op TracerProvider - this is a known architectural limitation.

---

### healthcheckextension
**Stability:** Alpha
**Repo:** `open-telemetry/opentelemetry-collector-contrib`

| Issue | Link |
|---|---|
| Deprecate healthcheckextension in favor of healthcheckv2extension | [#42256](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/42256) |
| "check_collector_pipeline" is always healthy, even when exporter is failing | [#11780](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/11780) |

**Summary:** There is consensus on merging v2 functionality into v1 rather than replacing the component entirely. The migration plan proposed by @evan-bradley (code owner) was executed in 4 phases:

1. Move v2 code to an `internal` module
2. Have v1 pull from the internal module, enabling v2 features via config
3. Deprecate the separate v2 module
4. Transition v2 behavior to become the default

**Current status (since v0.142.0):** The v1 extension **already uses the v2 implementation under the hood** (confirmed by @iblancasa in the issue thread). To enable v2 features (component status reporting), enable the feature gate `extension.healthcheck.useComponentStatus`.

**Remaining steps upstream:**
1. Mark `healthcheckv2extension` module as deprecated
2. Move `extension.healthcheck.useComponentStatus` feature gate to beta (enabled by default)

**Impact for Tulip LTS:** No component swap needed - the `healthcheckextension` in our manifest is the correct path forward. Consider enabling `useComponentStatus` by default in the Tulip LTS config for more accurate pipeline health reporting (solves the "always healthy" bug in #11780).

---

### pprofextension
**Stability:** Beta
**Repo:** `open-telemetry/opentelemetry-collector-contrib`

| Issue | Link |
|---|---|
| Memory leak on system with 128 x86_64 cores | [#36574](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/36574) |

**Summary:** There is a memory leak report on systems with many cores (128 x86_64 cores). The leak appears to be related to Go's runtime profiling, not the pprof extension itself. Otherwise stable with no critical issues.

---

### basicauthextension
**Stability:** Beta
**Repo:** `open-telemetry/opentelemetry-collector-contrib`

No component-specific issues open. Only general issues like "Add Warning header to all necessary components" ([#19172](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/19172)).

**Summary:** Stable component with no bugs or feature requests directly related to it.

---

### bearertokenauthextension
**Stability:** Beta
**Repo:** `open-telemetry/opentelemetry-collector-contrib`

| Issue | Link |
|---|---|
| Add comments to token file | [#46100](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/46100) |
| Ability to track usage of Bearer tokens | [#45047](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/45047) |

**Summary:** Two active feature requests: support for comments in token files, and the ability to track bearer token usage (auditing). Both are enhancements, not bugs. Component is functional and stable.

---

### oauth2clientauthextension
**Stability:** Beta
**Repo:** `open-telemetry/opentelemetry-collector-contrib`

No component-specific issues open. Only general issues like "Add Warning header" ([#19172](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/19172)).

**Summary:** Stable component with no reported issues.

---

### oidcauthextension
**Stability:** Beta
**Repo:** `open-telemetry/opentelemetry-collector-contrib`

| Issue | Link |
|---|---|
| Report for failed tests on main | [#47652](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/47652) |
| Support JWT signatures of type ES256 | [#47845](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/47845) |
| Allow ignoring issuer check | [#46791](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/46791) |
| Collector doesn't start when any OIDC issuer is unreachable | [#45206](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/45206) |

**Summary:** Component with the most active issues among extensions. **Critical bug:** the collector fails to start if any OIDC issuer is unreachable (#45206) - this can impact availability in production. There are also requests for ES256 support (currently only RS256) and the ability to skip issuer verification (useful in dev environments). Tests are failing on main (#47652). **This component deserves special attention in the LTS context.**

---

## Receivers

### nopreceiver
**Stability:** Beta (traces/metrics/logs), Alpha (profiles)
**Repo:** `open-telemetry/opentelemetry-collector`

No relevant open issues.

**Summary:** Simple utility component (no-op). No known problems.

---

### otlpreceiver
**Stability:** Stable (traces/metrics/logs), Alpha (profiles)
**Repo:** `open-telemetry/opentelemetry-collector`

| Issue | Link |
|---|---|
| Support Rate Limiting | [#6725](https://github.com/open-telemetry/opentelemetry-collector/issues/6725) |
| Option to omit detailed error response | [#14072](https://github.com/open-telemetry/opentelemetry-collector/issues/14072) |
| Unable to monitor ResourceExhausted errors | [#13685](https://github.com/open-telemetry/opentelemetry-collector/issues/13685) |
| Accepts non-utf8 data | [#11449](https://github.com/open-telemetry/opentelemetry-collector/issues/11449) |
| HTTP server span name missing route pattern | [#14508](https://github.com/open-telemetry/opentelemetry-collector/issues/14508) |
| Unable to parse IPv6 addresses with zones | [#14545](https://github.com/open-telemetry/opentelemetry-collector/issues/14545) |

**Summary:** Several active discussions. The most relevant is **rate limiting support** (#6725), a long-standing request (since 2022) to protect the collector against overload - still unresolved. There is also a security request to **omit error details in responses** (#14072) - important for production. The receiver accepts non-UTF8 data (#11449), which can cause downstream issues. **The otlpreceiver is the most critical Tulip component and deserves continuous monitoring.**

---

## Exporters

### debugexporter
**Stability:** Alpha
**Repo:** `open-telemetry/opentelemetry-collector`

| Issue | Link |
|---|---|
| Make output configurable | [#9372](https://github.com/open-telemetry/opentelemetry-collector/issues/9372) |

**Summary:** Feature request to make the output more configurable. This is a development/debug exporter, not meant for production. Alpha status is expected and acceptable for its intended use.

---

### nopexporter
**Stability:** Beta (traces/metrics/logs), Alpha (profiles)
**Repo:** `open-telemetry/opentelemetry-collector`

| Issue | Link |
|---|---|
| Remove one of nopexporter.NewFactory or exportertest.NewNopFactory | [#11369](https://github.com/open-telemetry/opentelemetry-collector/issues/11369) |

**Summary:** Internal issue about factory duplication in the code. No impact on end users.

---

### otlpexporter
**Stability:** Stable (traces/metrics/logs), Alpha (profiles)
**Repo:** `open-telemetry/opentelemetry-collector`

| Issue | Link |
|---|---|
| Duplicate spans generated by retries | [#11056](https://github.com/open-telemetry/opentelemetry-collector/issues/11056) |
| Exporter eventually failing on retryable errors not reported in healthcheck v2 | [#13013](https://github.com/open-telemetry/opentelemetry-collector/issues/13013) |
| Partial Success with 0 rejected records should not be a warning | [#13476](https://github.com/open-telemetry/opentelemetry-collector/issues/13476) |
| Allow adding option to existing components | [#14497](https://github.com/open-telemetry/opentelemetry-collector/issues/14497) |

**Summary:** **Important bug:** retries can generate duplicate spans (#11056) - this may affect data accuracy in production. There is also an issue with retry failures not being properly reported to healthcheck v2 (#13013). The Partial Success issue (#13476) is about reducing log noise. **Monitor the duplication bug for the LTS.**

---

### otlphttpexporter
**Stability:** Stable (traces/metrics/logs), Alpha (profiles)
**Repo:** `open-telemetry/opentelemetry-collector`

| Issue | Link |
|---|---|
| Bad request because it exceeds known default sizes | [#15177](https://github.com/open-telemetry/opentelemetry-collector/issues/15177) |
| Add source_vip / outbound_vip configuration | [#12586](https://github.com/open-telemetry/opentelemetry-collector/issues/12586) |
| HTTP export metric lost resource attributes | [#9961](https://github.com/open-telemetry/opentelemetry-collector/issues/9961) |
| Allow configuring retryable status codes | [#14228](https://github.com/open-telemetry/opentelemetry-collector/issues/14228) |
| Add bandwidth limit after data marshal and compression | [#7414](https://github.com/open-telemetry/opentelemetry-collector/issues/7414) |
| Traces not going out after "Preparing to make HTTP request" | [#10660](https://github.com/open-telemetry/opentelemetry-collector/issues/10660) |
| Add debug logging to the exporter | [#5629](https://github.com/open-telemetry/opentelemetry-collector/issues/5629) |

**Summary:** **Critical bug:** metrics lose resource attributes when exporting via HTTP (#9961) - can cause loss of context in backends. There are also issues with requests exceeding 4 MiB default sizes without splitting (#15177). Bug where traces get stuck at the HTTP send stage (#10660). Relevant feature requests: configuring retry status codes (#14228), bandwidth limiting (#7414), and debug logging (#5629, very old). Note: the `otlphttp` alias is being deprecated - docs now recommend `otlp_http`. **Critical Tulip component - monitor the resource attributes and payload size issues.**

---

### fileexporter
**Stability:** Alpha (traces/metrics/logs), Development (profiles)
**Repo:** `open-telemetry/opentelemetry-collector-contrib`

| Issue | Link |
|---|---|
| nil pointer dereference during collector shutdown | [#46871](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/46871) |
| Promote fileexporter to beta | [#41669](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/41669) |
| Invalid zstd file | [#44077](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/44077) |

**Summary:** **Critical bug:** nil pointer dereference during collector shutdown (#46871) - can cause crashes. There is also a bug with zstd compression generating invalid files (#44077). There is a discussion to promote fileexporter to Beta (#41669), but current bugs likely block that progression. **Alpha component - use with caution in production.**

---

## Processors

### memorylimiterprocessor
**Stability:** Beta (traces/metrics/logs), Alpha (profiles)
**Repo:** `open-telemetry/opentelemetry-collector`

| Issue | Link |
|---|---|
| Implement Component Status reporting | [#14700](https://github.com/open-telemetry/opentelemetry-collector/issues/14700) |
| Degenerate collector performance when exporter has problems | [#4981](https://github.com/open-telemetry/opentelemetry-collector/issues/4981) |
| Stricter memory limiter | [#8694](https://github.com/open-telemetry/opentelemetry-collector/issues/8694) |
| Refactor to use built-in memory limiting in Go | [#5708](https://github.com/open-telemetry/opentelemetry-collector/issues/5708) |
| Allow disabling GC | [#15081](https://github.com/open-telemetry/opentelemetry-collector/issues/15081) |
| Differentiate between refused and dropped | [#12463](https://github.com/open-telemetry/opentelemetry-collector/issues/12463) |
| Ensure reliable data delivery in erroneous situations | [#7460](https://github.com/open-telemetry/opentelemetry-collector/issues/7460) |

**Summary:** Several active architectural discussions. The most important is **performance degradation when exporters have problems** (#4981) - the memory limiter doesn't handle backpressure well. There is a proposal to use Go's native memory limiting (#5708) instead of the current polling-based approach. A "stricter" mode (#8694) for more aggressive limits is also discussed. Users stacking multiple memory limiters want to **disable GC** on some of them (#15081). The **distinction between refused vs dropped data** (#12463) is important for observability - knowing whether there was backpressure or data loss. Docs recommend placing this processor first in every pipeline and setting `GOMEMLIMIT` to 80% of the hard limit. **Critical component for Tulip production stability.**

---

### attributesprocessor
**Stability:** Beta (traces/metrics/logs)
**Repo:** `open-telemetry/opentelemetry-collector-contrib`

| Issue | Link |
|---|---|
| Support default values for resource processor attributes | [#45352](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/45352) |

**Summary:** Feature request to support default values for attributes. Stable and mature component with no open bugs.

---

### resourceprocessor
**Stability:** Beta (traces/metrics/logs), Development (profiles)
**Repo:** `open-telemetry/opentelemetry-collector-contrib`

| Issue | Link |
|---|---|
| Support default values for resource processor attributes | [#45352](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/45352) |

**Summary:** Same feature request as attributesprocessor about default values. Stable component with no known bugs.

---

### spanprocessor
**Stability:** Alpha (traces)
**Repo:** `open-telemetry/opentelemetry-collector-contrib`

No component-specific issues open beyond general project issues.

**Summary:** Alpha component but stable. Operates only on traces. No relevant bugs or active discussions.

---

### probabilisticsamplerprocessor
**Stability:** Beta (traces), Alpha (logs)
**Repo:** `open-telemetry/opentelemetry-collector-contrib`

| Issue | Link |
|---|---|
| Align usage of hashing functions | [#6136](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/6136) |
| Review units of telemetry in metadata.yaml | [#34143](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/34143) |
| Different sampling rates for different applications | [#31562](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/31562) |

**Summary:** Discussion about **aligning hashing functions** (#6136) for consistent sampling across instances. Feature request for **different sampling rates per application** (#31562) - very useful for production but still not implemented. Log support is in Alpha.

---

### filterprocessor
**Stability:** Alpha (traces/metrics/logs), Development (profiles)
**Repo:** `open-telemetry/opentelemetry-collector-contrib`

| Issue | Link |
|---|---|
| Build fails for v0.146.0 and above | [#47289](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/47289) |
| Deprecate and remove old filterprocessor OTTL format | [#41176](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/41176) |
| Add Filter Action Identifier | [#42321](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/42321) |
| Standardize configuration on OTTL | [#18642](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/18642) |
| Add "Span Drop Audit" debugging capability | [#44479](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/44479) |

**Summary:** Component in active transition. There is a **build bug** in versions >= 0.146.0 (#47289). The old configuration format is being deprecated in favor of OTTL (#41176, #18642). Important feature request: **dropped span audit** (#44479) for debugging. **Attention: the OTTL migration is important for the LTS - ensure customer configs use the new format.**

---

### transformprocessor
**Stability:** Beta (traces/metrics/logs), Development (profiles)
**Repo:** `open-telemetry/opentelemetry-collector-contrib`

| Issue | Link |
|---|---|
| Inconsistency in transformprocessor logs | [#48052](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/48052) |
| Add OTTL exemplar context for timestamp manipulation | [#47322](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/47322) |
| Basic support for Resource EntityRefs | [#41092](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/41092) |
| Remove revive var-naming exception | [#45002](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/45002) |
| Emit counter metrics when truncate_all or limit modify attributes | [#47731](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/47731) |
| Drop exemplars / span_id (high cardinality fixes) | [#44478](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/44478) |

**Summary:** Active component with several improvements under discussion. Recent **log inconsistency** bug (#48052). Feature request for **observability metrics** when attributes are modified (#47731) - useful for auditing. Discussion about **dropping exemplars and span_ids** to reduce cardinality (#44478). The overall OTTL roadmap (#18643) directly impacts this component.

---

## Connectors

### forwardconnector
**Stability:** Beta (traces/metrics/logs), Alpha (profiles)
**Repo:** `open-telemetry/opentelemetry-collector`

No component-specific issues open.

**Summary:** Simple routing (forward) component. No bugs or active discussions. Works as expected.

---

## Providers (confmap)

### envprovider, fileprovider, yamlprovider
**Stability:** Part of the collector core `confmap` package
**Repo:** `open-telemetry/opentelemetry-collector`

These providers are part of the collector core and are generally at **Beta/Stable** status. No relevant component-specific issues.

---

## Highlights for Tulip LTS May 2026

### Critical Issues to Monitor
1. **healthcheckextension** - v1 already uses v2 implementation since v0.142.0; enable `extension.healthcheck.useComponentStatus` for accurate pipeline health ([#42256](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/42256))
2. **oidcauthextension** - collector fails to start if OIDC issuer is unreachable ([#45206](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/45206))
3. **otlpexporter** - duplicate spans generated by retries ([#11056](https://github.com/open-telemetry/opentelemetry-collector/issues/11056))
4. **otlpexporter** - retry failures not reported in healthcheck v2 ([#13013](https://github.com/open-telemetry/opentelemetry-collector/issues/13013))
5. **otlphttpexporter** - loss of resource attributes ([#9961](https://github.com/open-telemetry/opentelemetry-collector/issues/9961))
6. **otlphttpexporter** - payloads exceed 4 MiB without splitting ([#15177](https://github.com/open-telemetry/opentelemetry-collector/issues/15177))
7. **fileexporter** - nil pointer dereference on shutdown ([#46871](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/46871))
8. **filterprocessor** - build fails >= v0.146.0 ([#47289](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/47289))
9. **memorylimiterprocessor** - performance degrades with exporter problems ([#4981](https://github.com/open-telemetry/opentelemetry-collector/issues/4981))
10. **otlpreceiver** - no rate limiting support (open since 2022) ([#6725](https://github.com/open-telemetry/opentelemetry-collector/issues/6725))

### Alpha Components (higher risk)
- healthcheckextension
- debugexporter (intentionally Alpha - output formats have no stability guarantee)
- spanprocessor
- filterprocessor
- fileexporter

### Known Limitations to Document in LTS
- **otlpreceiver** accepts invalid non-UTF8 data ([#11449](https://github.com/open-telemetry/opentelemetry-collector/issues/11449))
- **otlpreceiver** has no native rate limiting support ([#6725](https://github.com/open-telemetry/opentelemetry-collector/issues/6725))
- **zpagesextension** does not work with `service::telemetry::traces::level: none`
- **memorylimiterprocessor** does not differentiate "refused" from "dropped" data in metrics ([#12463](https://github.com/open-telemetry/opentelemetry-collector/issues/12463))
- **filterprocessor** transitioning from old format to OTTL - old configs will be removed ([#41176](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/41176))

### healthcheckextension v1/v2 Consolidation (Detail)

As confirmed in [#42256](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/42256), the v1 and v2 healthcheck extensions have been consolidated:

**Migration plan (proposed by @evan-bradley, code owner):**
1. Move v2 code to an `internal` module
2. Have v1 pull from the internal module, enabling v2 features via config
3. Deprecate the separate `healthcheckv2extension` module
4. Transition v2 behavior to become the default

**Current state (since v0.142.0):**
- The `healthcheckextension` (v1) **already uses the v2 implementation internally**
- To enable v2 features (component status reporting), set the feature gate: `extension.healthcheck.useComponentStatus`
- This solves the long-standing bug where pipeline checks always report "healthy" even when exporters are failing (#11780)

**Pending upstream work:**
1. Mark the `healthcheckv2extension` module as deprecated
2. Move `extension.healthcheck.useComponentStatus` feature gate to beta (enabled by default)

**Tulip LTS recommendation:** No component swap needed. The `healthcheckextension` in our manifest is the correct path. Consider enabling `useComponentStatus` by default in Tulip's config for more accurate health reporting.
