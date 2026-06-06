package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const version = "0.0.0-dev"

var placeholders = []struct{ use, short string }{
	{"setup", "Prepare a Fuda run for a GitHub Issue"},
	{"writer", "Run the writer agent on the current Fuda"},
	{"reviewer", "Run the reviewer agent on the current Fuda"},
	{"resolve", "Mark the current Fuda as resolved"},
	{"status", "Show the status of the current Fuda"},
	{"answer", "Provide a human answer to a blocked question"},
	{"resume", "Resume a paused Fuda run"},
	{"abort", "Abort the current Fuda run"},
	{"close", "Close a completed Fuda run"},
}

func main() {
	root := &cobra.Command{
		Use:     "fuda",
		Short:   "Human-controlled AI handoff for issue-driven development",
		Version: version,
	}

	for _, p := range placeholders {
		root.AddCommand(&cobra.Command{
			Use:   p.use,
			Short: p.short,
			RunE: func(cmd *cobra.Command, args []string) error {
				return fmt.Errorf("%s: not yet implemented", cmd.Use)
			},
		})
	}

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
