---
name: release-version
description: Orchestrate a complete Tulip version release (component audit → version bump → documentation → image sync → web update)
---

# Tulip Release Skill

Orchestrate a complete Tulip quarterly release from start to finish, covering component evaluation, version updates across all distributions, documentation generation, and downstream synchronization.

## Prerequisites

**Before invoking this skill:**
- [ ] Decide target OTel version (check https://github.com/open-telemetry/opentelemetry-collector-releases/releases for latest)
- [ ] Planned release version (e.g., v26.08.0, follows vYY.MM.0 for quarterly)
- [ ] Linear project created for this release (e.g., "Tulip v26.08 Release")
- [ ] Access to: tulip, bonsai, seedbed-charts repos (and ollygarden.com if updating web)
- [ ] Time allocated (full audit + documentation ~4-6 hours)

---

## Release Workflow

### Phase 1: Component Audit (1-2 hours)

#### 1.1 Check Latest OTel Version

```bash
# Find latest stable release
curl -s https://api.github.com/repos/open-telemetry/opentelemetry-collector-releases/releases/latest | jq '.tag_name'

# Alternatively, check: https://github.com/open-telemetry/opentelemetry-collector-releases/releases
```

**Decision:** Choose version to release with (recommend latest for regular quarterly releases).

#### 1.2 Audit ALL Components (30+ components)

Run comprehensive upstream audit using an agent:

```
Agent Task:
- Audit all 29 components at target version
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

For each new component (e.g., drain, logdedup):

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

### Phase 2: Version Updates (1 hour)

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

#### 2.2 Update Bonsai Distributions

```bash
cd bonsai
# Update each distribution manifest:
# distributions/nike/manifest.yaml
# distributions/jjteste/manifest.yaml
# distributions/external-collector/manifest.yaml
# distributions/ollygarden-collector/manifest.yaml
#
# For each:
# - Update dist.version: NEW_VERSION
# - Replace all component versions to match Tulip
# - Sync with Tulip's manifest.yaml exactly
```

**Validation per distribution:**
```bash
make build DISTRIBUTION=nike       # Each must build
docker build -t nike:local distributions/nike/
```

#### 2.3 Update seedbed-charts Image Versions

**File:** `charts/observability/values.og-ext-prod-gke.yaml` (and similar for dev/internal)

```yaml
# External collector
- name: external-collector
  image: "ghcr.io/ollygarden/bonsai/external-collector:NEW_VERSION"

# Ollygarden collector gateway
- name: otel-collector
  image: "ghcr.io/ollygarden/bonsai/ollygarden-collector:NEW_VERSION"
```

**Validation:**
```bash
helm lint charts/observability
helm template test . > /tmp/manifests.yaml  # Verify valid K8s YAML
```

#### 2.4 Update seedbed Image Versions (Dev/Internal)

**Files:**
- `clusters/aws-eu-central-1-dev/infrastructure-config/patches/otel-collector*.yaml`
- `clusters/aws-eu-central-1-internal/infrastructure-config/patches/otel-collector*.yaml`

```yaml
- repository: ghcr.io/ollygarden/bonsai/ollygarden-collector
  tag: "NEW_VERSION"  # Update both dev and internal
```

---

### Phase 3: Documentation (1-2 hours)

#### 3.1 Release Notes Document

**File:** `docs/tulip-vYY.MM-release-notes.md`

**Contents:**
- Component version (v0.X.Y)
- What's new (added/removed/changed components)
- Full component manifest table (extensions, receivers, exporters, processors, connectors, providers)
- Configuration examples for new processors
- Integration notes and gotchas

**Reference:** Use prior LTS or quarterly release docs as template (e.g., `docs/tulip-lts-may2026-upgrade-plan.md`)

#### 3.2 Upgrade Guide

**File:** `docs/tulip-vYY.MM-upgrade-guide.md`

**Contents:**
- Component health summary table
- Critical issues + detailed mitigations
- Breaking changes and migration path
- Deployment checklist
- Integration examples

**Reference:** Use `docs/tulip-v26.08-upgrade-guide.md` as template

#### 3.3 Known Issues Document (if significant issues)

**File:** `docs/tulip-vYY.MM-known-issues.md`

**Contents:**
- Each issue: severity, component, status, symptoms, mitigation
- Security advisories (if any)
- Performance considerations
- Monitoring/alerting recommendations
- Timeline to upstream fixes

**Reference:** Use `docs/tulip-v26.08-known-issues.md` as comprehensive template

#### 3.4 Update Component Tracker (if applicable)

**File:** `docs/ollygarden-tracker.md`

**Process:**
- Regenerate tracker for new manifest version
- Diff against prior tracker
- Call out newly resolved issues in release notes

**Reference:** `README.md` §Releases and `CLAUDE.md` §Adding or Updating Components explain the process

---

### Phase 4: Git & PR Workflow (30 minutes)

#### 4.1 Create Release Branches & Commit

```bash
# Tulip
cd tulip
git checkout -b jonathan/tulip-release-vYY.MM.0
git add distributions/tulip/manifest.yaml go.sum docs/tulip-vYY.MM-*.md
git commit -m "feat(tulip): release vYY.MM.0 with v0.X.Y components

- Manifest: v0.X.Y components
- Added: release-notes.md, upgrade-guide.md, known-issues.md (if applicable)
- Component audit: N components audited, X CRITICAL/HIGH issues documented
- Validation: build ✓, test ✓, goreleaser ✓

Refs: E-XXXX (Linear card)"

# Bonsai
cd bonsai
git checkout -b jonathan/bonsai-sync-tulip-vYY.MM.0
git add distributions/*/manifest.yaml
git commit -m "chore(distributions): sync all to v0.X.Y for Tulip vYY.MM.0

- nike, jjteste, external-collector, ollygarden-collector: all → v0.X.Y
- Aligns with Tulip vYY.MM.0 release
- See tulip#PR_NUMBER for known issues and mitigations

Refs: E-XXXX"
```

#### 4.2 Create Coordinated PRs

```bash
# Push both branches
cd tulip && git push -u origin jonathan/tulip-release-vYY.MM.0
cd bonsai && git push -u origin jonathan/bonsai-sync-tulip-vYY.MM.0

# Create Tulip PR (body includes summary of component audit, known issues, deployment checklist)
cd tulip && gh pr create --title "feat(tulip): release vYY.MM.0 with v0.X.Y" --body "..."

# Create Bonsai PR (cross-reference Tulip PR)
cd bonsai && gh pr create --title "chore(distributions): sync all to v0.X.Y for Tulip vYY.MM.0" --body "..."
```

#### 4.3 Update Linear Release Card

**Card:** "tulip: release vYY.MM.0 and sync bonsai distributions"

**Add to description:**
```markdown
## In Flight

**Tulip PR:** {github_url}
- Manifest: v0.X.Y components
- Release docs: release-notes.md, upgrade-guide.md, known-issues.md
- Component audit results (N components, X critical issues)

**Bonsai PR:** {github_url}
- All distributions: v0.X.Y

**Critical findings:** [Link to known-issues.md or inline key issues]
```

---

### Phase 5: Code Review & Validation (1 hour)

#### 5.1 Review Component Audit Results

**Checklist:**
- [ ] All 29+ components listed
- [ ] CRITICAL/HIGH issues documented
- [ ] Security advisories flagged
- [ ] Breaking changes identified
- [ ] Migration guidance provided

#### 5.2 Review Documentation

**Checklist for each doc:**
- [ ] Accurate component counts and versions
- [ ] Configuration examples tested (or note as template)
- [ ] Deployment checklist is realistic
- [ ] Known issues are specific (issue numbers, links)
- [ ] Mitigations are actionable
- [ ] Monitoring recommendations are specific

#### 5.3 Validate Test Results

**Checklist:**
- [ ] `make build` passes
- [ ] `make test` passes all pipelines
- [ ] `make ensure-goreleaser-up-to-date` passes
- [ ] All distribution manifests generate valid Go code

#### 5.4 Get Approvals

- [ ] Code review: PR approved by team lead or maintainer
- [ ] Linear card: Acceptance criteria reviewed, critical issues acknowledged
- [ ] Deployment risk: Team concurs on risk level and mitigations

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

### Phase 7: Downstream Sync (1 hour)

#### 7.1 Create seedbed-charts PR

```bash
cd seedbed-charts
git checkout -b jonathan/sync-tulip-vYY.MM.0
# Update values files (external-collector, ollygarden-collector image versions)
git add charts/observability/values*.yaml
git commit -m "chore(observability): sync image versions to Tulip vYY.MM.0"
gh pr create
```

#### 7.2 Create seedbed PR

```bash
cd seedbed
git checkout -b jonathan/sync-tulip-vYY.MM.0-{dev,internal}
# Update otel-collector patch files (image tags for dev/internal)
git add clusters/aws-eu-central-1-dev/infrastructure-config/patches/otel-collector*.yaml
git add clusters/aws-eu-central-1-internal/infrastructure-config/patches/otel-collector*.yaml
git commit -m "chore(otel-collector): sync image versions to Tulip vYY.MM.0"
gh pr create
```

#### 7.3 Merge & Deploy

- [ ] seedbed-charts PR: Merge, Flux deploys to prod
- [ ] seedbed PR: Merge, Flux deploys to dev/internal
- [ ] Verify all environments running new version:
  ```bash
  kubectl get deployment otel-collector -A -o jsonpath='{.items[*].spec.template.spec.containers[0].image}'
  ```

---

### Phase 8: Web Page Update (30 minutes)

#### 8.1 Update ollygarden.com/tulip

**Location:** ollygarden.com site repo (may be separate from tulip/bonsai)

**Update:**
- Version number (e.g., "Latest: vYY.MM.0")
- Release date
- Key features / what's new (2-3 bullet points)
- Link to GitHub Release with full notes
- Link to upgrade guide (docs/tulip-vYY.MM-upgrade-guide.md)

**Example:**
```markdown
## Latest Release: vYY.MM.0

Released: [DATE]

**Latest OTel Collector:** v0.X.Y (updated [DATE])

**New in this release:**
- [Component] — brief description
- [Feature] — brief description
- [Improvement] — brief description

[Read full release notes](https://github.com/ollygarden/tulip/releases/tag/vYY.MM.0)
[Upgrade guide](https://github.com/ollygarden/tulip/blob/main/docs/tulip-vYY.MM-upgrade-guide.md)
```

#### 8.2 Create PR for Web Updates

```bash
cd [ollygarden.com site repo]
git checkout -b jonathan/tulip-vYY.MM.0-website
# Update release info
git add [files]
git commit -m "chore(tulip): update to vYY.MM.0"
gh pr create
# Merge when approved
```

---

## Verification Checklist

**Before declaring release complete:**

- [ ] Tulip vYY.MM.0 tagged in GitHub
- [ ] Binary published: `ghcr.io/ollygarden/tulip:vYY.MM.0`
- [ ] GitHub Release published with notes
- [ ] Release notes doc: `docs/tulip-vYY.MM-release-notes.md`
- [ ] Upgrade guide doc: `docs/tulip-vYY.MM-upgrade-guide.md`
- [ ] Known issues doc: `docs/tulip-vYY.MM-known-issues.md` (if applicable)
- [ ] Bonsai distributions synced (all 4 at v0.X.Y)
- [ ] seedbed-charts PR merged (prod image versions updated)
- [ ] seedbed PRs merged (dev/internal image versions updated)
- [ ] ollygarden.com/tulip website updated with new version
- [ ] Linear card marked complete
- [ ] Announcement sent (Slack, email, etc. if applicable)

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

## Time Estimate

**Full release workflow:** 4-6 hours
- Audit: 1-2 hours
- Updates: 1 hour
- Docs: 1-2 hours
- PR/review: 1 hour
- Release: 30 min
- Downstream sync: 1 hour
- Web update: 30 min

**Parallel tasks:** Audit and updates can run in parallel → reduces to 3-4 hours with concurrent work.

---

## Questions?

Refer to:
- Tulip `README.md` for release philosophy
- Tulip `CLAUDE.md` for component workflow details
- Bonsai `CLAUDE.md` for distribution rules
- Prior releases (`docs/tulip-*-*.md` files) for examples
