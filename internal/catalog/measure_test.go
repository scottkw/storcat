package catalog

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// buildMeasureFixture creates a small tree (two regular files under one
// subdirectory, plus a hidden file at the root) and returns its root.
func buildMeasureFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	write := func(rel, data string) {
		p := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(p, []byte(data), 0644); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}
	write("a.txt", "hello")
	write("sub/b.txt", "world!")
	write(".hidden", "secret")
	return dir
}

// TestMeasureTree_CountsFilesAndBytes verifies MeasureTree's count-only
// pass agrees exactly with the real walk (traverseDirectory) over the
// same tree and the same options.
func TestMeasureTree_CountsFilesAndBytes(t *testing.T) {
	dir := buildMeasureFixture(t)
	svc := NewService()

	tree, err := svc.traverseDirectory(context.Background(), dir, dir, &walkState{opts: Options{IncludeHidden: false}})
	if err != nil {
		t.Fatalf("traverseDirectory: %v", err)
	}
	wantFiles := svc.countFiles(tree)
	wantBytes := tree.Size

	files, bytes, err := MeasureTree(context.Background(), dir, Options{IncludeHidden: false}, nil)
	if err != nil {
		t.Fatalf("MeasureTree: %v", err)
	}
	if files != wantFiles {
		t.Errorf("files = %d, want %d (matching the real walk)", files, wantFiles)
	}
	if bytes != wantBytes {
		t.Errorf("bytes = %d, want %d (matching the real walk)", bytes, wantBytes)
	}
}

// TestMeasureTree_RespectsIncludeHidden verifies the hidden-file rule
// matches the walk's rule exactly: excluded when IncludeHidden is false,
// counted when true.
func TestMeasureTree_RespectsIncludeHidden(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".hidden"), []byte("12345"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	files, bytes, err := MeasureTree(context.Background(), dir, Options{IncludeHidden: false}, nil)
	if err != nil {
		t.Fatalf("MeasureTree: %v", err)
	}
	if files != 0 || bytes != 0 {
		t.Errorf("with IncludeHidden=false: files=%d bytes=%d, want 0,0", files, bytes)
	}

	files, bytes, err = MeasureTree(context.Background(), dir, Options{IncludeHidden: true}, nil)
	if err != nil {
		t.Fatalf("MeasureTree: %v", err)
	}
	if files != 1 || bytes != 5 {
		t.Errorf("with IncludeHidden=true: files=%d bytes=%d, want 1,5", files, bytes)
	}
}

// TestMeasureTree_HonoursCancellation verifies a cancelled context aborts
// the measurement and returns the cancellation error.
func TestMeasureTree_HonoursCancellation(t *testing.T) {
	dir := buildMeasureFixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _, err := MeasureTree(ctx, dir, Options{}, nil)
	if err == nil {
		t.Fatal("MeasureTree with a cancelled context returned a nil error")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("error = %v, want context.Canceled", err)
	}
}

// TestMeasureTree_TolerantOfUnreadableEntries verifies an unreadable
// subdirectory is skipped without failing the measurement, matching the
// walk's single-entry tolerance -- files outside the locked subdirectory
// are still counted.
func TestMeasureTree_TolerantOfUnreadableEntries(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root bypasses permission bits")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("hello"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	locked := filepath.Join(dir, "locked")
	if err := os.Mkdir(locked, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(locked, "inside.txt"), []byte("secret"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := os.Chmod(locked, 0311); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(locked, 0755) })

	files, bytes, err := MeasureTree(context.Background(), dir, Options{}, nil)
	if err != nil {
		t.Fatalf("MeasureTree: %v", err)
	}
	if files != 1 || bytes != 5 {
		t.Errorf("files=%d bytes=%d, want 1,5 (locked subdir's contents skipped, not fatal)", files, bytes)
	}
}
