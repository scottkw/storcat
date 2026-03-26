package cli

import (
	"fmt"
	"os"
)

// Run is the CLI entry point. args is os.Args[1:] (subcommand is args[0]).
// Returns exit code: 0 = success, 1 = runtime error, 2 = usage error.
func Run(args []string, version string) int {
	if len(args) == 0 {
		printUsage(version)
		return 0
	}

	switch args[0] {
	case "version":
		return runVersion(args[1:], version)
	case "create":
		return runCreate(args[1:])
	case "search":
		return runSearch(args[1:])
	case "list":
		return runList(args[1:])
	case "show":
		return runShow(args[1:])
	case "open":
		return runOpen(args[1:])
	case "help", "--help", "-h":
		printUsage(version)
		return 0
	default:
		fmt.Fprintf(os.Stderr, "storcat: unknown command %q\nRun 'storcat --help' for usage.\n", args[0])
		return 2
	}
}

func printUsage(version string) {
	fmt.Fprintf(os.Stdout, "storcat %s\n\nUsage:\n  storcat [command]\n\nCommands:\n  create    Create a catalog from a directory\n  search    Search catalogs for a term\n  list      List catalogs in a directory\n  show      Display a catalog's tree structure\n  open      Open a catalog's HTML in the default browser\n  version   Print the version\n\nFlags:\n  -h, --help   Show help for a command\n\nRun 'storcat <command> --help' for command-specific help.\n", version)
}
