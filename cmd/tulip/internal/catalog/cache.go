package catalog

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	cacheDirName = ".tulip"
	cacheSubDir  = "catalog"
	cacheFile    = "catalog.json"
	// CacheTTL is how long the cache is valid.
	CacheTTL = 7 * 24 * time.Hour // 1 week
)

// CachedCatalog wraps a Catalog with a timestamp for expiry checks.
type CachedCatalog struct {
	Catalog   *Catalog  `json:"catalog"`
	FetchedAt time.Time `json:"fetched_at"`
}

// CacheDir returns the default cache directory path (~/.tulip/catalog/).
func CacheDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home directory: %w", err)
	}
	return filepath.Join(home, cacheDirName, cacheSubDir), nil
}

// ReadCache reads the catalog from the local cache.
// Returns an error if the cache doesn't exist or can't be read.
func ReadCache() (*CachedCatalog, error) {
	dir, err := CacheDir()
	if err != nil {
		return nil, err
	}
	return ReadCacheFrom(filepath.Join(dir, cacheFile))
}

// ReadCacheFrom reads the catalog from a specific cache file path.
func ReadCacheFrom(path string) (*CachedCatalog, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading cache: %w", err)
	}
	var cached CachedCatalog
	if err := json.Unmarshal(data, &cached); err != nil {
		return nil, fmt.Errorf("parsing cache: %w", err)
	}
	return &cached, nil
}

// WriteCache writes the catalog to the local cache.
func WriteCache(catalog *Catalog) error {
	dir, err := CacheDir()
	if err != nil {
		return err
	}
	return WriteCacheTo(filepath.Join(dir, cacheFile), catalog)
}

// WriteCacheTo writes the catalog to a specific cache file path.
func WriteCacheTo(path string, catalog *Catalog) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating cache directory: %w", err)
	}

	cached := CachedCatalog{
		Catalog:   catalog,
		FetchedAt: time.Now(),
	}

	data, err := json.MarshalIndent(cached, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling cache: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

// IsExpired returns true if the cache is older than CacheTTL.
func (c *CachedCatalog) IsExpired() bool {
	return time.Since(c.FetchedAt) > CacheTTL
}

// LoadOrFetch tries to load the catalog from cache. If the cache is missing
// or expired, it fetches from upstream and updates the cache.
func LoadOrFetch() (*Catalog, error) {
	cached, err := ReadCache()
	if err == nil && !cached.IsExpired() {
		return cached.Catalog, nil
	}

	catalog, err := FetchCatalog()
	if err != nil {
		return nil, err
	}

	// Best-effort cache write
	_ = WriteCache(catalog)
	return catalog, nil
}
