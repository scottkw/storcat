package catalog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strings"

	"storcat-wails/internal/osutil"
)

// RenameCatalog sets a catalog's title -- written to the JSON root's "title"
// key and, when a sibling .html exists, patched into both HTML title sites
// (<title> and <h1>). Filenames are never touched: only the title recorded
// inside the two files changes.
//
// The JSON root is rebuilt key-by-key in document order (setRootStringField
// below) rather than unmarshalled into the catalog item struct and
// re-marshalled. A full round-trip would silently drop a v1 array
// envelope's trailing report element and any key that struct doesn't
// recognize, and a raw-message-valued string-keyed Go map re-marshals in
// alphabetical key order -- both unacceptable for a v1 bash-script catalog
// this repo must keep reading byte-faithfully.
func RenameCatalog(jsonPath string, newTitle string) error {
	trimmed := strings.TrimSpace(newTitle)
	if trimmed == "" {
		return fmt.Errorf("rename %s: title is empty", jsonPath)
	}

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return fmt.Errorf("rename %s: %w", jsonPath, err)
	}

	out, err := setTitleInDocument(data, trimmed)
	if err != nil {
		return fmt.Errorf("rename %s: %w", jsonPath, err)
	}

	// Resolve, containment-check, and read the .html sibling BEFORE writing
	// the JSON. Every step here can fail (symlink escape via
	// resolveContainedSibling, a permission error on the read) -- doing them
	// first means a rejected rename leaves the JSON file completely
	// untouched instead of silently half-renamed (WR-01, 27-REVIEW.md
	// iteration 2: the previous JSON-first ordering let a rejected sibling
	// step return an error while the title had already been mutated).
	htmlPath := strings.TrimSuffix(jsonPath, ".json") + ".html"
	hasHTML := true
	resolvedHTMLPath, err := resolveContainedSibling(htmlPath, filepath.Dir(jsonPath))
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("rename %s: %w", jsonPath, err)
		}
		// Per 27-CONTEXT.md: rename is allowed on a catalog with no
		// .html -- the JSON title is written and there is simply
		// nothing to rewrite.
		hasHTML = false
	}

	var patchedHTML []byte
	if hasHTML {
		htmlData, err := os.ReadFile(resolvedHTMLPath)
		if err != nil {
			return fmt.Errorf("rename %s: %w", jsonPath, err)
		}
		patchedHTML = rewriteHTMLTitle(htmlData, trimmed)
	}

	if err := WriteFileAtomic(jsonPath, out, 0644); err != nil {
		return fmt.Errorf("rename %s: %w", jsonPath, err)
	}

	if hasHTML {
		// Residual, unavoidable-without-more-machinery gap: if the process
		// crashes or the disk fails between this write and the JSON write
		// above, the two files can end up with different titles. Both
		// individual writes are still crash-safe (WriteFileAtomic), but the
		// pair is not a single atomic transaction -- see the fix report for
		// why a full two-file atomic swap is out of scope here.
		if err := WriteFileAtomic(resolvedHTMLPath, patchedHTML, 0644); err != nil {
			return fmt.Errorf("rename %s: title updated in %s but failed to update %s: %w",
				jsonPath, filepath.Base(jsonPath), filepath.Base(resolvedHTMLPath), err)
		}
	}
	return nil
}

// resolveContainedSibling resolves siblingPath -- a filename this package
// derives itself via TrimSuffix(jsonPath, ".json") + ".html", never accepted
// from the renderer -- and confirms the resolved path still falls within
// baseDir. This mirrors osutil.TrashPaths's own belt-and-braces
// re-validation of every path it is handed (trash.go), applied here to the
// .html sibling RenameCatalog/DuplicateCatalog derive themselves rather than
// receive pre-validated from app.go's own containment gate (WR-03): without
// this, a symlink planted at the derived path inside an otherwise-trusted
// catalog directory would be read from -- and, for rename, written to --
// wherever it points.
//
// A missing sibling returns an os.IsNotExist-compatible error (surfaced by
// filepath.EvalSymlinks) so callers can keep their existing "no .html"
// branch unchanged.
func resolveContainedSibling(siblingPath, baseDir string) (string, error) {
	abs, err := filepath.Abs(siblingPath)
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", err
	}
	ok, err := osutil.ContainsPath(baseDir, resolved)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", fmt.Errorf("%s escapes catalog directory", siblingPath)
	}
	return resolved, nil
}

// setTitleInDocument locates the catalog's tree-root object within the
// on-disk envelope -- a bare object (v2) or an array whose element 0 is the
// tree root and every remaining element is carried through verbatim (v1
// bash-script/`tree -J` output) -- and returns the document with that
// root's "title" key set to value. A document that is neither shape, or
// that fails to parse, returns an error and no output.
func setTitleInDocument(doc []byte, value string) ([]byte, error) {
	trimmed := bytes.TrimLeft(doc, " \t\r\n")
	if len(trimmed) == 0 {
		return nil, fmt.Errorf("empty document")
	}

	switch trimmed[0] {
	case '[':
		var elems []json.RawMessage
		if err := json.Unmarshal(doc, &elems); err != nil {
			return nil, fmt.Errorf("invalid array envelope: %w", err)
		}
		if len(elems) == 0 {
			return nil, fmt.Errorf("empty array envelope")
		}
		root, err := setRootStringField(elems[0], "title", value)
		if err != nil {
			return nil, err
		}
		var out bytes.Buffer
		out.WriteByte('[')
		out.Write(root)
		for _, e := range elems[1:] {
			out.WriteByte(',')
			out.Write(e)
		}
		out.WriteByte(']')
		return out.Bytes(), nil
	case '{':
		return setRootStringField(doc, "title", value)
	default:
		return nil, fmt.Errorf("unrecognized JSON document shape")
	}
}

// setRootStringField rebuilds JSON object obj with key set to value,
// preserving the document order of every other key and the exact source
// bytes of every other key's value via json.RawMessage. Deliberately never
// a raw-message-valued string-keyed Go map, nor an unmarshal into the
// catalog item struct -- both would re-order or drop keys a v1 catalog
// carries that this codebase doesn't otherwise interpret.
func setRootStringField(obj []byte, key string, value string) ([]byte, error) {
	dec := json.NewDecoder(bytes.NewReader(obj))

	tok, err := dec.Token()
	if err != nil {
		return nil, fmt.Errorf("invalid JSON object: %w", err)
	}
	delim, ok := tok.(json.Delim)
	if !ok || delim != '{' {
		return nil, fmt.Errorf("expected a JSON object at the document root")
	}

	type kv struct {
		key string
		raw json.RawMessage
	}
	var pairs []kv
	found := false

	valBytes, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return nil, fmt.Errorf("invalid JSON object: %w", err)
		}
		k, ok := keyTok.(string)
		if !ok {
			return nil, fmt.Errorf("expected a string key in the JSON object")
		}

		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			return nil, fmt.Errorf("invalid value for key %q: %w", k, err)
		}

		if k == key {
			raw = valBytes
			found = true
		}
		pairs = append(pairs, kv{key: k, raw: raw})
	}

	if !found {
		pairs = append(pairs, kv{key: key, raw: valBytes})
	}

	var out bytes.Buffer
	out.WriteByte('{')
	for i, p := range pairs {
		if i > 0 {
			out.WriteByte(',')
		}
		kb, err := json.Marshal(p.key)
		if err != nil {
			return nil, err
		}
		out.Write(kb)
		out.WriteByte(':')
		out.Write(p.raw)
	}
	out.WriteByte('}')
	return out.Bytes(), nil
}

// rewriteHTMLTitle replaces the text between the first <title> and the
// following </title>, and the text between the first <h1> and the
// following </h1>, each with html.EscapeString(newTitle) -- matching the
// escaping internal/catalog/service.go's writeHTMLFile already applies at
// both sites. A tag pair that is absent is skipped, not treated as an
// error. This is a surgical substring replacement, never a full HTML
// regeneration, so the tree structure, counts and VERSION footer stay
// byte-identical.
func rewriteHTMLTitle(doc []byte, newTitle string) []byte {
	escaped := html.EscapeString(newTitle)
	out := replaceTagContent(string(doc), "<title>", "</title>", escaped)
	out = replaceTagContent(out, "<h1>", "</h1>", escaped)
	return []byte(out)
}

// replaceTagContent replaces the text between the first openTag and the
// following closeTag with replacement. If either tag is not found, doc is
// returned unchanged.
func replaceTagContent(doc, openTag, closeTag, replacement string) string {
	start := strings.Index(doc, openTag)
	if start == -1 {
		return doc
	}
	contentStart := start + len(openTag)
	relEnd := strings.Index(doc[contentStart:], closeTag)
	if relEnd == -1 {
		return doc
	}
	end := contentStart + relEnd
	return doc[:contentStart] + replacement + doc[end:]
}
