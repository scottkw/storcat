package search

import (
	"fmt"
	"os"
	"path/filepath"

	"storcat-wails/internal/config"
	"storcat-wails/pkg/models"
)

// maxFlattenDepth guards LoadCatalogFlat's recursive walk against a
// pathologically deep catalog exhausting the goroutine stack.
const maxFlattenDepth = 512

// LoadCatalogFlat reads and parses a catalog JSON file via the unmodified
// LoadCatalog (reusing its dual-format v1/v2 parse -- this function performs
// no JSON decoding of its own), then flattens the nested tree into a single
// render-ready slice. The root itself is excluded from the returned Nodes;
// the root's direct children get Depth 0 and ParentIdx -1.
func (s *Service) LoadCatalogFlat(filePath string) (*models.FlatCatalog, error) {
	root, err := s.LoadCatalog(filePath)
	if err != nil {
		return nil, fmt.Errorf("load catalog for flatten: %w", err)
	}

	if root.Type != "file" && root.Type != "directory" {
		return nil, fmt.Errorf("%s: not a catalog (root has no recognized type)", filePath)
	}

	flat := &models.FlatCatalog{Nodes: []models.FlatNode{}}

	var walk func(item *models.CatalogItem, depth, parentIdx int) error
	walk = func(item *models.CatalogItem, depth, parentIdx int) error {
		if depth >= maxFlattenDepth {
			return fmt.Errorf("%s: exceeds maximum catalog depth of %d", item.Name, maxFlattenDepth)
		}

		idx := len(flat.Nodes)
		flat.Nodes = append(flat.Nodes, models.FlatNode{
			Name:        filepath.Base(item.Name),
			Path:        item.Name,
			Type:        item.Type,
			Size:        item.Size,
			Depth:       depth,
			ParentIdx:   parentIdx,
			HasChildren: item.Type == "directory" && len(item.Contents) > 0,
		})

		if item.Type == "file" {
			flat.FileCount++
			flat.TotalBytes += item.Size
		}

		for _, child := range item.Contents {
			if err := walk(child, depth+1, idx); err != nil {
				return err
			}
		}
		return nil
	}

	for _, child := range root.Contents {
		if err := walk(child, 0, -1); err != nil {
			return nil, err
		}
	}

	// Opportunistic cache fill: this walk already computed FileCount and
	// TotalBytes for free, so persist them under the same key BrowseCatalogs
	// looks up by. A failure to persist is ignored -- the counts are
	// convenience data, and a cache problem must never become a load
	// problem.
	if s.countsCache != nil {
		if info, statErr := os.Stat(filePath); statErr == nil {
			key := config.CountsKey(filePath, info.ModTime(), info.Size())
			_ = s.countsCache.Put(key, config.CountEntry{
				FileCount:  flat.FileCount,
				TotalBytes: flat.TotalBytes,
			})
		}
	}

	return flat, nil
}
