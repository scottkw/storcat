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

func runList(args []string) int {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	jsonFlag := fs.Bool("json", false, "Output as JSON")
	fs.Usage = func() {
		fmt.Fprintf(os.Stdout, "Usage: storcat list [directory] [flags]\n\nList catalogs in a directory.\nIf no directory is specified, the current directory is used.\n\nFlags:\n  --json   Output results as JSON\n")
	}
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}

	dir := "."
	if fs.NArg() >= 1 {
		dir = fs.Arg(0)
	}

	var err error
	dir, err = filepath.Abs(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "storcat list: invalid path: %v\n", err)
		return 1
	}

	svc := search.NewService()
	catalogs, err := svc.BrowseCatalogs(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "storcat list: %v\n", err)
		return 1
	}

	if *jsonFlag {
		if catalogs == nil {
			catalogs = []*models.CatalogMetadata{}
		}
		return printJSON(catalogs)
	}

	return printListTable(catalogs)
}

func printListTable(catalogs []*models.CatalogMetadata) int {
	if len(catalogs) == 0 {
		fmt.Fprintln(os.Stdout, "No catalogs found.")
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
	table.Header("Name", "Title", "Size", "Modified")
	for _, c := range catalogs {
		table.Append(strings.TrimSuffix(c.Filename, ".json"), c.Title, formatBytes(c.Size), c.Modified)
	}
	table.Render()
	return 0
}
