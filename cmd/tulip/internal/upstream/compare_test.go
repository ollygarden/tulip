package upstream

import (
	"testing"

	"github.com/ollygarden/tulip/cmd/tulip/internal/manifest"
)

func TestCompareSameVersions(t *testing.T) {
	local := &manifest.Manifest{
		Dist: manifest.Dist{Name: "test", Version: "0.145.0"},
		Receivers: []manifest.Component{
			{GoMod: "go.opentelemetry.io/collector/receiver/otlpreceiver v0.145.0"},
		},
	}
	upstream := &manifest.Manifest{
		Dist: manifest.Dist{Name: "tulip", Version: "0.145.0"},
		Receivers: []manifest.Component{
			{GoMod: "go.opentelemetry.io/collector/receiver/otlpreceiver v0.145.0"},
		},
	}

	report := Compare(local, upstream)

	if report.Outdated {
		t.Error("Outdated = true, want false")
	}
	if report.OutdatedComponents != 0 {
		t.Errorf("OutdatedComponents = %d, want 0", report.OutdatedComponents)
	}
}

func TestCompareOlderLocal(t *testing.T) {
	local := &manifest.Manifest{
		Dist: manifest.Dist{Name: "test", Version: "0.145.0"},
		Receivers: []manifest.Component{
			{GoMod: "go.opentelemetry.io/collector/receiver/otlpreceiver v0.145.0"},
			{GoMod: "go.opentelemetry.io/collector/receiver/nopreceiver v0.145.0"},
		},
	}
	upstream := &manifest.Manifest{
		Dist: manifest.Dist{Name: "tulip", Version: "0.148.0"},
		Receivers: []manifest.Component{
			{GoMod: "go.opentelemetry.io/collector/receiver/otlpreceiver v0.148.0"},
			{GoMod: "go.opentelemetry.io/collector/receiver/nopreceiver v0.148.0"},
		},
	}

	report := Compare(local, upstream)

	if !report.Outdated {
		t.Error("Outdated = false, want true")
	}
	if report.OutdatedComponents != 2 {
		t.Errorf("OutdatedComponents = %d, want 2", report.OutdatedComponents)
	}
	if report.LocalVersion != "v0.145.0" {
		t.Errorf("LocalVersion = %q, want %q", report.LocalVersion, "v0.145.0")
	}
	if report.UpstreamVersion != "v0.148.0" {
		t.Errorf("UpstreamVersion = %q, want %q", report.UpstreamVersion, "v0.148.0")
	}
}

func TestCompareNewBaseComponent(t *testing.T) {
	local := &manifest.Manifest{
		Dist: manifest.Dist{Name: "test", Version: "0.145.0"},
		Receivers: []manifest.Component{
			{GoMod: "go.opentelemetry.io/collector/receiver/otlpreceiver v0.145.0"},
		},
	}
	upstream := &manifest.Manifest{
		Dist: manifest.Dist{Name: "tulip", Version: "0.145.0"},
		Receivers: []manifest.Component{
			{GoMod: "go.opentelemetry.io/collector/receiver/otlpreceiver v0.145.0"},
			{GoMod: "go.opentelemetry.io/collector/receiver/nopreceiver v0.145.0"},
		},
	}

	report := Compare(local, upstream)

	if len(report.NewBaseComponents) != 1 {
		t.Fatalf("len(NewBaseComponents) = %d, want 1", len(report.NewBaseComponents))
	}
	if report.NewBaseComponents[0] != "nopreceiver" {
		t.Errorf("NewBaseComponents[0] = %q, want %q", report.NewBaseComponents[0], "nopreceiver")
	}
	if !report.Outdated {
		t.Error("Outdated = false, want true (new base component available)")
	}
}

func TestCompareCustomComponentVersionBump(t *testing.T) {
	local := &manifest.Manifest{
		Dist: manifest.Dist{Name: "test", Version: "0.145.0"},
		Receivers: []manifest.Component{
			{GoMod: "go.opentelemetry.io/collector/receiver/otlpreceiver v0.145.0"},
			// Custom component at old version
			{GoMod: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kafkareceiver v0.145.0"},
		},
	}
	upstream := &manifest.Manifest{
		Dist: manifest.Dist{Name: "tulip", Version: "0.148.0"},
		Receivers: []manifest.Component{
			{GoMod: "go.opentelemetry.io/collector/receiver/otlpreceiver v0.148.0"},
		},
	}

	report := Compare(local, upstream)

	if !report.Outdated {
		t.Error("Outdated = false, want true")
	}
	// 1 base component outdated + 1 custom component needing version bump
	if report.OutdatedComponents != 2 {
		t.Errorf("OutdatedComponents = %d, want 2", report.OutdatedComponents)
	}
}

func TestCompareAll(t *testing.T) {
	upstream := &manifest.Manifest{
		Dist: manifest.Dist{Name: "tulip", Version: "0.148.0"},
		Receivers: []manifest.Component{
			{GoMod: "go.opentelemetry.io/collector/receiver/otlpreceiver v0.148.0"},
		},
	}

	locals := []*manifest.Manifest{
		{
			Dist: manifest.Dist{Name: "dist1", Version: "0.145.0"},
			Receivers: []manifest.Component{
				{GoMod: "go.opentelemetry.io/collector/receiver/otlpreceiver v0.145.0"},
			},
		},
		{
			Dist: manifest.Dist{Name: "dist2", Version: "0.148.0"},
			Receivers: []manifest.Component{
				{GoMod: "go.opentelemetry.io/collector/receiver/otlpreceiver v0.148.0"},
			},
		},
	}

	reports := CompareAll(locals, upstream)

	if len(reports) != 2 {
		t.Fatalf("len(reports) = %d, want 2", len(reports))
	}
	if !reports[0].Outdated {
		t.Error("reports[0].Outdated = false, want true")
	}
	if reports[1].Outdated {
		t.Error("reports[1].Outdated = true, want false")
	}
}
