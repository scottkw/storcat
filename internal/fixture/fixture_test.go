package fixture

import (
	"encoding/json"
	"os"
	"testing"

	"storcat-wails/pkg/models"
)

func TestWriteDCIMCatalog_DefaultShape(t *testing.T) {
	dir := t.TempDir()

	path, nodeCount, sizeBytes, err := WriteDCIMCatalog(dir, 50, 50, 16)
	if err != nil {
		t.Fatalf("WriteDCIMCatalog failed: %v", err)
	}

	if nodeCount < 40000 {
		t.Errorf("expected default DCIM shape to report at least 40000 nodes, got %d", nodeCount)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read written fixture: %v", err)
	}

	var root models.CatalogItem
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatalf("written fixture does not re-parse as models.CatalogItem: %v", err)
	}
	if root.Name != "./" {
		t.Errorf("expected root Name to be './', got %q", root.Name)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("failed to stat written fixture: %v", err)
	}
	if info.Size() != sizeBytes {
		t.Errorf("reported sizeBytes %d does not match os.Stat size %d", sizeBytes, info.Size())
	}
}

func TestWriteFlatCatalog_ExactFileCount(t *testing.T) {
	dir := t.TempDir()

	path, nodeCount, _, err := WriteFlatCatalog(dir, 250)
	if err != nil {
		t.Fatalf("WriteFlatCatalog failed: %v", err)
	}
	if nodeCount != 250 {
		t.Errorf("expected node count 250, got %d", nodeCount)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read written fixture: %v", err)
	}

	var root models.CatalogItem
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatalf("written fixture does not re-parse: %v", err)
	}
	if len(root.Contents) != 250 {
		t.Errorf("expected 250 direct root contents, got %d", len(root.Contents))
	}
}

func TestWriteDeepCatalog_ReParses(t *testing.T) {
	dir := t.TempDir()

	path, nodeCount, _, err := WriteDeepCatalog(dir, 20)
	if err != nil {
		t.Fatalf("WriteDeepCatalog failed: %v", err)
	}
	if nodeCount != 21 { // 20 nested directories + 1 leaf file
		t.Errorf("expected node count 21, got %d", nodeCount)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read written fixture: %v", err)
	}

	var root models.CatalogItem
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatalf("written fixture does not re-parse: %v", err)
	}
}
