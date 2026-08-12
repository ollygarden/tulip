---
name: release-version
description: Orchestrate a complete Tulip version release (component audit → version bump → documentation → image sync → web update)
---

# Tulip Release Skill

Orchestrate a complete Tulip quarterly release from start to finish, covering component evaluation, version updates across all distributions, documentation generation, and downstream synchronization.

## Prerequisites

**Before invoking this skill:**

1. **Decide target version**
   - [ ] Check latest: https://github.com/open-telemetry/opentelemetry-collector-releases/releases
   - [ ] Decide: release with latest or choose specific version

2. **Identify downstream repos**
   - [ ] Search for repos that reference Tulip/bonsai distributions:
     ```bash
     grep -r "ghcr.io/ollygarden/bonsai\|ollygarden/bonsai" /path/to/your/repos
     ```
   - [ ] Document which repos need updates in Phase 7

3. **Administrative**
   - [ ] Create Linear card/issue for this release
   - [ ] Ensure access to all identified repos
   - [ ] Plan time for component audit (varies by scope)

---

## Release Workflow

### Phase 1: Component Audit

#### 1.1 Check Latest OTel Version

```bash
# Find latest stable release
curl -s https://api.github.com/repos/open-telemetry/opentelemetry-collector-releases/releases/latest | jq '.tag_name'

# Alternatively, check: https://github.com/open-telemetry/opentelemetry-collector-releases/releases
```

**Decision:** Choose version to release with (recommend latest for regular quarterly releases).

#### 1.2 Audit All Components

Run comprehensive upstream audit using an agent:

```
Agent Task:
- Audit all components in the manifest at target version
- List ALL open GitHub issues (not just closed)
- Categorize by severity (CRITICAL, HIGH, MEDIUM, LOW)
- Flag performance-affecting issues
- Identify security vulnerabilities
- Check breaking changes vs prior version
- Provide recommendation table (Use/Caution/Investigate)
```

**Deliverables:**
- Component health table (stability, maintainers, open issues, risk level)
- CRITICAL/HIGH issues list with details
- Breaking changes documentation
- Security advisories affecting this version

#### 1.3 Evaluate New/Alpha Components (if adding)

For each new component being added to the release:

```
Agent Task:
- Fetch upstream issue history
- Check maintainer activity (recent commits, PR reviews)
- Run hands-on validation (build, test, smoke test with sample data)
- Document integration requirements
- List any gotchas or configuration quirks
- Assess production readiness
```

**Deliverables:**
- Integration guide with examples
- Test results (build, test suite)
- Known limitations or gotchas

---

### Phase 2: Version Updates

#### 2.1 Update Tulip Manifest

```bash
cd tulip
# Edit distributions/tulip/manifest.yaml:
# - Update dist.version: NEW_VERSION
# - Replace all v0.X.Y → vTARGET.X.Y (component versions)
# - Update provider versions (envprovider, fileprovider, yamlprovider)
```

**Validation:**
```bash
make generate-sources    # Regenerate Go scaffolding
go mod tidy              # Update dependencies
make build               # Compile binary
make test                # Run full test suite
make ensure-goreleaser-up-to-date  # Validate release config
```

#### 2.2 Update Downstream Distributions (if applicable)

```bash
cd downstream-repo
# Update each distribution manifest in the repository:
# - Update dist.version: NEW_VERSION
# - Replace all component versions to match Tulip's manifest
# - Ensure version synchronization across distributions
```

**Validation:**
```bash
# Per distribution:
make build              # Must build successfully
docker build -t dist:local .
```

#### 2.3 Update Deployment Configuration (if applicable)

Update deployment configs that reference distribution versions:

```yaml
# For each deployment config:
- image: "ghcr.io/ollygarden/distributions/name:NEW_VERSION"
```

**Validation:**
```bash
# Verify configuration syntax
helm lint .
helm template . > /tmp/manifests.yaml  # Verify valid structure
```

---

### Phase 3: Documentation

#### 3.1 Unified Release Documentation

Create a single comprehensive release document that covers all aspects:

**File:** `docs/tulip-vYY.MM-release.md`

**Contents:**
- **Release Summary**
  - Version number and OTel component version
  - Release date
  - High-level summary of changes

- **Component Changes**
  - What's new (added/removed/changed components)
  - Full component manifest table (by category: extensions, receivers, exporters, processors, connectors, providers)
  - Configuration examples for new components

- **Component Health & Issues**
  - Component health summary table (stability, maintainers, open issues, risk level)
  - CRITICAL/HIGH issues with details and mitigations
  - Security advisories (if any)
  - Performance considerations
  - Breaking changes and migration paths

- **Integration Guide**
  - Configuration examples for key components
  - Integration notes and gotchas
  - Upstream compatibility information

- **Deployment Checklist**
  - Pre-deployment validation steps
  - Testing recommendations
  - Monitoring and alerting setup
  - Rollback plan

- **Monitoring & Alerting**
  - Metrics to monitor
  - Alert rules for critical conditions
  - Timeline to upstream fixes (if applicable)

**Process:**
- Reference prior release docs for structure and examples
- Include component audit findings from Phase 1
- Document all breaking changes
- Provide concrete, tested configuration examples

#### 3.2 Update Component Tracker (if applicable)

**File:** `docs/component-tracker.md` (or equivalent)

**Process:**
- Regenerate tracker for new manifest version
- Diff against prior tracker
- Call out newly resolved issues in release documentation

**Reference:** `README.md` and `CLAUDE.md` for process details

---

### Phase 4: Git & PR Workflow

#### 4.1 Create Release Branches & Commit

```bash
# Tulip repository
cd tulip
git checkout -b release-vYY.MM.0
git add distributions/tulip/manifest.yaml go.sum docs/tulip-vYY.MM-release.md
git commit -m "feat(tulip): release vYY.MM.0 with v0.X.Y components

- Updated manifest to v0.X.Y components
- Component audit: all components reviewed, X CRITICAL/HIGH issues documented
- Comprehensive release documentation with health audit, mitigations, deployment checklist
- Validation: build ✓, test ✓, goreleaser ✓

Refs: E-XXXX"

# Downstream repositories (if applicable)
# Identify which repos contain Tulip image references before proceeding
# Common patterns:
#   - Kubernetes Helm values files
#   - Infrastructure-as-code configs (Kustomize, Terraform, etc.)
#   - Deployment patches or overlays

cd downstream-repo
git checkout -b sync-release-vYY.MM.0
git add [files-with-version-references]
git commit -m "chore: sync component versions for Tulip vYY.MM.0

- Updated all distribution image references to v0.X.Y
- Updated deployment configurations to reference new versions
- Aligns with upstream Tulip vYY.MM.0 release

See tulip#PR_NUMBER for component health audit and known issues.
Refs: E-XXXX"
```

#### 4.2 Create Pull Requests

```bash
# Push branches
cd tulip && git push -u origin release-vYY.MM.0
# Push any downstream repos

# Create PRs with comprehensive descriptions
cd tulip && gh pr create --title "feat(tulip): release vYY.MM.0 with v0.X.Y" --body "
## Summary
[Release overview with component audit findings, key issues, deployment checklist]

## Changes
- Manifest updated to v0.X.Y
- Release documentation generated
- Component health audit completed

## Testing
- Build: ✓
- Tests: ✓
- Goreleaser: ✓

See docs/tulip-vYY.MM-release.md for full audit results, known issues, and mitigations.
"
```

#### 4.3 Update Linear Release Card

**Card:** Release card for this version

**Add to description/comments:**
```markdown
## Status

**Tulip PR:** {github_url}
- Manifest: v0.X.Y components
- Release documentation: docs/tulip-vYY.MM-release.md
- Component audit: [summary of key findings]

**Downstream PRs:** [links if applicable]

**Critical findings documented:** [reference to doc]
```

---

### Phase 5: Code Review & Validation

#### 5.1 Review Component Audit Results

**Checklist:**
- [ ] All components in manifest listed
- [ ] CRITICAL/HIGH issues documented with specifics
- [ ] Security advisories flagged and assessed
- [ ] Breaking changes identified with migration paths
- [ ] Component health assessment reasonable (stability, maintainers, issues)

#### 5.2 Review Documentation

**Checklist for release documentation:**
- [ ] Component manifest accurate and complete
- [ ] Health audit summary clear and actionable
- [ ] Known issues include specific symptoms and mitigations
- [ ] Breaking changes documented with migration examples
- [ ] Configuration examples provided and (ideally) tested
- [ ] Deployment checklist covers critical pre-deployment steps
- [ ] Monitoring recommendations specific and measurable

#### 5.3 Validate Build & Test Results

**Checklist:**
- [ ] `make build` passes
- [ ] `make test` passes all pipelines
- [ ] `make ensure-goreleaser-up-to-date` passes (if applicable)
- [ ] All manifests generate valid output

#### 5.4 Get Approvals

- [ ] Code review: PR approved by maintainer
- [ ] Documentation reviewed for accuracy and actionability
- [ ] Release strategy and risk assessment approved

---

### Phase 6: Release (30 minutes)

#### 6.1 Merge PRs

Once approved:
```bash
# Merge Tulip PR first (images depend on manifest)
cd tulip && gh pr merge --squash --delete-branch

# Then merge Bonsai PR
cd bonsai && gh pr merge --squash --delete-branch
```

#### 6.2 Tag Release in Tulip

```bash
cd tulip
git checkout main
git pull origin main

make push-tags TAG=vYY.MM.0  # Creates signed git tag, pushes to origin
# This triggers .github/workflows/release-tulip.yaml
# - Builds multi-platform binaries (linux/darwin, amd64/arm64)
# - Publishes to ghcr.io/ollygarden/tulip:vYY.MM.0
```

#### 6.3 Publish GitHub Release

```bash
# Once release workflow completes, create GitHub Release with notes
gh release create vYY.MM.0 --title "Tulip vYY.MM.0" --notes "
# Tulip vYY.MM.0 Release

## Components
OTel Collector v0.X.Y

## New
- [component1] — purpose
- [component2] — purpose

## Known Issues
[Summary of CRITICAL/HIGH issues from docs]

See docs/tulip-vYY.MM-release-notes.md for full details.
See docs/tulip-vYY.MM-known-issues.md for deployment checklist.
"
```

---

### Phase 7: Deployment Synchronization (if applicable)

#### 7.1 Identify & Update Deployment Repositories

**Before proceeding, identify which repos reference Tulip distribution images:**

```bash
# Search for bonsai/Tulip image references in your repos
grep -r "ghcr.io/ollygarden/bonsai" /path/to/repos/**/*.yaml

# Common locations:
#   - Kubernetes Helm values files (*.yaml in charts/)
#   - Infrastructure config patches (Kustomize, Terraform, etc.)
#   - Docker Compose or deployment manifests
```

For each repo with version references:

```bash
cd repo-with-tulip-refs
git checkout -b sync-release-vYY.MM.0
# Update all distribution image tags to v0.X.Y
git add [files-with-version-references]
git commit -m "chore: sync to Tulip vYY.MM.0

Updated all distribution image references to v0.X.Y
See tulip#PR_NUMBER for release notes and audit.

Refs: E-XXXX"
gh pr create
```

#### 7.2 Merge & Deploy

- [ ] Merge deployment PRs in order of dependency
- [ ] Verify deployment pipeline executes
- [ ] Confirm services are running new version

---

### Phase 8: Public Communication (if applicable)

#### 8.1 Update Public Channels (if applicable)

If maintaining a public website or documentation site:

```bash
# Update version information, release notes, upgrade guides
cd public-docs-repo
git checkout -b release-vYY.MM.0-website
# Update version references, release summary, upgrade instructions
git add [docs/config files]
git commit -m "chore: publish Tulip vYY.MM.0 release information"
gh pr create
```

**Information to publish:**
- Release version and OTel component version
- Key changes and new features
- Link to comprehensive release documentation
- Upgrade instructions
- Known issues and mitigations

---

## Verification Checklist

**Before declaring release complete:**

- [ ] Release tagged in repository
- [ ] Artifacts published (binaries, container images, etc.)
- [ ] GitHub Release created with comprehensive notes
- [ ] Release documentation: `docs/tulip-vYY.MM-release.md`
  - [ ] Component manifest complete and accurate
  - [ ] Audit findings documented
  - [ ] Breaking changes identified
  - [ ] Known issues with mitigations
  - [ ] Deployment checklist included
- [ ] All downstream distribution PRs merged
- [ ] Deployment PRs merged (if applicable)
- [ ] Public communication updated (if applicable)
- [ ] Linear card/issue marked complete
- [ ] Stakeholders notified

---

## Key Files & References

**Tulip Repository:**
- `README.md` §Releases — release cadence and strategy
- `CLAUDE.md` §Adding or Updating Components — component workflow
- `Makefile` §push-tags — tagging mechanism
- `.github/workflows/release-tulip.yaml` — release automation
- `distributions/tulip/manifest.yaml` — component manifest (source of truth)

**Bonsai Repository:**
- `CLAUDE.md` — distribution structure, module path rules
- `distributions/*/manifest.yaml` — 4 distributions (nike, jjteste, external-collector, ollygarden-collector)

**Downstream:**
- seedbed-charts: `charts/observability/values*.yaml`
- seedbed: `clusters/aws-eu-central-1-{dev,internal}/infrastructure-config/patches/otel-collector*.yaml`

**Linear:**
- Create umbrella card per quarterly release
- Sub-cards: component audit, docs, distribution sync, web update (if separate cards)
- Reference in commit messages: `Refs: E-XXXX`

---

## Common Gotchas & Tips

### Version Synchronization
**Rule:** All OTel contrib components in a manifest must use **the same minor version**. Do not mix v0.151.0 and v0.152.0 in the same manifest.

```bash
# Check version consistency:
grep -h "gomod.*v0\." distributions/*/manifest.yaml | sort | uniq -c
# Should show one major.minor version (may have different patch levels, e.g., v1.57.0 vs v1.64.0 for providers)
```

### Provider Version Tracking
Providers (`envprovider`, `fileprovider`, `yamlprovider`) have a separate version track (v1.X.Y). Check upstream releases for the correct provider version for your OTel version.

### Breaking Changes
Always review breaking changes between versions. Update configs and document migrations:
- Example: drain processor config changed v0.151.0 → v0.158.0 (`extract_parameters` → `masking_rules`)
- Include migration examples in upgrade guide

### Component Health Audit
**Do not skip this step.** The audit identifies critical issues, security vulnerabilities, and performance concerns. Use the findings to:
1. Document known issues users must understand
2. Provide mitigation strategies
3. Set deployment expectations (production-ready vs. caution)

### Testing at Scale
Always test new receiver/processor versions at realistic scale before recommending production use:
- filelogreceiver: test at your file count (CPU issue known)
- hostmetricsreceiver: test on target platform (Windows, Linux, containers)
- filterprocessor: test filtering logic on real data (panic and logic bugs known)

### Rollback Plan
Keep previous version images available for quick rollback:
```bash
# Example: keep v0.151.0 images when releasing v0.158.0
docker pull ghcr.io/ollygarden/tulip:v0.151.0
docker tag ... gcr.io/my-registry/tulip:v0.151.0-backup
```

---

## Execution Notes

- **Phase 1 (Component Audit):** Can run in parallel with other prep work
- **Phase 2 (Version Updates):** Requires clean state; manifests must match across repositories
- **Phase 3 (Documentation):** Should incorporate Phase 1 findings comprehensively in single document
- **Phases 4-8:** Sequential; each depends on completion of previous phases
- **Dependencies:** Downstream deployment PRs depend on upstream release PRs being merged
- **Approval gates:** Code review and risk assessment required before Phase 6

---

## Questions?

Refer to:
- Tulip `README.md` for release philosophy
- Tulip `CLAUDE.md` for component workflow details
- Bonsai `CLAUDE.md` for distribution rules
- Prior releases (`docs/tulip-*-*.md` files) for examples
