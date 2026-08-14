// Package fixture generates synthetic catalog JSON files for correctness and
// performance testing against realistic node counts (40,000+), without
// committing generated blobs to the repository. Every generator writes a
// catalog whose CatalogItem.Name follows the exact "./"-prefixed relative
// display path convention internal/catalog/service.go's traverseDirectory
// produces, so downstream flatten/parse code exercises real shapes.
package fixture

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"storcat-wails/pkg/models"
)

// fileBaseSize seeds file sizes so byte sums are not trivially uniform.
const fileBaseSize = 3500000

// WriteDCIMCatalog builds the nested shape 23-RESEARCH.md measured: topDirs
// top-level directories, subDirs directories inside each, filesPerDir files
// inside each of those. It writes a bare v2 object (a single marshalled root
// CatalogItem, never array-wrapped) to dir/fixture-dcim.json and returns the
// written path, the node count below the root (the root itself is not
// counted, matching the flat-array contract this phase adopts), and the
// written file's size from os.Stat.
func WriteDCIMCatalog(dir string, topDirs, subDirs, filesPerDir int) (string, int, int64, error) {
	root := &models.CatalogItem{Type: "directory", Name: "./"}
	fileOrdinal := 0

	for t := 0; t < topDirs; t++ {
		topName := fmt.Sprintf("./VOL%02d", t+1)
		topDir := &models.CatalogItem{Type: "directory", Name: topName}
		var topSize int64

		for s := 0; s < subDirs; s++ {
			subName := fmt.Sprintf("%s/%dCANON", topName, 100+s)
			subDir := &models.CatalogItem{Type: "directory", Name: subName}
			var subSize int64

			for f := 0; f < filesPerDir; f++ {
				fileOrdinal++
				size := int64(fileBaseSize + fileOrdinal)
				subDir.Contents = append(subDir.Contents, &models.CatalogItem{
					Type: "file",
					Name: fmt.Sprintf("%s/IMG_%04d.JPG", subName, f+1),
					Size: size,
				})
				subSize += size
			}

			subDir.Size = subSize
			topDir.Contents = append(topDir.Contents, subDir)
			topSize += subSize
		}

		topDir.Size = topSize
		root.Contents = append(root.Contents, topDir)
		root.Size += topSize
	}

	return writeCatalog(filepath.Join(dir, "fixture-dcim.json"), root)
}

// WriteFlatCatalog builds a root whose direct contents are `files` file
// nodes named "./FILE_000001.BIN" upward, and writes it to
// dir/fixture-flat.json. This shape exists so the virtualization gate can be
// measured with nothing expanded -- every node is already visible.
func WriteFlatCatalog(dir string, files int) (string, int, int64, error) {
	root := &models.CatalogItem{Type: "directory", Name: "./"}

	for f := 0; f < files; f++ {
		size := int64(fileBaseSize + f + 1)
		root.Contents = append(root.Contents, &models.CatalogItem{
			Type: "file",
			Name: fmt.Sprintf("./FILE_%06d.BIN", f+1),
			Size: size,
		})
		root.Size += size
	}

	return writeCatalog(filepath.Join(dir, "fixture-flat.json"), root)
}

// WriteDeepCatalog builds a single chain of `depth` nested directories with
// one file at the bottom, each level named after its ancestor path, and
// writes it to dir/fixture-deep.json. It exists to exercise the flattener's
// depth guard and long-path row rendering; a caller asking for a depth above
// the flattener's cap is how that guard gets tested.
func WriteDeepCatalog(dir string, depth int) (string, int, int64, error) {
	root := &models.CatalogItem{Type: "directory", Name: "./"}

	if depth > 0 {
		top := buildDeepChain(1, ".", depth)
		root.Contents = []*models.CatalogItem{top}
		root.Size = top.Size
	}

	return writeCatalog(filepath.Join(dir, "fixture-deep.json"), root)
}

// buildDeepChain recursively builds the directory at `level`, named after
// its ancestor path, terminating in a single file once level reaches depth.
func buildDeepChain(level int, parentPath string, depth int) *models.CatalogItem {
	myPath := fmt.Sprintf("%s/D%04d", parentPath, level)

	if level == depth {
		fileSize := int64(fileBaseSize + 1)
		leaf := &models.CatalogItem{Type: "file", Name: myPath + "/leaf.bin", Size: fileSize}
		return &models.CatalogItem{
			Type:     "directory",
			Name:     myPath,
			Size:     fileSize,
			Contents: []*models.CatalogItem{leaf},
		}
	}

	child := buildDeepChain(level+1, myPath, depth)
	return &models.CatalogItem{
		Type:     "directory",
		Name:     myPath,
		Size:     child.Size,
		Contents: []*models.CatalogItem{child},
	}
}

// writeCatalog creates dir if absent, marshals root as a bare v2 object,
// writes it to path, and returns the path, the node count below the root
// (root itself excluded), and the written file's size from os.Stat.
func writeCatalog(path string, root *models.CatalogItem) (string, int, int64, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", 0, 0, fmt.Errorf("create fixture directory: %w", err)
	}

	data, err := json.Marshal(root)
	if err != nil {
		return "", 0, 0, fmt.Errorf("marshal fixture catalog: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", 0, 0, fmt.Errorf("write fixture catalog: %w", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		return "", 0, 0, fmt.Errorf("stat fixture catalog: %w", err)
	}

	return path, countBelowRoot(root), info.Size(), nil
}

// countBelowRoot counts every node in item's subtree, excluding item itself.
func countBelowRoot(item *models.CatalogItem) int {
	n := 0
	for _, child := range item.Contents {
		n++
		n += countBelowRoot(child)
	}
	return n
}
