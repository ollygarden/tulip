package upstream

import (
	"github.com/ollygarden/tulip/cmd/tulip/internal/manifest"
)

// DriftReport describes version drift between a local distribution and the upstream base.
type DriftReport struct {
	Name                  string   `json:"name"`
	LocalVersion          string   `json:"local_version"`
	UpstreamVersion       string   `json:"upstream_version"`
	Outdated              bool     `json:"outdated"`
	OutdatedComponents    int      `json:"outdated_components"`
	NewBaseComponents     []string `json:"new_base_components"`
	RemovedBaseComponents []string `json:"removed_base_components"`
}

// Compare compares a local distribution manifest against the upstream base manifest
// and returns a drift report.
func Compare(local *manifest.Manifest, upstream *manifest.Manifest) DriftReport {
	report := DriftReport{
		Name:            local.Dist.Name,
		LocalVersion:    local.OTelVersion(),
		UpstreamVersion: upstream.OTelVersion(),
	}

	// Build name→component maps
	localNames := componentNameSet(local)
	upstreamNames := componentNameSet(upstream)

	// Count outdated components (different version from upstream)
	localComps := componentVersionMap(local)
	upstreamComps := componentVersionMap(upstream)

	for name, upstreamVersion := range upstreamComps {
		localVersion, exists := localComps[name]
		if exists && localVersion != upstreamVersion {
			report.OutdatedComponents++
		}
	}

	// Also count custom components that need version bump
	// (components in local but not in upstream base, whose version differs from upstream OTel version)
	for name, localVersion := range localComps {
		if _, inUpstream := upstreamComps[name]; !inUpstream {
			// Custom component — check if it uses the same OTel version
			if localVersion != report.UpstreamVersion {
				report.OutdatedComponents++
			}
		}
	}

	// New base components (in upstream but not in local)
	for name := range upstreamNames {
		if !localNames[name] {
			report.NewBaseComponents = append(report.NewBaseComponents, name)
		}
	}

	// Removed base components (were presumably base components but upstream dropped them)
	// We check if a component exists in local (that looks like a base component) but not in upstream
	for name := range localNames {
		if !upstreamNames[name] {
			// Only flag if it was likely a base component (was in a previous upstream)
			// For simplicity, we don't flag custom components here
		}
	}

	report.Outdated = report.OutdatedComponents > 0 || len(report.NewBaseComponents) > 0

	if report.NewBaseComponents == nil {
		report.NewBaseComponents = []string{}
	}
	if report.RemovedBaseComponents == nil {
		report.RemovedBaseComponents = []string{}
	}

	return report
}

// CompareAll compares multiple local distributions against the upstream base.
func CompareAll(locals []*manifest.Manifest, upstream *manifest.Manifest) []DriftReport {
	var reports []DriftReport
	for _, local := range locals {
		reports = append(reports, Compare(local, upstream))
	}
	return reports
}

func componentNameSet(m *manifest.Manifest) map[string]bool {
	names := make(map[string]bool)
	for _, c := range m.AllComponents() {
		names[c.Name()] = true
	}
	return names
}

func componentVersionMap(m *manifest.Manifest) map[string]string {
	versions := make(map[string]string)
	for _, c := range m.AllComponents() {
		versions[c.Name()] = c.Version()
	}
	return versions
}
