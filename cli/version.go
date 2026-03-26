package cli

import (
	"flag"
	"fmt"
	"os"
)

func runVersion(args []string, version string) int {
	fs := flag.NewFlagSet("version", flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stdout, "Usage: storcat version\n\nPrint the StorCat version.\n")
	}
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		fmt.Fprintf(os.Stderr, "storcat version: %v\n", err)
		return 2
	}
	fmt.Fprintf(os.Stdout, "storcat %s\n", version)
	return 0
}
