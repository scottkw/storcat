package config

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestCountsCache_HitOnUnchangedKey(t *testing.T) {
	dir := t.TempDir()
	c, err := NewCountsCacheAt(filepath.Join(dir, "counts-cache.json"))
	if err != nil {
		t.Fatalf("NewCountsCacheAt: %v", err)
	}

	mt := time.Now()
	key := CountsKey("/catalogs/a.json", mt, 1234)
	if err := c.Put(key, CountEntry{FileCount: 10, TotalBytes: 2048}); err != nil {
		t.Fatalf("Put: %v", err)
	}

	entry, ok := c.Get(key)
	if !ok {
		t.Fatal("expected hit, got miss")
	}
	if entry.FileCount != 10 || entry.TotalBytes != 2048 {
		t.Errorf("got %+v, want FileCount=10 TotalBytes=2048", entry)
	}
}

func TestCountsCache_MissOnChangedModTime(t *testing.T) {
	dir := t.TempDir()
	c, err := NewCountsCacheAt(filepath.Join(dir, "counts-cache.json"))
	if err != nil {
		t.Fatalf("NewCountsCacheAt: %v", err)
	}

	mt := time.Now()
	key := CountsKey("/catalogs/a.json", mt, 1234)
	if err := c.Put(key, CountEntry{FileCount: 10, TotalBytes: 2048}); err != nil {
		t.Fatalf("Put: %v", err)
	}

	otherKey := CountsKey("/catalogs/a.json", mt.Add(time.Second), 1234)
	if _, ok := c.Get(otherKey); ok {
		t.Error("expected miss for changed modtime, got hit")
	}
}

func TestCountsCache_MissOnChangedSize(t *testing.T) {
	dir := t.TempDir()
	c, err := NewCountsCacheAt(filepath.Join(dir, "counts-cache.json"))
	if err != nil {
		t.Fatalf("NewCountsCacheAt: %v", err)
	}

	mt := time.Now()
	key := CountsKey("/catalogs/a.json", mt, 1234)
	if err := c.Put(key, CountEntry{FileCount: 10, TotalBytes: 2048}); err != nil {
		t.Fatalf("Put: %v", err)
	}

	otherKey := CountsKey("/catalogs/a.json", mt, 9999)
	if _, ok := c.Get(otherKey); ok {
		t.Error("expected miss for changed size, got hit")
	}
}

func TestCountsCache_InvalidJSON_DegradesToAllMiss(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "counts-cache.json")
	if err := os.WriteFile(path, []byte("{not valid json"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	c, err := NewCountsCacheAt(path)
	if err != nil {
		t.Fatalf("NewCountsCacheAt returned error for corrupt file: %v", err)
	}

	if _, ok := c.Get(CountsKey("/anything", time.Now(), 0)); ok {
		t.Error("expected miss against corrupt cache, got hit")
	}
}

func TestCountsCache_MissingFile_LoadsWithoutError(t *testing.T) {
	dir := t.TempDir()
	c, err := NewCountsCacheAt(filepath.Join(dir, "does-not-exist.json"))
	if err != nil {
		t.Fatalf("NewCountsCacheAt returned error for missing file: %v", err)
	}
	if _, ok := c.Get(CountsKey("/anything", time.Now(), 0)); ok {
		t.Error("expected miss against empty cache, got hit")
	}
}

func TestCountsCache_PersistsAcrossFreshInstance(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "counts-cache.json")

	c1, err := NewCountsCacheAt(path)
	if err != nil {
		t.Fatalf("NewCountsCacheAt: %v", err)
	}
	key := CountsKey("/catalogs/b.json", time.Now(), 555)
	if err := c1.Put(key, CountEntry{FileCount: 3, TotalBytes: 999}); err != nil {
		t.Fatalf("Put: %v", err)
	}

	c2, err := NewCountsCacheAt(path)
	if err != nil {
		t.Fatalf("NewCountsCacheAt (fresh): %v", err)
	}
	entry, ok := c2.Get(key)
	if !ok {
		t.Fatal("expected hit on fresh instance pointed at same file")
	}
	if entry.FileCount != 3 || entry.TotalBytes != 999 {
		t.Errorf("got %+v, want FileCount=3 TotalBytes=999", entry)
	}
}

func TestCountsCache_AtomicWrite_NoLeftoverTempFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "counts-cache.json")
	c, err := NewCountsCacheAt(path)
	if err != nil {
		t.Fatalf("NewCountsCacheAt: %v", err)
	}

	if err := c.Put(CountsKey("/x", time.Now(), 1), CountEntry{FileCount: 1, TotalBytes: 1}); err != nil {
		t.Fatalf("Put: %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected exactly 1 file in cache dir, got %d: %v", len(entries), entries)
	}
	if entries[0].Name() != "counts-cache.json" {
		t.Errorf("expected only counts-cache.json, got %q", entries[0].Name())
	}
}

func TestCountsCache_ConcurrentPutGet_NoRace(t *testing.T) {
	dir := t.TempDir()
	c, err := NewCountsCacheAt(filepath.Join(dir, "counts-cache.json"))
	if err != nil {
		t.Fatalf("NewCountsCacheAt: %v", err)
	}

	const n = 60
	sharedKey := CountsKey("/shared", time.Now(), 0)
	distinctKeys := make([]string, n)
	for i := range distinctKeys {
		distinctKeys[i] = CountsKey("/catalog", time.Now(), int64(i))
	}

	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			c.Put(distinctKeys[i], CountEntry{FileCount: i, TotalBytes: int64(i * 100)})
			c.Put(sharedKey, CountEntry{FileCount: i, TotalBytes: int64(i)})
			c.Get(distinctKeys[i])
			c.Get(sharedKey)
		}(i)
	}
	wg.Wait()

	for i, key := range distinctKeys {
		entry, ok := c.Get(key)
		if !ok {
			t.Errorf("distinct key %d: expected hit after concurrent writes complete, got miss", i)
			continue
		}
		if entry.FileCount != i {
			t.Errorf("distinct key %d: FileCount = %d, want %d", i, entry.FileCount, i)
		}
	}
	if _, ok := c.Get(sharedKey); !ok {
		t.Error("expected shared key present after concurrent contended writes")
	}
}
