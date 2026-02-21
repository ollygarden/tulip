package upstream

import (
	"testing"

	"github.com/ollygarden/tulip/cmd/tulip/internal/manifest"
)

func TestUpgradeUpdatesBaseVersions(t *testing.T) {
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

	Upgrade(local, upstream)

	for _, c := range local.Receivers {
		if c.Version() != "v0.148.0" {
			t.Errorf("Component %s version = %q, want %q", c.Name(), c.Version(), "v0.148.0")
		}
	}
}

func TestUpgradeUpdatesCustomVersions(t *testing.T) {
	local := &manifest.Manifest{
		Dist: manifest.Dist{Name: "test", Version: "0.145.0"},
		Receivers: []manifest.Component{
			{GoMod: "go.opentelemetry.io/collector/receiver/otlpreceiver v0.145.0"},
			{GoMod: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kafkareceiver v0.145.0"},
		},
	}
	upstream := &manifest.Manifest{
		Dist: manifest.Dist{Name: "tulip", Version: "0.148.0"},
		Receivers: []manifest.Component{
			{GoMod: "go.opentelemetry.io/collector/receiver/otlpreceiver v0.148.0"},
		},
	}

	Upgrade(local, upstream)

	kafka := local.Receivers[1]
	if kafka.Version() != "v0.148.0" {
		t.Errorf("kafkareceiver version = %q, want %q", kafka.Version(), "v0.148.0")
	}
}

func TestUpgradeUpdatesDistVersion(t *testing.T) {
	local := &manifest.Manifest{
		Dist: manifest.Dist{Name: "test", Version: "0.145.0"},
	}
	upstream := &manifest.Manifest{
		Dist: manifest.Dist{Name: "tulip", Version: "0.148.0"},
	}

	Upgrade(local, upstream)

	if local.Dist.Version != "0.148.0" {
		t.Errorf("Dist.Version = %q, want %q", local.Dist.Version, "0.148.0")
	}
}

func TestUpgradePreservesCustomComponents(t *testing.T) {
	local := &manifest.Manifest{
		Dist: manifest.Dist{Name: "test", Version: "0.145.0"},
		Receivers: []manifest.Component{
			{GoMod: "go.opentelemetry.io/collector/receiver/otlpreceiver v0.145.0"},
			{GoMod: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kafkareceiver v0.145.0"},
		},
	}
	upstream := &manifest.Manifest{
		Dist: manifest.Dist{Name: "tulip", Version: "0.148.0"},
		Receivers: []manifest.Component{
			{GoMod: "go.opentelemetry.io/collector/receiver/otlpreceiver v0.148.0"},
		},
	}

	Upgrade(local, upstream)

	if len(local.Receivers) != 2 {
		t.Errorf("len(Receivers) = %d, want 2 (custom component should be preserved)", len(local.Receivers))
	}

	if !local.HasComponent("kafkareceiver") {
		t.Error("kafkareceiver was removed during upgrade")
	}
}

func TestUpgradeAddsNewBaseComponents(t *testing.T) {
	local := &manifest.Manifest{
		Dist: manifest.Dist{Name: "test", Version: "0.145.0"},
		Receivers: []manifest.Component{
			{GoMod: "go.opentelemetry.io/collector/receiver/otlpreceiver v0.145.0"},
		},
	}
	upstream := &manifest.Manifest{
		Dist: manifest.Dist{Name: "tulip", Version: "0.148.0"},
		Receivers: []manifest.Component{
			{GoMod: "go.opentelemetry.io/collector/receiver/otlpreceiver v0.148.0"},
			{GoMod: "go.opentelemetry.io/collector/receiver/nopreceiver v0.148.0"},
		},
	}

	Upgrade(local, upstream)

	if !local.HasComponent("nopreceiver") {
		t.Error("nopreceiver should have been added from upstream")
	}
	if len(local.Receivers) != 2 {
		t.Errorf("len(Receivers) = %d, want 2", len(local.Receivers))
	}
}
