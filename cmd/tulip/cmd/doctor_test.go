package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ollygarden/tulip/cmd/tulip/internal/manifest"
	"github.com/ollygarden/tulip/cmd/tulip/internal/upstream"
)

// TestDoctorOutputSchema tests that the DoctorOutput struct matches the expected JSON schema.
func TestDoctorOutputSchema(t *testing.T) {
	output := DoctorOutput{
		UpstreamVersion: "0.148.0",
		Outdated:        true,
		Distributions: []upstream.DriftReport{
			{
				Name:                  "tulip",
				LocalVersion:          "v0.145.0",
				UpstreamVersion:       "v0.148.0",
				Outdated:              true,
				OutdatedComponents:    18,
				NewBaseComponents:     []string{"memorylimiterprocessor"},
				RemovedBaseComponents: []string{},
			},
		},
	}

	data, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	// Verify it round-trips
	var parsed DoctorOutput
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if parsed.UpstreamVersion != "0.148.0" {
		t.Errorf("UpstreamVersion = %q, want %q", parsed.UpstreamVersion, "0.148.0")
	}
	if !parsed.Outdated {
		t.Error("Outdated = false, want true")
	}
	if len(parsed.Distributions) != 1 {
		t.Fatalf("len(Distributions) = %d, want 1", len(parsed.Distributions))
	}
	if parsed.Distributions[0].OutdatedComponents != 18 {
		t.Errorf("OutdatedComponents = %d, want 18", parsed.Distributions[0].OutdatedComponents)
	}
}

// TestDoctorCompareLogic tests the doctor comparison logic without network calls.
func TestDoctorCompareLogic(t *testing.T) {
	local := &manifest.Manifest{
		Dist: manifest.Dist{Name: "test", Version: "0.145.0"},
		Receivers: []manifest.Component{
			{GoMod: "go.opentelemetry.io/collector/receiver/otlpreceiver v0.145.0"},
		},
	}
	upstreamM := &manifest.Manifest{
		Dist: manifest.Dist{Name: "tulip", Version: "0.148.0"},
		Receivers: []manifest.Component{
			{GoMod: "go.opentelemetry.io/collector/receiver/otlpreceiver v0.148.0"},
		},
	}

	reports := upstream.CompareAll([]*manifest.Manifest{local}, upstreamM)

	output := DoctorOutput{
		UpstreamVersion: upstreamM.OTelVersion(),
		Distributions:   reports,
	}

	anyOutdated := false
	for _, r := range reports {
		if r.Outdated {
			anyOutdated = true
		}
	}
	output.Outdated = anyOutdated

	if !output.Outdated {
		t.Error("Doctor output should be outdated")
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	// Verify valid JSON
	var check map[string]any
	if err := json.Unmarshal(data, &check); err != nil {
		t.Errorf("Output is not valid JSON: %v", err)
	}
}

// TestDoctorReportUpToDate tests that an up-to-date distribution reports correctly.
func TestDoctorReportUpToDate(t *testing.T) {
	dir := t.TempDir()

	tulipDir := filepath.Join(dir, "distributions", "tulip")
	os.MkdirAll(tulipDir, 0755)
	m := &manifest.Manifest{
		Dist: manifest.Dist{Name: "tulip", Version: "0.148.0"},
		Receivers: []manifest.Component{
			{GoMod: "go.opentelemetry.io/collector/receiver/otlpreceiver v0.148.0"},
		},
	}
	manifest.Write(filepath.Join(tulipDir, "manifest.yaml"), m)

	upstreamM := &manifest.Manifest{
		Dist: manifest.Dist{Name: "tulip", Version: "0.148.0"},
		Receivers: []manifest.Component{
			{GoMod: "go.opentelemetry.io/collector/receiver/otlpreceiver v0.148.0"},
		},
	}

	reports := upstream.CompareAll([]*manifest.Manifest{m}, upstreamM)

	for _, r := range reports {
		if r.Outdated {
			t.Error("Up-to-date distribution reported as outdated")
		}
	}
}
