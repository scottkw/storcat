package cli

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"storcat-wails/internal/catalog"
)

func runCreate(args []string) int {
	fs := flag.NewFlagSet("create", flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stdout, "Usage: storcat create <directory> [flags]\n\nCreate a JSON and HTML catalog from a directory.\n\nFlags:\n  --title    Catalog title (default: directory name)\n  --name     Output filename stem (default: directory name)\n  --output   Copy catalog files to this directory\n  --json     Output result as JSON\n")
	}

	// Split at the first flag: positional args come before flags.
	// This handles: storcat create <dir> --title "Foo" --name bar --json
	var positional []string
	var flagArgs []string
	splitDone := false
	for _, a := range args {
		if !splitDone && strings.HasPrefix(a, "-") {
			splitDone = true
		}
		if splitDone {
			flagArgs = append(flagArgs, a)
		} else {
			positional = append(positional, a)
		}
	}

	title := fs.String("title", "", "Catalog title (default: directory name)")
	name := fs.String("name", "", "Output filename stem (default: directory name)")
	output := fs.String("output", "", "Output directory for copies")
	jsonFlag := fs.Bool("json", false, "Output result as JSON")

	if err := fs.Parse(flagArgs); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}

	if len(positional) < 1 {
		fmt.Fprintf(os.Stderr, "storcat create: directory argument required\nRun 'storcat create --help' for usage.\n")
		return 2
	}

	dir, err := filepath.Abs(positional[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "storcat create: invalid path: %v\n", err)
		return 1
	}

	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		fmt.Fprintf(os.Stderr, "storcat create: %s is not a directory\n", dir)
		return 1
	}

	if *title == "" {
		*title = filepath.Base(dir)
	}
	if *name == "" {
		*name = filepath.Base(dir)
	}

	outputDir := ""
	if *output != "" {
		outputDir, err = filepath.Abs(*output)
		if err != nil {
			fmt.Fprintf(os.Stderr, "storcat create: invalid output path: %v\n", err)
			return 1
		}
	}

	svc := catalog.NewService()
	result, err := svc.CreateCatalog(*title, dir, *name, outputDir, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "storcat create: %v\n", err)
		return 1
	}

	if *jsonFlag {
		return printJSON(result)
	}

	fmt.Fprintf(os.Stdout, "Created catalog: %s\n", result.JsonPath)
	fmt.Fprintf(os.Stdout, "  HTML:  %s\n", result.HtmlPath)
	fmt.Fprintf(os.Stdout, "  Files: %d\n", result.FileCount)
	fmt.Fprintf(os.Stdout, "  Size:  %s\n", formatBytes(result.TotalSize))
	if result.CopyJsonPath != "" {
		fmt.Fprintf(os.Stdout, "  Copied to: %s\n", filepath.Dir(result.CopyJsonPath))
	}
	return 0
}
