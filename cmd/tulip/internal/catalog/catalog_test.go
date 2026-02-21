package catalog

import "testing"

var testBuilderConfig = `
dist:
  version: 0.145.0
extensions:
  - gomod: github.com/open-telemetry/opentelemetry-collector-contrib/extension/healthcheckextension v0.145.0
  - gomod: github.com/open-telemetry/opentelemetry-collector-contrib/extension/pprofextension v0.145.0
receivers:
  - gomod: github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kafkareceiver v0.145.0
  - gomod: github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kafkametricsreceiver v0.145.0
  - gomod: github.com/open-telemetry/opentelemetry-collector-contrib/receiver/prometheusreceiver v0.145.0
exporters:
  - gomod: github.com/open-telemetry/opentelemetry-collector-contrib/exporter/kafkaexporter v0.145.0
  - gomod: github.com/open-telemetry/opentelemetry-collector-contrib/exporter/fileexporter v0.145.0
processors:
  - gomod: github.com/open-telemetry/opentelemetry-collector-contrib/processor/batchprocessor v0.145.0
connectors:
  - gomod: github.com/open-telemetry/opentelemetry-collector-contrib/connector/forwardconnector v0.145.0
`

func TestParseBuilderConfig(t *testing.T) {
	cat, err := ParseBuilderConfig([]byte(testBuilderConfig))
	if err != nil {
		t.Fatalf("ParseBuilderConfig() error = %v", err)
	}

	if cat.Version != "0.145.0" {
		t.Errorf("Version = %q, want %q", cat.Version, "0.145.0")
	}

	if len(cat.Entries) != 9 {
		t.Errorf("len(Entries) = %d, want 9", len(cat.Entries))
	}
}

func TestFilterByType(t *testing.T) {
	cat, _ := ParseBuilderConfig([]byte(testBuilderConfig))

	tests := []struct {
		componentType string
		wantCount     int
	}{
		{"extension", 2},
		{"receiver", 3},
		{"exporter", 2},
		{"processor", 1},
		{"connector", 1},
	}

	for _, tt := range tests {
		results := cat.FilterByType(tt.componentType)
		if len(results) != tt.wantCount {
			t.Errorf("FilterByType(%q) returned %d entries, want %d", tt.componentType, len(results), tt.wantCount)
		}
	}
}

func TestSearch(t *testing.T) {
	cat, _ := ParseBuilderConfig([]byte(testBuilderConfig))

	tests := []struct {
		query     string
		wantCount int
	}{
		{"kafka", 3}, // kafkareceiver, kafkametricsreceiver, kafkaexporter
		{"prometheus", 1},
		{"nonexistent", 0},
		{"KAFKA", 3}, // case-insensitive
	}

	for _, tt := range tests {
		results := cat.Search(tt.query)
		if len(results) != tt.wantCount {
			t.Errorf("Search(%q) returned %d results, want %d", tt.query, len(results), tt.wantCount)
		}
	}
}

func TestFind(t *testing.T) {
	cat, _ := ParseBuilderConfig([]byte(testBuilderConfig))

	entry := cat.Find("kafkareceiver")
	if entry == nil {
		t.Fatal("Find(kafkareceiver) returned nil")
	}
	if entry.Type != "receiver" {
		t.Errorf("entry.Type = %q, want %q", entry.Type, "receiver")
	}

	notFound := cat.Find("nonexistent")
	if notFound != nil {
		t.Error("Find(nonexistent) expected nil")
	}
}

func TestEntryFromGoMod(t *testing.T) {
	entry := entryFromGoMod("github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kafkareceiver v0.145.0", "receiver")

	if entry.Name != "kafkareceiver" {
		t.Errorf("Name = %q, want %q", entry.Name, "kafkareceiver")
	}
	if entry.Type != "receiver" {
		t.Errorf("Type = %q, want %q", entry.Type, "receiver")
	}
}
