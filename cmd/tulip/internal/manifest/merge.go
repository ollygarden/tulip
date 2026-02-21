package manifest

import "fmt"

// MergeResult holds the result of merging a base manifest with additions/removals.
type MergeResult struct {
	Manifest *Manifest
	Added    []string
	Removed  []string
}

// Merge creates a new manifest by starting from the base and applying additions and removals.
// The dist metadata (name, module, description) comes from distMeta.
func Merge(base *Manifest, distMeta Dist, additions []Component, removals []string) (*MergeResult, error) {
	// Deep copy the base manifest
	merged := &Manifest{
		Dist:       distMeta,
		Extensions: copyComponents(base.Extensions),
		Receivers:  copyComponents(base.Receivers),
		Exporters:  copyComponents(base.Exporters),
		Processors: copyComponents(base.Processors),
		Connectors: copyComponents(base.Connectors),
		Providers:  copyComponents(base.Providers),
	}

	result := &MergeResult{Manifest: merged}

	// Apply removals first
	for _, name := range removals {
		if err := merged.RemoveComponent(name); err != nil {
			return nil, fmt.Errorf("removing component: %w", err)
		}
		result.Removed = append(result.Removed, name)
	}

	// Apply additions
	for _, c := range additions {
		if err := merged.AddComponent(c); err != nil {
			return nil, fmt.Errorf("adding component: %w", err)
		}
		result.Added = append(result.Added, c.Name())
	}

	return result, nil
}

// MergeVersions updates all component versions in the manifest to match the target version.
// For core components it uses coreVersion, for contrib it uses contribVersion.
func MergeVersions(m *Manifest, contribVersion, coreVersion string) {
	updateVersions(m.Extensions, contribVersion, coreVersion)
	updateVersions(m.Receivers, contribVersion, coreVersion)
	updateVersions(m.Exporters, contribVersion, coreVersion)
	updateVersions(m.Processors, contribVersion, coreVersion)
	updateVersions(m.Connectors, contribVersion, coreVersion)
	updateVersions(m.Providers, contribVersion, coreVersion)
}

func updateVersions(components []Component, contribVersion, coreVersion string) {
	for i, c := range components {
		modPath := c.ModPath()
		version := contribVersion
		if c.IsCore() {
			version = coreVersion
		}
		components[i] = Component{GoMod: modPath + " " + version}
	}
}

func copyComponents(src []Component) []Component {
	if src == nil {
		return nil
	}
	dst := make([]Component, len(src))
	copy(dst, src)
	return dst
}
