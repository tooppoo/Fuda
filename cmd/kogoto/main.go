package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const version = "0.0.0-dev"

var placeholders = []struct{ use, short string }{
	{"setup", "Start a Kogoto run for a GitHub Issue"},
	{"writer", "Run the writer agent on the current run"},
	{"reviewer", "Run the reviewer agent on the current run"},
	{"resolve", "Mark the current run as resolved"},
	{"status", "Show the status of the current run"},
	{"answer", "Provide a human answer to a blocked question"},
	{"resume", "Resume a paused run"},
	{"abort", "Abort the current run"},
	{"close", "Close a completed run"},
}

func main() {
	root := &cobra.Command{
		Use:     "kogoto",
		Short:   "Human-in-the-loop-first AI runner for issue-driven development",
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
