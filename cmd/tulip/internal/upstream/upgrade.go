package upstream

import (
	"github.com/ollygarden/tulip/cmd/tulip/internal/manifest"
)

// Upgrade updates the local manifest to match the upstream versions.
// It preserves custom components but updates their versions to match the upstream OTel version.
func Upgrade(local *manifest.Manifest, upstream *manifest.Manifest) {
	contribVersion := upstream.OTelVersion()
	coreVersion := findCoreVersion(upstream)

	// Update dist version to match upstream
	local.Dist.Version = upstream.Dist.Version

	// Update all component versions
	manifest.MergeVersions(local, contribVersion, coreVersion)

	// Add any new base components that are in upstream but not in local
	upstreamComps := upstream.AllComponents()
	for _, uc := range upstreamComps {
		if !local.HasComponent(uc.Name()) {
			_ = local.AddComponent(uc)
		}
	}
}

// findCoreVersion extracts the version used by core components in the manifest.
// Core components (go.opentelemetry.io/collector/...) may use a different version scheme.
func findCoreVersion(m *manifest.Manifest) string {
	for _, c := range m.AllComponents() {
		if c.IsCore() {
			return c.Version()
		}
	}
	return m.OTelVersion()
}
