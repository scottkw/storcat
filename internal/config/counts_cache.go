package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

// CountEntry is one sidecar cache entry: a catalog's file count and total
// byte size, as last observed by a full walk of its contents.
type CountEntry struct {
	FileCount  int   `json:"fileCount"`
	TotalBytes int64 `json:"totalBytes"`
}

// CountsCache is a mutex-guarded, disk-persisted cache of per-catalog file
// counts and byte totals, keyed on path+modtime+size so a changed catalog
// misses rather than serving a stale count.
//
// Unlike config.Manager (its nearest analog), this cache's load-mutate-save
// cycle is guarded by a mutex: a background fill and an opportunistic fill
// from LoadCatalogFlat can run concurrently while BrowseCatalogs reads it,
// and Wails does not guarantee bound methods are invoked serially from the
// frontend.
type CountsCache struct {
	mu      sync.Mutex
	path    string
	entries map[string]CountEntry
}

// CountsKey builds the cache key for a catalog file from its path,
// modification time and size -- a concatenation rather than a hash, which
// is cheaper to build and collision-safe at this cardinality (tens of
// catalogs).
func CountsKey(path string, modTime time.Time, size int64) string {
	return path + "|" + modTime.UTC().Format(time.RFC3339Nano) + "|" + strconv.FormatInt(size, 10)
}

// NewCountsCache resolves the shared storcat configuration directory,
// points the cache at counts-cache.json beside config.json, and loads any
// existing entries.
func NewCountsCache() (*CountsCache, error) {
	dir, err := storcatConfigDir()
	if err != nil {
		return nil, err
	}
	return NewCountsCacheAt(filepath.Join(dir, "counts-cache.json"))
}

// NewCountsCacheAt constructs a CountsCache pointed at an explicit file
// path instead of the real user configuration directory -- the seam tests
// (in this package and in internal/search) use to avoid ever touching the
// user's actual config dir.
func NewCountsCacheAt(path string) (*CountsCache, error) {
	c := &CountsCache{path: path, entries: make(map[string]CountEntry)}
	if err := c.Load(); err != nil {
		return nil, err
	}
	return c, nil
}

// Load reads cache entries from disk. A missing file, an unreadable file or
// invalid JSON all leave the cache empty and return no error -- a cache
// problem must never become a rail problem.
func (c *CountsCache) Load() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := os.ReadFile(c.path)
	if err != nil {
		c.entries = make(map[string]CountEntry)
		return nil
	}

	var entries map[string]CountEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		c.entries = make(map[string]CountEntry)
		return nil
	}
	c.entries = entries
	return nil
}

// Get returns the cached entry for key and whether it was found.
func (c *CountsCache) Get(key string) (CountEntry, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.entries[key]
	return entry, ok
}

// Put stores entry under key and persists the cache to disk.
func (c *CountsCache) Put(key string, entry CountEntry) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = entry
	return c.save()
}

// save marshals and atomically writes the cache file. Caller must hold mu.
// The write goes to a temporary file in the same directory, then is
// renamed into place, so a concurrent reader can never observe a
// half-written file.
func (c *CountsCache) save() error {
	data, err := json.MarshalIndent(c.entries, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(c.path)
	tmp, err := os.CreateTemp(dir, "counts-cache-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}

	if err := os.Rename(tmpPath, c.path); err != nil {
		os.Remove(tmpPath)
		return err
	}
	return nil
}
