package osutil

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// ResolveContainedFileURL validates raw -- either a bare absolute filesystem
// path or a "file://" URL -- and, if it resolves to a regular .json/.html
// file inside catalogDir, returns a canonical "file://" URL built from the
// resolved (symlink-followed) path. This closes FU-23-A for App.OpenExternal:
// that binding is reachable from any renderer JS, and its only live caller
// today opens a catalog's own HTML file, so restricting it to file:// paths
// contained in the configured catalog directory is simpler and safer than
// allow-listing URL schemes that would need to keep up with new hostile
// schemes (javascript:, data:, etc.) as they're invented.
//
// Every rejection returns ("", err) -- there is no partial-success return.
//
// ponytail: EvalSymlinks-then-Stat-then-open leaves the same TOCTOU window
// RevealInFileManager's doc comment already accepts: exploiting it requires
// local write access to the exact resolved path, which grants far stronger
// primitives than opening the wrong file. Upgrade path if revisited: open
// with O_NOFOLLOW and pass the descriptor.
func ResolveContainedFileURL(raw string, catalogDir string) (string, error) {
	if catalogDir == "" {
		return "", fmt.Errorf("open %s: no catalog directory configured", raw)
	}

	path, err := derivePath(raw)
	if err != nil {
		return "", fmt.Errorf("open %s: %w", raw, err)
	}

	if !filepath.IsAbs(path) {
		return "", fmt.Errorf("open %s: not an absolute file path", raw)
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("open %s: resolve path: %w", raw, err)
	}
	resolved, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		return "", fmt.Errorf("open %s: %w", raw, err)
	}

	info, err := os.Stat(resolved)
	if err != nil {
		return "", fmt.Errorf("open %s: %w", raw, err)
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("open %s: not a regular file", raw)
	}

	ext := strings.ToLower(filepath.Ext(resolved))
	if !allowedRevealExtensions[ext] {
		return "", fmt.Errorf("open %s: unsupported extension %q", raw, ext)
	}

	ok, err := ContainsPath(catalogDir, resolved)
	if err != nil {
		return "", fmt.Errorf("open %s: resolve catalog directory: %w", raw, err)
	}
	if !ok {
		return "", fmt.Errorf("open %s: outside configured catalog directory", raw)
	}

	// The resolved -- not the caller's -- path is what gets returned, so a
	// caller cannot re-introduce the symlink this function just followed.
	out := url.URL{Scheme: "file", Path: filepath.ToSlash(resolved)}
	return out.String(), nil
}

// derivePath extracts a filesystem path from raw. A raw value containing a
// scheme separator ("://") is parsed as a URL and must be scheme "file"
// with an empty or "localhost" host; its (percent-decoded) path is taken
// through filepath.FromSlash. Anything else is treated as a bare
// filesystem path, which is what the details panel passes today. A
// non-file scheme (http, https, javascript, data, ...) is rejected here
// rather than via a denylist, so a new hostile scheme can't slip through.
func derivePath(raw string) (string, error) {
	if !strings.Contains(raw, "://") {
		return raw, nil
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("parse URL: %w", err)
	}
	if u.Scheme != "file" {
		return "", fmt.Errorf("unsupported URL scheme %q", u.Scheme)
	}
	if u.Host != "" && u.Host != "localhost" {
		return "", fmt.Errorf("unsupported URL host %q", u.Host)
	}
	return filepath.FromSlash(u.Path), nil
}
