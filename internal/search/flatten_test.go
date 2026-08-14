package search

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"storcat-wails/internal/config"
	"storcat-wails/internal/fixture"
)

func writeFlattenTestCatalog(t *testing.T, dir, name, content string) string {
	t.Helper()
	filePath := filepath.Join(dir, name)
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test catalog: %v", err)
	}
	return filePath
}

// nestedCatalogJSON is a small catalog with a directory, a nested file, and
// an empty directory -- enough to exercise ordering, depth/parentIdx,
// Name/Path split, HasChildren, FileCount and TotalBytes together.
const nestedCatalogJSON = `{
  "type": "directory",
  "name": "./",
  "size": 300,
  "contents": [
    {
      "type": "directory",
      "name": "./sub",
      "size": 200,
      "contents": [
        {"type": "file", "name": "./sub/dir/file.txt", "size": 200, "contents": null}
      ]
    },
    {"type": "directory", "name": "./empty", "size": 0, "contents": []},
    {"type": "file", "name": "./top.txt", "size": 100, "contents": null}
  ]
}`

func TestLoadCatalogFlat_Structure(t *testing.T) {
	s := NewService()
	dir := t.TempDir()
	filePath := writeFlattenTestCatalog(t, dir, "nested.json", nestedCatalogJSON)

	flat, err := s.LoadCatalogFlat(filePath)
	if err != nil {
		t.Fatalf("LoadCatalogFlat failed: %v", err)
	}

	if len(flat.Nodes) != 4 {
		t.Fatalf("expected 4 flat nodes (root excluded), got %d", len(flat.Nodes))
	}

	// DFS pre-order: ./sub, ./sub/dir/file.txt, ./empty, ./top.txt
	wantPaths := []string{"./sub", "./sub/dir/file.txt", "./empty", "./top.txt"}
	for i, want := range wantPaths {
		if flat.Nodes[i].Path != want {
			t.Errorf("node %d: expected Path %q, got %q", i, want, flat.Nodes[i].Path)
		}
	}

	sub := flat.Nodes[0]
	if sub.Depth != 0 || sub.ParentIdx != -1 {
		t.Errorf("root's direct child ./sub: expected Depth 0, ParentIdx -1, got Depth %d, ParentIdx %d", sub.Depth, sub.ParentIdx)
	}
	if !sub.HasChildren {
		t.Error("./sub has one file child, expected HasChildren true")
	}

	nestedFile := flat.Nodes[1]
	if nestedFile.Name != "file.txt" {
		t.Errorf("expected Name (basename) 'file.txt', got %q", nestedFile.Name)
	}
	if nestedFile.Path != "./sub/dir/file.txt" {
		t.Errorf("expected Path (verbatim) './sub/dir/file.txt', got %q", nestedFile.Path)
	}
	if nestedFile.Depth != 1 || nestedFile.ParentIdx != 0 {
		t.Errorf("expected nested file Depth 1, ParentIdx 0, got Depth %d, ParentIdx %d", nestedFile.Depth, nestedFile.ParentIdx)
	}
	if nestedFile.HasChildren {
		t.Error("a file must never report HasChildren true")
	}

	empty := flat.Nodes[2]
	if empty.HasChildren {
		t.Error("an empty directory must report HasChildren false")
	}

	if flat.FileCount != 2 {
		t.Errorf("expected FileCount 2, got %d", flat.FileCount)
	}
	if flat.TotalBytes != 300 {
		t.Errorf("expected TotalBytes 300 (200+100), got %d", flat.TotalBytes)
	}
}

func TestLoadCatalogFlat_DualFormat(t *testing.T) {
	s := NewService()
	dir := t.TempDir()

	v2Path := writeFlattenTestCatalog(t, dir, "v2.json", nestedCatalogJSON)
	v1Path := writeFlattenTestCatalog(t, dir, "v1.json", "["+nestedCatalogJSON+"]")

	v2Flat, err := s.LoadCatalogFlat(v2Path)
	if err != nil {
		t.Fatalf("LoadCatalogFlat (v2 bare object) failed: %v", err)
	}
	v1Flat, err := s.LoadCatalogFlat(v1Path)
	if err != nil {
		t.Fatalf("LoadCatalogFlat (v1 array-wrapped) failed: %v", err)
	}

	if !reflect.DeepEqual(v1Flat.Nodes, v2Flat.Nodes) {
		t.Errorf("v1 array-wrapped and v2 bare-object fixtures with identical content produced different node slices:\nv1=%+v\nv2=%+v", v1Flat.Nodes, v2Flat.Nodes)
	}
}

func TestLoadCatalogFlat_EmptyRoot(t *testing.T) {
	s := NewService()
	dir := t.TempDir()
	filePath := writeFlattenTestCatalog(t, dir, "empty-root.json", `{"type":"directory","name":"./","size":0,"contents":[]}`)

	flat, err := s.LoadCatalogFlat(filePath)
	if err != nil {
		t.Fatalf("expected nil error for an empty root, got %v", err)
	}
	if len(flat.Nodes) != 0 {
		t.Errorf("expected a zero-length node slice for an empty root, got %d nodes", len(flat.Nodes))
	}
}

func TestLoadCatalogFlat_NotACatalog(t *testing.T) {
	s := NewService()
	dir := t.TempDir()
	filePath := writeFlattenTestCatalog(t, dir, "not-a-catalog.json", `{"foo":"bar"}`)

	_, err := s.LoadCatalogFlat(filePath)
	if err == nil {
		t.Fatal("expected an error for valid JSON that is not a catalog, got nil")
	}
	if !strings.Contains(err.Error(), "not a catalog") {
		t.Errorf("expected error to mention 'not a catalog', got: %v", err)
	}
}

func TestLoadCatalogFlat_DepthCap(t *testing.T) {
	s := NewService()
	dir := t.TempDir()

	filePath, _, _, err := fixture.WriteDeepCatalog(dir, 600)
	if err != nil {
		t.Fatalf("WriteDeepCatalog failed: %v", err)
	}

	_, err = s.LoadCatalogFlat(filePath)
	if err == nil {
		t.Fatal("expected an error for a catalog nested deeper than the 512-level cap, got nil")
	}
	if !strings.Contains(err.Error(), "512") {
		t.Errorf("expected error to name the 512 depth limit, got: %v", err)
	}
}

func TestLoadCatalogFlat_MissingFile(t *testing.T) {
	s := NewService()
	_, err := s.LoadCatalogFlat("/nonexistent/does-not-exist.json")
	if err == nil {
		t.Fatal("expected an error for a missing file")
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("expected a not-exist error, got: %v", err)
	}
}

func TestLoadCatalogFlat_PermissionDenied(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("skipping permission test: running as root ignores mode bits")
	}

	s := NewService()
	dir := t.TempDir()
	filePath := writeFlattenTestCatalog(t, dir, "noperm.json", `{"type":"directory","name":"./","size":0,"contents":[]}`)
	if err := os.Chmod(filePath, 0000); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() { os.Chmod(filePath, 0644) })

	_, err := s.LoadCatalogFlat(filePath)
	if err == nil {
		t.Fatal("expected an error for a permission-denied file")
	}
	if !errors.Is(err, os.ErrPermission) {
		t.Errorf("expected a permission error, got: %v", err)
	}
	if errors.Is(err, os.ErrNotExist) {
		t.Error("permission-denied error must be distinguishable from not-exist, got a not-exist error")
	}
}

func TestLoadCatalogFlat_OpportunisticCacheFill(t *testing.T) {
	s := NewService()
	dir := t.TempDir()
	filePath := writeFlattenTestCatalog(t, dir, "nested.json", nestedCatalogJSON)

	cache, err := config.NewCountsCacheAt(filepath.Join(t.TempDir(), "counts-cache.json"))
	if err != nil {
		t.Fatalf("NewCountsCacheAt: %v", err)
	}
	s.SetCountsCache(cache)

	flat, err := s.LoadCatalogFlat(filePath)
	if err != nil {
		t.Fatalf("LoadCatalogFlat: %v", err)
	}

	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	key := config.CountsKey(filePath, info.ModTime(), info.Size())

	entry, ok := cache.Get(key)
	if !ok {
		t.Fatal("expected LoadCatalogFlat to opportunistically fill the counts cache")
	}
	if entry.FileCount != flat.FileCount {
		t.Errorf("cached FileCount = %d, want %d (matching the walk's own count)", entry.FileCount, flat.FileCount)
	}
	if entry.TotalBytes != flat.TotalBytes {
		t.Errorf("cached TotalBytes = %d, want %d (matching the walk's own total)", entry.TotalBytes, flat.TotalBytes)
	}
}
