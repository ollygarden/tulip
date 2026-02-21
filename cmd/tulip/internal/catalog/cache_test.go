package catalog

import (
	"path/filepath"
	"testing"
	"time"
)

func TestWriteAndReadCache(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "catalog.json")

	original := &Catalog{
		Version: "0.145.0",
		Entries: []Entry{
			{Name: "kafkareceiver", Type: "receiver", GoMod: "test v0.145.0"},
			{Name: "kafkaexporter", Type: "exporter", GoMod: "test v0.145.0"},
		},
	}

	if err := WriteCacheTo(path, original); err != nil {
		t.Fatalf("WriteCacheTo() error = %v", err)
	}

	cached, err := ReadCacheFrom(path)
	if err != nil {
		t.Fatalf("ReadCacheFrom() error = %v", err)
	}

	if cached.Catalog.Version != original.Version {
		t.Errorf("Version = %q, want %q", cached.Catalog.Version, original.Version)
	}
	if len(cached.Catalog.Entries) != len(original.Entries) {
		t.Errorf("len(Entries) = %d, want %d", len(cached.Catalog.Entries), len(original.Entries))
	}
}

func TestCacheExpiry(t *testing.T) {
	// Fresh cache should not be expired
	fresh := &CachedCatalog{
		FetchedAt: time.Now(),
		Catalog:   &Catalog{},
	}
	if fresh.IsExpired() {
		t.Error("Fresh cache should not be expired")
	}

	// Old cache should be expired
	old := &CachedCatalog{
		FetchedAt: time.Now().Add(-8 * 24 * time.Hour),
		Catalog:   &Catalog{},
	}
	if !old.IsExpired() {
		t.Error("Old cache (8 days) should be expired")
	}
}

func TestReadMissingCache(t *testing.T) {
	_, err := ReadCacheFrom("/nonexistent/path/catalog.json")
	if err == nil {
		t.Error("ReadCacheFrom() expected error for missing file")
	}
}
