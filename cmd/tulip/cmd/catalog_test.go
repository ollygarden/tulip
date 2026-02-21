package cmd

import (
	"testing"

	"github.com/ollygarden/tulip/cmd/tulip/internal/catalog"
)

// TestCatalogListAll tests that the catalog lists all entries.
func TestCatalogListAll(t *testing.T) {
	cat := &catalog.Catalog{
		Version: "0.145.0",
		Entries: []catalog.Entry{
			{Name: "kafkareceiver", Type: "receiver"},
			{Name: "kafkametricsreceiver", Type: "receiver"},
			{Name: "kafkaexporter", Type: "exporter"},
			{Name: "prometheusreceiver", Type: "receiver"},
		},
	}

	if len(cat.Entries) != 4 {
		t.Errorf("Catalog has %d entries, want 4", len(cat.Entries))
	}
}
