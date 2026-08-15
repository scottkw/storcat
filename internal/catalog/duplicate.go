package catalog

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DuplicateCatalog copies jsonPath (and its sibling .html, if present) to
// the next free "-copy"/"-copy-N" filename root in the same directory,
// returning the new .json path. This is a byte copy: the title inside the
// duplicated JSON is inherited verbatim, exactly like every other byte --
// per 27-CONTEXT.md's post-research resolution, ACT-03 speaks to the
// filename, not the title, and a duplicate genuinely is an identical copy.
// A user who wants a different title renames afterwards (rename.go, 27-01).
//
// Both the .json and (when present) the .html are written through
// WriteFileAtomic -- the same crash-safe primitive service.go's copyFile
// uses -- rather than a truncate-then-stream copy: a truncating write would
// leave a partially-written file at the destination if this process is
// interrupted mid-copy.
func DuplicateCatalog(jsonPath string) (string, error) {
	if filepath.Ext(jsonPath) != ".json" {
		return "", fmt.Errorf("duplicate %s: not a catalog JSON file", jsonPath)
	}

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return "", fmt.Errorf("duplicate %s: %w", jsonPath, err)
	}

	dir := filepath.Dir(jsonPath)
	root := strings.TrimSuffix(filepath.Base(jsonPath), ".json")

	newRoot, err := nextCopyRoot(dir, root)
	if err != nil {
		return "", fmt.Errorf("duplicate %s: %w", jsonPath, err)
	}

	newJSONPath := filepath.Join(dir, newRoot+".json")
	if err := WriteFileAtomic(newJSONPath, data, 0644); err != nil {
		return "", fmt.Errorf("duplicate %s: %w", jsonPath, err)
	}

	// Derive the source HTML with the repo's one .json/.html pairing
	// convention (internal/search/service.go:214). A missing HTML is not
	// an error -- not every catalog has one.
	htmlPath := strings.TrimSuffix(jsonPath, ".json") + ".html"
	htmlData, err := os.ReadFile(htmlPath)
	if err != nil {
		if os.IsNotExist(err) {
			return newJSONPath, nil
		}
		return newJSONPath, fmt.Errorf("duplicate %s: %w", jsonPath, err)
	}

	newHTMLPath := filepath.Join(dir, newRoot+".html")
	if err := WriteFileAtomic(newHTMLPath, htmlData, 0644); err != nil {
		// The JSON copy already landed and is real -- report exactly what
		// succeeded rather than pretending nothing happened. No rollback:
		// removing a file the user can already see would be a second,
		// unasked-for destructive act.
		return newJSONPath, fmt.Errorf("duplicate %s: copy html: %w", jsonPath, err)
	}

	return newJSONPath, nil
}

// nextCopyRoot finds the first "<root>-copy", "<root>-copy-2",
// "<root>-copy-3" ... candidate for which NEITHER "<candidate>.json" NOR
// "<candidate>.html" exists in dir. Checking both extensions is what stops
// a duplicate from clobbering an orphaned .html left behind by a partial
// delete. The search is capped at 1000 candidates so a pathological
// directory cannot spin forever.
func nextCopyRoot(dir, root string) (string, error) {
	for i := 0; i < 1000; i++ {
		candidate := root + "-copy"
		if i > 0 {
			candidate = fmt.Sprintf("%s-copy-%d", root, i+1)
		}

		free, err := isCandidateRootFree(dir, candidate)
		if err != nil {
			return "", err
		}
		if free {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("no free copy root found for %q after 1000 attempts", root)
}

// isCandidateRootFree reports whether neither "<dir>/<root>.json" nor
// "<dir>/<root>.html" exists. A Stat error other than "not exists" aborts
// the search rather than being treated as free.
func isCandidateRootFree(dir, root string) (bool, error) {
	for _, ext := range []string{".json", ".html"} {
		if _, err := os.Stat(filepath.Join(dir, root+ext)); err == nil {
			return false, nil
		} else if !os.IsNotExist(err) {
			return false, fmt.Errorf("stat %s%s: %w", root, ext, err)
		}
	}
	return true, nil
}
