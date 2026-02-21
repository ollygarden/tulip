package manifest

import "testing"

func TestMergeAdditions(t *testing.T) {
	base := &Manifest{
		Dist: Dist{Name: "tulip", Version: "0.145.0"},
		Receivers: []Component{
			{GoMod: "go.opentelemetry.io/collector/receiver/otlpreceiver v0.145.0"},
		},
	}

	additions := []Component{
		{GoMod: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kafkareceiver v0.145.0"},
	}

	distMeta := Dist{
		Module:  "github.com/test/test",
		Name:    "test",
		Version: "0.145.0",
	}

	result, err := Merge(base, distMeta, additions, nil)
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}

	if len(result.Manifest.Receivers) != 2 {
		t.Errorf("len(Receivers) = %d, want 2", len(result.Manifest.Receivers))
	}
	if len(result.Added) != 1 {
		t.Errorf("len(Added) = %d, want 1", len(result.Added))
	}
	if result.Added[0] != "kafkareceiver" {
		t.Errorf("Added[0] = %q, want %q", result.Added[0], "kafkareceiver")
	}
}

func TestMergeRemovals(t *testing.T) {
	base := &Manifest{
		Dist: Dist{Name: "tulip", Version: "0.145.0"},
		Receivers: []Component{
			{GoMod: "go.opentelemetry.io/collector/receiver/otlpreceiver v0.145.0"},
			{GoMod: "go.opentelemetry.io/collector/receiver/nopreceiver v0.145.0"},
		},
	}

	distMeta := Dist{Name: "test", Version: "0.145.0"}
	result, err := Merge(base, distMeta, nil, []string{"nopreceiver"})
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}

	if len(result.Manifest.Receivers) != 1 {
		t.Errorf("len(Receivers) = %d, want 1", len(result.Manifest.Receivers))
	}
	if len(result.Removed) != 1 {
		t.Errorf("len(Removed) = %d, want 1", len(result.Removed))
	}
}

func TestMergeNoChanges(t *testing.T) {
	base := &Manifest{
		Dist: Dist{Name: "tulip", Version: "0.145.0"},
		Receivers: []Component{
			{GoMod: "go.opentelemetry.io/collector/receiver/otlpreceiver v0.145.0"},
		},
	}

	distMeta := Dist{Name: "test", Version: "0.145.0"}
	result, err := Merge(base, distMeta, nil, nil)
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}

	if len(result.Manifest.Receivers) != 1 {
		t.Errorf("len(Receivers) = %d, want 1", len(result.Manifest.Receivers))
	}
	if len(result.Added) != 0 {
		t.Errorf("len(Added) = %d, want 0", len(result.Added))
	}
	if len(result.Removed) != 0 {
		t.Errorf("len(Removed) = %d, want 0", len(result.Removed))
	}
}

func TestMergeDuplicateDetection(t *testing.T) {
	base := &Manifest{
		Dist: Dist{Name: "tulip", Version: "0.145.0"},
		Receivers: []Component{
			{GoMod: "go.opentelemetry.io/collector/receiver/otlpreceiver v0.145.0"},
		},
	}

	// Try to add a component that already exists in base
	additions := []Component{
		{GoMod: "go.opentelemetry.io/collector/receiver/otlpreceiver v0.145.0"},
	}

	distMeta := Dist{Name: "test", Version: "0.145.0"}
	_, err := Merge(base, distMeta, additions, nil)
	if err == nil {
		t.Error("Merge() expected error for duplicate component")
	}
}

func TestMergeVersions(t *testing.T) {
	m := &Manifest{
		Receivers: []Component{
			{GoMod: "go.opentelemetry.io/collector/receiver/otlpreceiver v0.145.0"},
			{GoMod: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kafkareceiver v0.145.0"},
		},
		Providers: []Component{
			{GoMod: "go.opentelemetry.io/collector/confmap/provider/envprovider v1.51.0"},
		},
	}

	MergeVersions(m, "v0.148.0", "v0.148.0")

	for _, c := range m.Receivers {
		if c.Version() != "v0.148.0" {
			t.Errorf("Component %s version = %q, want %q", c.Name(), c.Version(), "v0.148.0")
		}
	}
	for _, c := range m.Providers {
		if c.Version() != "v0.148.0" {
			t.Errorf("Provider %s version = %q, want %q", c.Name(), c.Version(), "v0.148.0")
		}
	}
}

func TestMergeDoesNotMutateBase(t *testing.T) {
	base := &Manifest{
		Dist: Dist{Name: "tulip", Version: "0.145.0"},
		Receivers: []Component{
			{GoMod: "go.opentelemetry.io/collector/receiver/otlpreceiver v0.145.0"},
		},
	}

	additions := []Component{
		{GoMod: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kafkareceiver v0.145.0"},
	}

	distMeta := Dist{Name: "test", Version: "0.145.0"}
	_, err := Merge(base, distMeta, additions, nil)
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}

	// Base should not be mutated
	if len(base.Receivers) != 1 {
		t.Errorf("base.Receivers mutated: len = %d, want 1", len(base.Receivers))
	}
}
