# Tulip CLI & Multi-Distribution Platform

## Overview

Transform Tulip from a single-distribution repo into a **multi-distribution platform** where users fork once and create as many custom OpenTelemetry Collector images as they need — each with its own component set, config, and Docker image.

The experience is powered by:

- **`tulip` CLI** — interactive component browser and distribution scaffolder
- **GitHub Actions** — auto-builds and pushes a Docker image for every distribution on push
- **One folder = one image** — simple mental model, zero magic

---

## Architecture

### Repository Layout

```
distributions/
├── tulip/                    # OllyGarden base (always present, supported)
│   ├── manifest.yaml         # Source of truth: all supported components
│   ├── config.yaml           # Default runtime config
│   ├── Dockerfile
│   ├── entrypoint.sh
│   ├── tulip.service
│   ├── tulip.conf
│   ├── preinstall.sh
│   ├── postinstall.sh
│   └── preremove.sh
│
├── my-k8s-collector/         # User-created: Kubernetes-focused
│   ├── manifest.yaml         # tulip base + k8sattributes, prometheus, etc.
│   ├── config.yaml
│   ├── Dockerfile
│   ├── entrypoint.sh
│   ├── my-k8s-collector.service
│   ├── my-k8s-collector.conf
│   ├── preinstall.sh
│   ├── postinstall.sh
│   └── preremove.sh
│
├── my-aws-collector/         # User-created: AWS-focused
│   ├── manifest.yaml         # tulip base + awsxray, awsemf, etc.
│   ├── config.yaml
│   ├── Dockerfile
│   └── ...
│
└── my-lightweight/           # User-created: Minimal forwarder
    ├── manifest.yaml         # tulip base minus some components
    ├── config.yaml
    ├── Dockerfile
    └── ...
```

### Build Flow

```
distributions/my-k8s-collector/manifest.yaml
        │
        ▼
    ocb (OpenTelemetry Collector Builder)
        │
        ▼
    _build/components.go, main.go, go.mod
        │
        ▼
    go build → binary
        │
        ▼
    docker build → ghcr.io/<user>/tulip/my-k8s-collector:latest
```

### CI Output (one image per folder)

```
ghcr.io/<user>/tulip/tulip              ← base (OllyGarden supported)
ghcr.io/<user>/tulip/my-k8s-collector   ← custom
ghcr.io/<user>/tulip/my-aws-collector   ← custom
ghcr.io/<user>/tulip/my-lightweight     ← custom
```

---

## CLI Design

### Tech Stack

| Layer          | Library                  | Purpose                                          |
|----------------|--------------------------|--------------------------------------------------|
| Commands       | `spf13/cobra`            | Subcommand routing, flags, help, shell completion |
| Interactive UX | `charmbracelet/huh` v2   | Filterable multi-select, forms, spinners          |
| Styled output  | `charmbracelet/lipgloss` | Tables, colors, borders                           |

### Project Structure

```
cmd/tulip/
├── main.go
├── cmd/
│   ├── root.go           # cobra root command
│   ├── create.go         # tulip create <name>
│   ├── add.go            # tulip add -d <dist>
│   ├── remove.go         # tulip remove -d <dist>
│   ├── list.go           # tulip list [-d <dist>]
│   ├── build.go          # tulip build [-d <dist>]
│   ├── validate.go       # tulip validate [-d <dist>]
│   ├── doctor.go         # tulip doctor (fetch upstream, compare versions)
│   ├── upgrade.go        # tulip upgrade [-d <dist>]
│   └── catalog.go        # tulip catalog search <query>
└── internal/
    ├── catalog/           # OTel component registry fetcher + cache
    │   ├── catalog.go     # Fetch & parse builder-config.yaml from contrib
    │   ├── metadata.go    # Fetch per-component metadata.yaml for descriptions
    │   └── cache.go       # Local file cache (~/.tulip/catalog/)
    ├── manifest/          # manifest.yaml read/write/merge logic
    │   ├── manifest.go    # Parse and generate OCB manifest files
    │   └── merge.go       # Merge base tulip manifest + user additions/removals
    ├── upstream/          # Upstream version fetching + comparison
    │   ├── fetch.go       # Fetch manifest from ollygarden/tulip via GitHub API
    │   ├── compare.go     # Compare local vs upstream versions, produce report
    │   └── upgrade.go     # Apply upstream version updates to local manifests
    ├── scaffold/          # Distribution folder scaffolding
    │   ├── scaffold.go    # Generate all files for a new distribution
    │   └── templates/     # Go templates for Dockerfile, config, systemd, etc.
    └── ui/                # Shared lipgloss styles and formatters
        └── styles.go
```

### Commands

#### `tulip create <name>`

Creates a new distribution folder with interactive component selection.

```
$ tulip create my-k8s-collector

  Creating new Tulip distribution: my-k8s-collector

  ℹ Starting with all OllyGarden-supported components (tulip base)

  Would you like to add extra components?  Yes

  ── Receivers ──────────────────────────────────────────────

  Filter: k8s▌

  ☑ k8sclusterreceiver        Kubernetes cluster metrics     beta
  ☑ k8seventsreceiver          Kubernetes events              alpha
  ☑ k8sobjectsreceiver         Kubernetes objects             alpha
  ☐ kafkareceiver              Consume from Kafka topics      beta
  ☐ kubeletstatsreceiver       Kubelet stats                  beta

  ↑/↓ navigate • space select • / filter • enter confirm

  ── Processors ─────────────────────────────────────────────

  ☑ k8sattributesprocessor     Enrich with K8s metadata       beta
  ☑ resourcedetectionprocessor Auto-detect cloud resources    beta

  ── Exporters ──────────────────────────────────────────────

  (no extra exporters selected)

  Would you like to remove any base components?  No

  ✓ Created distributions/my-k8s-collector/
    ├── manifest.yaml  (base + 5 custom components)
    ├── config.yaml
    ├── Dockerfile
    └── ...

  Next steps:
    tulip build -d my-k8s-collector    Build locally
    git add . && git push              CI builds & pushes image
```

**Non-interactive mode** (for scripting/CI):

```bash
tulip create my-k8s-collector \
  --add-receiver k8scluster,k8sevents \
  --add-processor k8sattributes,resourcedetection
```

#### `tulip add -d <distribution>`

Add components to an existing distribution.

```
$ tulip add -d my-k8s-collector

  Adding components to: my-k8s-collector

  Select component type:  Receiver

  Filter: prom▌

  ☑ prometheusreceiver         Scrape Prometheus targets      beta
  ☐ prometheussimplereceiver   Simple Prometheus scraper      alpha

  ✓ Added 1 receiver to my-k8s-collector
    Updated: distributions/my-k8s-collector/manifest.yaml
```

**Non-interactive:**

```bash
tulip add -d my-k8s-collector --receiver prometheus --exporter kafka
```

#### `tulip remove -d <distribution>`

Remove components from a distribution. Shows which are base (OllyGarden-supported) vs custom (user-added), warns when removing base components.

```
$ tulip remove -d my-k8s-collector

  Select components to remove from: my-k8s-collector

  ☐ otlpreceiver          receiver    base
  ☐ k8sclusterreceiver    receiver    custom
  ☑ k8seventsreceiver     receiver    custom
  ☐ batchprocessor        processor   base

  ✓ Removed 1 component from my-k8s-collector
```

#### `tulip list`

List all distributions in the repo.

```
$ tulip list

  Tulip Distributions

  ┌─────────────────────┬────────────┬─────────────────────────────────┬────────┐
  │ Distribution        │ Components │ Image                           │ Source │
  ├─────────────────────┼────────────┼─────────────────────────────────┼────────┤
  │ tulip               │ 22         │ ghcr.io/<user>/tulip/tulip      │ base   │
  │ my-k8s-collector    │ 27 (+5)    │ ghcr.io/<user>/tulip/my-k8s-…  │ custom │
  │ my-aws-collector    │ 25 (+3)    │ ghcr.io/<user>/tulip/my-aws-…  │ custom │
  │ my-lightweight      │ 18 (-4)    │ ghcr.io/<user>/tulip/my-ligh…  │ custom │
  └─────────────────────┴────────────┴─────────────────────────────────┴────────┘
```

#### `tulip list -d <distribution>`

List components in a specific distribution.

```
$ tulip list -d my-k8s-collector

  my-k8s-collector                                    v0.145.0

  ┌────────────────────────────────┬────────────┬─────────┬────────┐
  │ Component                      │ Type       │ Status  │ Source │
  ├────────────────────────────────┼────────────┼─────────┼────────┤
  │ otlpreceiver                   │ receiver   │ stable  │ base   │
  │ nopreceiver                    │ receiver   │ stable  │ base   │
  │ k8sclusterreceiver             │ receiver   │ beta    │ custom │
  │ k8seventsreceiver              │ receiver   │ alpha   │ custom │
  │ k8sobjectsreceiver             │ receiver   │ alpha   │ custom │
  │ batchprocessor                 │ processor  │ stable  │ base   │
  │ attributesprocessor            │ processor  │ stable  │ base   │
  │ k8sattributesprocessor         │ processor  │ beta    │ custom │
  │ resourcedetectionprocessor     │ processor  │ beta    │ custom │
  │ ...                            │            │         │        │
  └────────────────────────────────┴────────────┴─────────┴────────┘

  base = supported by OllyGarden │ custom = user-added
```

#### `tulip build [-d <distribution>]`

Build one or all distributions. Wraps `make build` with progress feedback.

```
$ tulip build -d my-k8s-collector

  ● Building my-k8s-collector...
    ├── Generating sources (ocb)       ✓
    ├── Resolving dependencies         ✓
    ├── Compiling binary               ✓
    └── Done in 42s

  Binary: distributions/my-k8s-collector/_build/my-k8s-collector

$ tulip build -d my-k8s-collector --docker

  ● Building my-k8s-collector (docker)...
    ├── Generating sources (ocb)       ✓
    ├── Compiling binary (linux/amd64) ✓
    ├── Building Docker image          ✓
    └── Done in 58s

  Image: my-k8s-collector:local

$ tulip build

  ● Building all distributions...
    ├── tulip                          ✓  (38s)
    ├── my-k8s-collector               ✓  (42s)
    ├── my-aws-collector               ✓  (45s)
    └── my-lightweight                 ✓  (35s)
```

#### `tulip validate [-d <distribution>]`

Validates manifest, version compatibility, and build.

```
$ tulip validate -d my-k8s-collector

  Validating my-k8s-collector...

  ✓ manifest.yaml syntax valid
  ✓ All component versions compatible (v0.145.0)
  ✓ No duplicate components
  ✓ Build compiles successfully
  ✓ All base components present

  my-k8s-collector is valid
```

#### `tulip catalog search <query>`

Search the OTel component catalog without leaving the terminal.

```
$ tulip catalog search kafka

  Results for "kafka"

  ┌──────────────────────────┬───────────┬─────────┬─────────────────────────────┐
  │ Component                │ Type      │ Status  │ Description                 │
  ├──────────────────────────┼───────────┼─────────┼─────────────────────────────┤
  │ kafkareceiver            │ receiver  │ beta    │ Consume from Kafka topics   │
  │ kafkametricsreceiver     │ receiver  │ beta    │ Kafka broker metrics        │
  │ kafkaexporter            │ exporter  │ beta    │ Export to Kafka topics      │
  └──────────────────────────┴───────────┴─────────┴─────────────────────────────┘
```

---

## `tulip doctor`

Fetches the latest tulip base manifest from the upstream `ollygarden/tulip` repo and validates every local distribution against it. This is the single command to answer: **"Are my distributions up to date with OllyGarden's supported base?"**

### What it Checks

| Check | What it does |
|-------|-------------|
| **Base version drift** | Compares each distribution's OllyGarden-supported component versions against upstream `ollygarden/tulip` |
| **Component alignment** | Verifies that custom-added components use the same OTel version as the base |
| **New base components** | Detects if OllyGarden added new supported components the user doesn't have yet |
| **Removed components** | Warns if the upstream base dropped a component the user still includes |
| **dist.version field** | Checks the manifest `dist.version` matches upstream |

### How it Works

1. Fetches `distributions/tulip/manifest.yaml` from `github.com/ollygarden/tulip` (via GitHub raw content API)
2. Parses the upstream manifest to extract current versions and component list
3. For each local `distributions/*/manifest.yaml`:
   - Compares gomod versions against upstream
   - Identifies base components, custom components, and missing new components
4. Outputs a per-distribution health report

### UX: Outdated Distributions

```
$ tulip doctor

  Fetching latest tulip base from ollygarden/tulip...
  Latest base version: v0.148.0

  ── tulip (base) ──────────────────────────────────────────
  ⚠ Outdated: local v0.145.0 → upstream v0.148.0
    18 components need version bump

  ── my-k8s-collector ──────────────────────────────────────
  ⚠ Outdated: local v0.145.0 → upstream v0.148.0
    23 components need version bump (18 base + 5 custom)
  ℹ New base component available:
    + memorylimiterprocessor (added in v0.147.0)

  ── my-lightweight ────────────────────────────────────────
  ⚠ Outdated: local v0.145.0 → upstream v0.148.0
    14 components need version bump

  Summary:
    3 distributions outdated

  Run 'tulip upgrade' to update all distributions automatically.
  Run 'tulip upgrade -d my-k8s-collector' to update one.
```

### UX: All Up to Date

```
$ tulip doctor

  Fetching latest tulip base from ollygarden/tulip...
  Latest base version: v0.148.0

  ── tulip (base) ──────────────────────────────────────────
  ✓ Up to date (v0.148.0)

  ── my-k8s-collector ──────────────────────────────────────
  ✓ Up to date (v0.148.0, +5 custom components)

  All distributions are up to date.
```

### Non-interactive Mode (for CI)

```bash
# Exit code 0 = up to date, exit code 1 = outdated
tulip doctor --check

# JSON output for tooling (always exits 0, check .outdated field)
tulip doctor --json
```

### JSON Output Schema

```json
{
  "upstream_version": "0.148.0",
  "outdated": true,
  "distributions": [
    {
      "name": "tulip",
      "local_version": "0.145.0",
      "upstream_version": "0.148.0",
      "outdated": true,
      "outdated_components": 18,
      "new_base_components": ["memorylimiterprocessor"],
      "removed_base_components": []
    },
    {
      "name": "my-k8s-collector",
      "local_version": "0.145.0",
      "upstream_version": "0.148.0",
      "outdated": true,
      "outdated_components": 23,
      "new_base_components": ["memorylimiterprocessor"],
      "removed_base_components": []
    }
  ]
}
```

### The `tulip upgrade` Companion

When `tulip doctor` reports drift, `tulip upgrade` fixes it:

```bash
# Update all distributions to match the latest upstream base
tulip upgrade

# Update a single distribution
tulip upgrade -d my-k8s-collector
```

What `tulip upgrade` does:
1. Fetches the latest upstream `distributions/tulip/manifest.yaml`
2. For each distribution:
   - Updates all base component versions to match upstream
   - Updates custom component versions to the matching OTel version
   - Updates the `dist.version` field
   - Adds any new base components
3. Runs `make generate-sources` to regenerate build files
4. Runs `go mod tidy`

---

## GitHub Action: `ollygarden/tulip-doctor`

A published GitHub Action that works like Dependabot. Users add one line to their workflow, and it automatically checks if their distributions are outdated and creates issues.

### Usage (for users)

```yaml
# .github/workflows/tulip-doctor.yaml
name: Tulip Doctor

on:
  schedule:
    - cron: '0 8 * * 1'  # Every Monday at 8am UTC
  workflow_dispatch:

jobs:
  doctor:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: ollygarden/tulip-doctor@v1
        with:
          github-token: ${{ secrets.GITHUB_TOKEN }}
```

That's it. Two steps after checkout.

### What It Does

```
Every Monday 8am UTC (or manual trigger)
        │
        ▼
  ollygarden/tulip-doctor@v1
        │
        ├── Installs tulip CLI (go install)
        ├── Runs tulip doctor --json
        ├── Fetches upstream ollygarden/tulip manifest
        ├── Compares versions across all local distributions
        │
        ▼
  For each outdated distribution:
        │
        ├── Check: does an issue already exist? (dedup by title)
        │     ├── Yes → skip
        │     └── No  → create issue
        │
        ▼
  Issues appear in your repo's Issues tab
        │
        ├── 📦 tulip: outdated v0.145.0 → v0.148.0
        ├── 📦 my-k8s-collector: outdated v0.145.0 → v0.148.0
        └── ...
```

### Action Definition

The action lives in its own repo: `ollygarden/tulip-doctor` (or as a path inside the tulip repo).

```yaml
# action.yml
name: 'Tulip Doctor'
description: 'Checks if your Tulip distributions are up to date with the OllyGarden supported base'
branding:
  icon: 'activity'
  color: 'green'

inputs:
  github-token:
    description: 'GitHub token for creating issues'
    required: true
  go-version:
    description: 'Go version to use'
    required: false
    default: '1.25'
  labels:
    description: 'Comma-separated labels for created issues'
    required: false
    default: 'tulip-doctor,dependencies'
  assignees:
    description: 'Comma-separated GitHub usernames to assign issues to'
    required: false
    default: ''
  create-issues:
    description: 'Create GitHub issues for outdated distributions'
    required: false
    default: 'true'

outputs:
  outdated:
    description: 'Whether any distributions are outdated (true/false)'
    value: ${{ steps.doctor.outputs.outdated }}
  report:
    description: 'Path to the JSON report file'
    value: ${{ steps.doctor.outputs.report }}

runs:
  using: 'composite'
  steps:
    - name: Setup Go
      uses: actions/setup-go@v5
      with:
        go-version: ${{ inputs.go-version }}

    - name: Install tulip CLI
      shell: bash
      run: go install github.com/ollygarden/tulip/cmd/tulip@latest

    - name: Run tulip doctor
      id: doctor
      shell: bash
      run: |
        tulip doctor --json > tulip-doctor-report.json 2>&1 || true
        echo "report=tulip-doctor-report.json" >> "$GITHUB_OUTPUT"

        if jq -e '.outdated == true' tulip-doctor-report.json > /dev/null 2>&1; then
          echo "outdated=true" >> "$GITHUB_OUTPUT"
        else
          echo "outdated=false" >> "$GITHUB_OUTPUT"
        fi

    - name: Create issues for outdated distributions
      if: steps.doctor.outputs.outdated == 'true' && inputs.create-issues == 'true'
      shell: bash
      env:
        GH_TOKEN: ${{ inputs.github-token }}
      run: |
        UPSTREAM=$(jq -r '.upstream_version' tulip-doctor-report.json)
        LABELS="${{ inputs.labels }}"
        ASSIGNEES="${{ inputs.assignees }}"

        ASSIGNEE_FLAG=""
        if [ -n "$ASSIGNEES" ]; then
          ASSIGNEE_FLAG="--assignee ${ASSIGNEES}"
        fi

        jq -c '.distributions[] | select(.outdated == true)' tulip-doctor-report.json | while read -r dist; do
          NAME=$(echo "$dist" | jq -r '.name')
          LOCAL=$(echo "$dist" | jq -r '.local_version')
          OUTDATED_COUNT=$(echo "$dist" | jq -r '.outdated_components')
          NEW_COMPONENTS=$(echo "$dist" | jq -r '.new_base_components | join(", ")')

          TITLE="tulip-doctor: ${NAME} is outdated (v${LOCAL} → v${UPSTREAM})"

          # Dedup: skip if issue already exists
          EXISTING=$(gh issue list \
            --label "tulip-doctor" \
            --search "${NAME} is outdated v${UPSTREAM} in:title" \
            --state open \
            --json number \
            --jq '.[0].number // empty')

          if [ -n "$EXISTING" ]; then
            echo "Issue #${EXISTING} already exists for ${NAME}, skipping."
            continue
          fi

          # Build issue body
          BODY="## Tulip Distribution Outdated

        The distribution **${NAME}** is using an older version of the OllyGarden-supported base.

        | | Version |
        |---|---------|
        | **Current** | v${LOCAL} |
        | **Latest (OllyGarden)** | v${UPSTREAM} |
        | **Components to update** | ${OUTDATED_COUNT} |"

          if [ -n "$NEW_COMPONENTS" ] && [ "$NEW_COMPONENTS" != "" ]; then
            BODY="${BODY}

        ### New base components available

        The upstream tulip base added new supported components that this distribution doesn't include yet:

        \`\`\`
        ${NEW_COMPONENTS}
        \`\`\`"
          fi

          BODY="${BODY}

        ### How to fix

        **Option 1: Automatic** (recommended)
        \`\`\`bash
        tulip upgrade -d ${NAME}
        \`\`\`

        **Option 2: Manual**
        Update the component versions in \`distributions/${NAME}/manifest.yaml\` to match \`v${UPSTREAM}\`.

        ---
        *This issue was automatically created by [tulip-doctor](https://github.com/ollygarden/tulip).*"

          gh issue create \
            --title "$TITLE" \
            --label "$LABELS" \
            $ASSIGNEE_FLAG \
            --body "$BODY"

          echo "Created issue for ${NAME}"
        done
```

### Issue Format

When a distribution is outdated, the action creates an issue like this:

**Title:** `tulip-doctor: my-k8s-collector is outdated (v0.145.0 → v0.148.0)`

**Labels:** `tulip-doctor`, `dependencies`

**Body:**

> ## Tulip Distribution Outdated
>
> The distribution **my-k8s-collector** is using an older version of the OllyGarden-supported base.
>
> | | Version |
> |---|---------|
> | **Current** | v0.145.0 |
> | **Latest (OllyGarden)** | v0.148.0 |
> | **Components to update** | 23 |
>
> ### New base components available
>
> The upstream tulip base added new supported components that this distribution doesn't include yet:
>
> ```
> memorylimiterprocessor
> ```
>
> ### How to fix
>
> **Option 1: Automatic** (recommended)
> ```bash
> tulip upgrade -d my-k8s-collector
> ```
>
> **Option 2: Manual**
> Update the component versions in `distributions/my-k8s-collector/manifest.yaml` to match `v0.148.0`.

### Deduplication

- Issues are deduplicated by searching for the distribution name + upstream version in the title
- If an open issue already exists for the same distribution and version, it's skipped
- When the user fixes the issue (runs `tulip upgrade`), the next scan won't recreate it because versions will match

### Advanced Usage

```yaml
# Custom labels and assignees
- uses: ollygarden/tulip-doctor@v1
  with:
    github-token: ${{ secrets.GITHUB_TOKEN }}
    labels: 'tulip-doctor,dependencies,team-platform'
    assignees: 'jonathansilva,teammate'

# Just check, don't create issues (use output in later steps)
- uses: ollygarden/tulip-doctor@v1
  id: check
  with:
    github-token: ${{ secrets.GITHUB_TOKEN }}
    create-issues: 'false'

- name: Fail if outdated
  if: steps.check.outputs.outdated == 'true'
  run: |
    echo "Distributions are outdated!"
    cat ${{ steps.check.outputs.report }}
    exit 1
```

### How It Compares to Dependabot

| | Dependabot | tulip-doctor |
|---|-----------|-------------|
| **What it checks** | Package manager dependencies | OllyGarden base component versions |
| **How users enable it** | `dependabot.yml` config | `uses: ollygarden/tulip-doctor@v1` |
| **Output** | PRs with version bumps | Issues alerting about drift |
| **Fix** | Merge the PR | Run `tulip upgrade` or merge a `tulip-upgrade` PR |
| **Custom config** | `dependabot.yml` | Action inputs (labels, assignees, schedule) |
| **Deduplication** | Built-in | By issue title search |

---

## Component Catalog

The CLI maintains a local catalog of all ~265 available OpenTelemetry Collector components.

### Data Sources

| Source | What it provides | URL |
|--------|-----------------|-----|
| `builder-config.yaml` | All component gomod paths + versions | `github.com/open-telemetry/opentelemetry-collector-contrib/cmd/otelcontribcol/builder-config.yaml` |
| Per-component `metadata.yaml` | Display name, description, stability, signals | `github.com/open-telemetry/opentelemetry-collector-contrib/{type}/{name}/metadata.yaml` |

### How it Works

1. `tulip catalog update` fetches the latest `builder-config.yaml` from the contrib repo
2. For each component, fetches its `metadata.yaml` to get description and stability info
3. Caches everything locally at `~/.tulip/catalog/` (refreshed on `catalog update` or weekly)
4. The `create` and `add` commands read from the cache for instant results

### Component Naming Convention

The gomod path for any component follows a predictable pattern:

```
# Contrib components
github.com/open-telemetry/opentelemetry-collector-contrib/{type}/{name}{type}

# Core components
go.opentelemetry.io/collector/{type}/{name}{type}
```

Examples:
```
github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kafkareceiver
github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sattributesprocessor
go.opentelemetry.io/collector/receiver/otlpreceiver
```

---

## Scaffold Templates

When `tulip create <name>` runs, it generates the following files from templates. All references to the distribution name are parameterized.

### `manifest.yaml`

Starts with the full tulip base manifest and appends user-selected components:

```yaml
dist:
  module: github.com/ollygarden/tulip/{{.Name}}
  name: {{.Name}}
  description: {{.Description}}
  version: 0.145.0
  output_path: ./_build
  build_tags: "grpcnotrace"

# === OllyGarden-supported components (base) ===
# These components are covered by OllyGarden support.
# Do not remove this section — it ensures your distribution
# stays compatible with OllyGarden updates.

extensions:
  - gomod: go.opentelemetry.io/collector/extension/zpagesextension v0.145.0
  # ... (all base extensions)

receivers:
  - gomod: go.opentelemetry.io/collector/receiver/nopreceiver v0.145.0
  - gomod: go.opentelemetry.io/collector/receiver/otlpreceiver v0.145.0

exporters:
  # ... (all base exporters)

processors:
  # ... (all base processors)

connectors:
  # ... (all base connectors)

providers:
  # ... (all base providers)

# === Custom components (user-added) ===
# These components are managed by you. OllyGarden support
# does not cover custom additions.
#
# Added via: tulip add -d {{.Name}}
# Remove via: tulip remove -d {{.Name}}

  # receivers:
  #   - gomod: github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kafkareceiver v0.145.0
```

> **Note:** The manifest.yaml is a single file. Base and custom components are separated by comments for clarity. The user owns this file entirely — they can hand-edit it any time. The CLI just makes editing easier.

### `Dockerfile`

```dockerfile
FROM alpine:latest

LABEL org.opencontainers.image.title="{{.Name}}" \
      org.opencontainers.image.description="{{.Description}}" \
      org.opencontainers.image.vendor="{{.Org}}"

ARG USER_UID=10001
USER ${USER_UID}

COPY --chmod=755 {{.Name}} /{{.Name}}
COPY --chmod=755 entrypoint.sh /entrypoint.sh
COPY config.yaml /etc/{{.Name}}/config.yaml
ENTRYPOINT ["/entrypoint.sh"]
CMD []
EXPOSE 4317
```

### `entrypoint.sh`

```bash
#!/bin/sh
set -e

CONFIG_PATHS="/etc/{{.Name}}/config.yaml /etc/tulip/config.yaml /etc/otelcol-contrib/config.yaml /etc/otelcol/config.yaml"

find_config() {
    for path in $CONFIG_PATHS; do
        if [ -f "$path" ]; then
            echo "$path"
            return 0
        fi
    done
    return 1
}

if ! echo "$@" | grep -q -- "--config"; then
    if CONFIG_FILE=$(find_config); then
        echo "Using config file: $CONFIG_FILE"
        exec /{{.Name}} --config="$CONFIG_FILE" "$@"
    else
        echo "ERROR: No config file found."
        exit 1
    fi
else
    exec /{{.Name}} "$@"
fi
```

### `config.yaml`

A starter config that references only the components included in the distribution.

### systemd files

- `{{.Name}}.service` — systemd unit file
- `{{.Name}}.conf` — environment variables (`OTELCOL_OPTIONS="--config=/etc/{{.Name}}/config.yaml"`)
- `preinstall.sh` — creates `{{.Name}}` system user/group
- `postinstall.sh` — enables the systemd service
- `preremove.sh` — stops and disables the service

---

## Implementation Plan

### Phase 1: Foundation

1. **Manifest parser** — read/write `manifest.yaml`, extract component list and versions
2. **Component catalog fetcher** — download and cache `builder-config.yaml` + per-component `metadata.yaml` from the OTel contrib repo
3. **Scaffold templates** — Dockerfile, entrypoint.sh, config.yaml, systemd files, all parameterized by distribution name
4. **Upstream fetcher** — fetch `distributions/tulip/manifest.yaml` from `ollygarden/tulip` via GitHub raw content API

### Phase 2: CLI Core

5. **CLI skeleton** — cobra root command with global flags (`-d` for distribution)
6. **`tulip create <name>`** — interactive flow with huh multi-select + scaffold generation
7. **`tulip add -d <name>`** — add components to existing distribution
8. **`tulip remove -d <name>`** — remove components
9. **`tulip list`** — list distributions and components with lipgloss tables

### Phase 3: Build & Validate

10. **`tulip build`** — wrapper around `make build` with progress feedback via huh spinner
11. **`tulip validate`** — check manifest syntax, version compatibility, and build
12. **`tulip catalog search`** — search component catalog from the terminal

### Phase 4: Doctor & Upgrade

13. **`tulip doctor`** — fetch upstream `ollygarden/tulip` base manifest, compare versions across all local distributions, report drift
14. **`tulip doctor --json`** — structured JSON output for CI/tooling consumption
15. **`tulip doctor --check`** — non-interactive mode (exit code 0 = up to date, 1 = outdated)
16. **`tulip upgrade`** — apply upstream version updates to all distributions automatically

### Phase 5: CI Automation

17. **`build-distributions.yaml`** — GitHub Actions workflow with dynamic matrix to build + push all distribution images
18. **`ollygarden/tulip-doctor` GitHub Action** — published composite action: installs CLI via `go install`, runs `tulip doctor`, creates one GitHub issue per outdated distribution with deduplication
19. **`tulip-upgrade.yaml`** — optional workflow: runs `tulip upgrade` and opens a PR when upstream changes
20. **Multi-arch support** — extend the build workflow for `linux/amd64` and `linux/arm64`
21. **Release workflow** — tag-based releases with goreleaser for each distribution

### Phase 6: Polish

22. **Pre-built profiles** — `tulip create my-collector --profile kubernetes` with curated component sets for common use cases (Kubernetes, AWS, GCP, minimal)
23. **Shell completion** — bash, zsh, fish auto-completion via cobra
24. **Documentation** — update README with the new workflow

---

## CLI Installation

```bash
# One command. That's it.
go install github.com/ollygarden/tulip/cmd/tulip@latest
```

Alternative install methods (no Go required):

```bash
# From GitHub releases
curl -sfL https://github.com/ollygarden/tulip/releases/latest/download/tulip_$(uname -s)_$(uname -m).tar.gz | tar xz
sudo mv tulip /usr/local/bin/

# Via Homebrew
brew install ollygarden/tap/tulip
```

### Use it in GitHub Actions

```yaml
- uses: actions/setup-go@v5
  with:
    go-version: '1.25'

- run: go install github.com/ollygarden/tulip/cmd/tulip@latest

- run: tulip doctor --check --json
```

Three steps. The CLI is the action.

---

## What Makes This a Great UX

| Feature | Why it matters |
|---------|---------------|
| **One folder = one image** | Simple mental model. No config merging surprises. |
| **Interactive component browser** | Users don't need to know OTel module paths. Type to filter, space to select. |
| **Dual-mode commands** | Every command works interactively AND via flags. Humans get TUI, CI gets flags. |
| **Base vs custom labeling** | Clear support boundaries everywhere. Users always know what OllyGarden covers. |
| **Push-to-build** | Edit manifest, push, get image. No local Go or Docker needed. |
| **Clean upgrades** | `git pull upstream main` updates the base. Custom distributions don't conflict. |
| **Parallel CI builds** | GitHub Actions matrix builds all distributions concurrently. |
| **Zero Go knowledge needed** | The CLI handles `ocb`, `go mod tidy`, build flags — all hidden behind simple verbs. |
