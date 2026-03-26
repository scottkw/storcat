package cli

import (
	"encoding/json"
	"fmt"
	"os"

	// tablewriter imported here to ensure the dependency is registered;
	// actual table creation happens in list.go and other command files.
	_ "github.com/olekukonko/tablewriter"
)

// printJSON encodes v as indented JSON to stdout.
// Returns 0 on success, 1 on encode error.
func printJSON(v any) int {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		fmt.Fprintf(os.Stderr, "storcat: failed to encode JSON: %v\n", err)
		return 1
	}
	return 0
}

// formatBytes converts bytes to human-readable format (e.g., "271K", "3.4M").
// Standalone function in cli package to keep cli/ independent of internal/ packages.
func formatBytes(bytes int64) string {
	if bytes == 0 {
		return "0B"
	}

	const unit = 1024
	sizes := []string{"B", "K", "M", "G", "T"}

	if bytes < unit {
		return fmt.Sprintf("%dB", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit && exp < len(sizes)-1; n /= unit {
		div *= unit
		exp++
	}

	value := float64(bytes) / float64(div)

	// Format to 1 decimal place, but strip .0
	formatted := fmt.Sprintf("%.1f", value)
	if len(formatted) > 2 && formatted[len(formatted)-2:] == ".0" {
		formatted = formatted[:len(formatted)-2]
	}

	return formatted + sizes[exp+1]
}
