package cli

import (
	"flag"
	"fmt"
	"os"
)



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
