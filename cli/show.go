package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"storcat-wails/internal/search"
	"storcat-wails/pkg/models"
)

var dirColor = color.New(color.FgBlue, color.Bold)

func runShow(args []string) int {
	fs := flag.NewFlagSet("show", flag.ContinueOnError)
	jsonFlag := fs.Bool("json", false, "Output raw catalog JSON")
	depthFlag := fs.Int("depth", -1, "Limit tree depth (-1 = unlimited, 0 = root only)")
	noColorFlag := fs.Bool("no-color", false, "Disable color output")
	fs.Usage = func() {
		fmt.Fprintf(os.Stdout, "Usage: storcat show <catalog.json> [flags]\n\nDisplay a catalog's tree structure.\n\nFlags:\n  --depth N   Limit tree depth (-1 = unlimited, 0 = root only)\n  --json      Output raw catalog JSON\n  --no-color  Disable color output\n")
	}

	// Separate positional args from flags (interspersed flag pattern).
	// Flags that take values (--depth N) must have their value grouped with the flag.
	// Strategy: collect all -flag and --flag args plus their value if next arg is non-flag.
	var positional []string
	var flagArgs []string
	valueFlags := map[string]bool{"--depth": true, "-depth": true}
	for i := 0; i < len(args); i++ {
		a := args[i]
		if strings.HasPrefix(a, "-") {
			flagArgs = append(flagArgs, a)
			// If this flag takes a value and next arg exists and is non-flag, consume it
			if valueFlags[a] && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				i++
				flagArgs = append(flagArgs, args[i])
			}
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
		fmt.Fprintf(os.Stderr, "storcat show: catalog file argument required\nUsage: storcat show <catalog.json> [flags]\n")
		return 2
	}

	catalogPath := positional[0]
	if !strings.HasSuffix(catalogPath, ".json") {
		fmt.Fprintf(os.Stderr, "storcat show: expected a .json catalog file\n")
		return 2
	}

	if *noColorFlag {
		color.NoColor = true
	}

	svc := search.NewService()
	root, err := svc.LoadCatalog(catalogPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "storcat show: %v\n", err)
		return 1
	}

	if *jsonFlag {
		return printJSON(root)
	}

	// Print root node
	fmt.Fprintf(os.Stdout, "%s\n", dirColor.Sprint(filepath.Base(root.Name)))

	// Print children
	for i, child := range root.Contents {
		isLast := i == len(root.Contents)-1
		printTree(os.Stdout, child, isLast, "", 0, *depthFlag)
	}

	return 0
}

// printTree recursively prints a catalog item as a tree.
// currentDepth is 0 for direct children of root; maxDepth -1 means unlimited.
// With maxDepth=0: only root node is shown (no children printed).
// With maxDepth=1: root + immediate children shown (no grandchildren).
func printTree(w io.Writer, item *models.CatalogItem, isLast bool, prefix string, currentDepth, maxDepth int) {
	// If this depth exceeds the max, skip this item and all descendants.
	if maxDepth >= 0 && currentDepth >= maxDepth {
		return
	}

	connector := "├── "
	if isLast {
		connector = "└── "
	}

	name := filepath.Base(item.Name)
	if item.Type == "directory" {
		name = dirColor.Sprint(name)
	}

	fmt.Fprintf(w, "%s%s%s\n", prefix, connector, name)

	if item.Type == "directory" && len(item.Contents) > 0 {
		var newPrefix string
		if isLast {
			newPrefix = prefix + "    "
		} else {
			newPrefix = prefix + "│   "
		}
		for i, child := range item.Contents {
			childIsLast := i == len(item.Contents)-1
			printTree(w, child, childIsLast, newPrefix, currentDepth+1, maxDepth)
		}
	}
}
