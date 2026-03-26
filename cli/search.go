package cli

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
	"storcat-wails/internal/search"
	"storcat-wails/pkg/models"
)

func runSearch(args []string) int {
	fs := flag.NewFlagSet("search", flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stdout, "Usage: storcat search <term> [directory] [flags]\n\nSearch catalogs for files matching a term.\nIf no directory is specified, the current directory is used.\n\nFlags:\n  --json   Output results as JSON\n")
	}

	// Separate positional args from flags so flags can appear after positional args.
	var positional []string
	var flagArgs []string
	for _, a := range args {
		if strings.HasPrefix(a, "-") {
			flagArgs = append(flagArgs, a)
		} else {
			positional = append(positional, a)
		}
	}

	jsonFlag := fs.Bool("json", false, "Output results as JSON")

	if err := fs.Parse(flagArgs); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}

	if len(positional) < 1 {
		fmt.Fprintf(os.Stderr, "storcat search: search term required\nRun 'storcat search --help' for usage.\n")
		return 2
	}

	term := positional[0]

	dir := "."
	if len(positional) >= 2 {
		dir = positional[1]
	}

	var err error
	dir, err = filepath.Abs(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "storcat search: invalid path: %v\n", err)
		return 1
	}

	svc := search.NewService()
	results, err := svc.SearchCatalogs(term, dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "storcat search: %v\n", err)
		return 1
	}

	if *jsonFlag {
		if results == nil {
			results = []*models.SearchResult{}
		}
		return printJSON(results)
	}

	return printSearchTable(results)
}

func printSearchTable(results []*models.SearchResult) int {
	if len(results) == 0 {
		fmt.Fprintln(os.Stdout, "No results found.")
		return 0
	}

	table := tablewriter.NewTable(os.Stdout,
		tablewriter.WithRendition(tw.Rendition{
			Borders: tw.Border{
				Left:   tw.Off,
				Right:  tw.Off,
				Top:    tw.Off,
				Bottom: tw.Off,
			},
		}),
	)
	table.Header("File", "Path", "Type", "Size", "Catalog")
	for _, r := range results {
		table.Append(r.Basename, r.FullPath, r.Type, formatBytes(r.Size), r.Catalog)
	}
	table.Render()
	return 0
}
