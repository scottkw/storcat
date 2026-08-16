package catalog

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"storcat-wails/pkg/models"
)

// ResolveMode is the write mode a re-scan's diff is resolved with. Discard
// is deliberately NOT a value here -- it has no Go call at all (the dialog
// simply closes and nothing is written), so a "discard mode" would be a
// write-path binding surface for an action that writes nothing.
type ResolveMode string

const (
	ResolveOverwrite ResolveMode = "overwrite"
	ResolveKeepBoth  ResolveMode = "keep-both"
)

// WriteRescanResult writes tree -- the tree a completed re-scan already
// walked and diffed -- to disk according to mode, through the exact same
// hardened, crash-safe write path Create uses (WriteCatalogFrom). There is
// no second write primitive here.
//
// jsonPath is the ORIGINAL catalog's own on-disk .json path; dir/root are
// derived from it exactly the way DuplicateCatalog derives them
// (duplicate.go:33-34). For overwrite, root is used as-is. For keep-both,
// nextCopyRoot (duplicate.go, unmodified) resolves the target instead --
// only the collision-resolution half of Duplicate is reused; DuplicateCatalog
// itself is not called, since it copies bytes from an existing JSON and a
// re-scan has a freshly walked tree to write instead.
//
// The .html sibling is written only when one already exists beside the
// ORIGINAL .json -- checked here from disk state, not trusted from opts or
// any caller -- so both modes rewrite an .html that was already there and
// never create one where none existed.
func (s *Service) WriteRescanResult(tree *models.CatalogItem, title, jsonPath string, mode ResolveMode, opts Options) (*models.CreateCatalogResult, error) {
	if filepath.Ext(jsonPath) != ".json" {
		return nil, fmt.Errorf("write re-scan result %s: not a catalog JSON file", jsonPath)
	}

	dir := filepath.Dir(jsonPath)
	root := strings.TrimSuffix(filepath.Base(jsonPath), ".json")

	outputRoot := root
	if mode == ResolveKeepBoth {
		newRoot, err := nextCopyRoot(dir, root)
		if err != nil {
			return nil, fmt.Errorf("write re-scan result %s: %w", jsonPath, err)
		}
		outputRoot = newRoot
	}

	writeOpts := opts
	if _, err := os.Stat(strings.TrimSuffix(jsonPath, ".json") + ".html"); err == nil {
		writeOpts.WriteHTML = true
	} else {
		writeOpts.WriteHTML = false
	}

	result, err := s.WriteCatalogFrom(tree, title, dir, outputRoot, "", writeOpts)
	if err != nil {
		return nil, fmt.Errorf("write re-scan result %s: %w", jsonPath, err)
	}
	return result, nil
}
