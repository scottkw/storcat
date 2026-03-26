package cli

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/browser"
	"storcat-wails/internal/search"
)

func runOpen(args []string) int {
	fs := flag.NewFlagSet("open", flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stdout, "Usage: storcat open <catalog.json>\n\nOpen the catalog's HTML report in the default browser.\n")
	}

	// Separate positional args from flags (interspersed flag pattern).
	var positional []string
	var flagArgs []string
	for _, a := range args {
		if strings.HasPrefix(a, "-") {
			flagArgs = append(flagArgs, a)
		} else {
			positional = append(positional, a)
		}
	}

	if err := fs.Parse(flagArgs); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}

	if len(positional) < 1 {
		fmt.Fprintf(os.Stderr, "storcat open: catalog file argument required\nUsage: storcat open <catalog.json>\n")
		return 2
	}

	if !strings.HasSuffix(positional[0], ".json") {
		fmt.Fprintf(os.Stderr, "storcat open: expected a .json catalog file\n")
		return 2
	}

	catalogPath := positional[0]

	// Validate catalog is readable and is a real catalog.
	svc := search.NewService()
	if _, err := svc.LoadCatalog(catalogPath); err != nil {
		fmt.Fprintf(os.Stderr, "storcat open: %v\n", err)
		return 1
	}

	// Derive HTML path from catalog path.
	htmlPath := strings.TrimSuffix(catalogPath, ".json") + ".html"

	// Check HTML file exists.
	if _, err := os.Stat(htmlPath); err != nil {
		fmt.Fprintf(os.Stderr, "storcat open: HTML file not found: %s\n", htmlPath)
		return 1
	}

	// Open in default browser.
	if err := browser.OpenFile(htmlPath); err != nil {
		fmt.Fprintf(os.Stderr, "storcat open: failed to open browser: %v\n", err)
		return 1
	}

	return 0
}
