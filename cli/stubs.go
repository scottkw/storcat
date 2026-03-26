package cli

import (
	"flag"
	"fmt"
	"os"
)

func runCreate(args []string) int {
	fs := flag.NewFlagSet("create", flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: storcat create <directory> [flags]\n\nCreate a catalog from a directory.\n\nFlags:\n  --title    Catalog title\n  --name     Catalog filename\n  --output   Output directory\n  --json     Output result as JSON\n")
	}
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}
	fmt.Fprintf(os.Stderr, "storcat create: not yet implemented\n")
	return 1
}

func runSearch(args []string) int {
	fs := flag.NewFlagSet("search", flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: storcat search <term> <directory> [flags]\n\nSearch catalogs for a term.\n\nFlags:\n  --json   Output results as JSON\n")
	}
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}
	fmt.Fprintf(os.Stderr, "storcat search: not yet implemented\n")
	return 1
}

func runList(args []string) int {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: storcat list <directory> [flags]\n\nList catalogs in a directory.\n\nFlags:\n  --json   Output results as JSON\n")
	}
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}
	fmt.Fprintf(os.Stderr, "storcat list: not yet implemented\n")
	return 1
}

func runShow(args []string) int {
	fs := flag.NewFlagSet("show", flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: storcat show <catalog.json> [flags]\n\nDisplay a catalog's tree structure.\n\nFlags:\n  --depth     Limit tree depth\n  --json      Output raw catalog JSON\n  --no-color  Disable color output\n")
	}
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}
	fmt.Fprintf(os.Stderr, "storcat show: not yet implemented\n")
	return 1
}

func runOpen(args []string) int {
	fs := flag.NewFlagSet("open", flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: storcat open <catalog.json>\n\nOpen the catalog's HTML report in the default browser.\n")
	}
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}
	fmt.Fprintf(os.Stderr, "storcat open: not yet implemented\n")
	return 1
}
