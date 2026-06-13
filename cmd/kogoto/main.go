package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tooppoo/Kogoto/internal/agent/fake"
	"github.com/tooppoo/Kogoto/internal/config"
	"github.com/tooppoo/Kogoto/internal/runner"
	github "github.com/tooppoo/Kogoto/internal/tracker/github"
)

const version = "0.0.0-dev"

func main() {
	root := &cobra.Command{
		Use:     "kogoto",
		Short:   "Human-in-the-loop-first AI runner for issue-driven development",
		Version: version,
	}

	root.AddCommand(newResolveCmd())
	root.AddCommand(newStatusCmd())

	for _, p := range []struct{ use, short string }{
		{"setup", "Configure Kogoto for a repository"},
		{"writer", "Configure the writer agent"},
		{"reviewer", "Configure the reviewer agent"},
		{"answer", "Provide a human answer to a blocked question"},
		{"resume", "Resume a paused run"},
		{"abort", "Abort the current run"},
		{"close", "Close a completed run"},
	} {
		use, short := p.use, p.short
		root.AddCommand(&cobra.Command{
			Use:   use,
			Short: short,
			RunE: func(cmd *cobra.Command, args []string) error {
				return fmt.Errorf("%s: not yet implemented", cmd.Use)
			},
		})
	}

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func newResolveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "resolve <issue-number>",
		Short: "Start a Kogoto run for a GitHub Issue",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			issueNumber, err := strconv.Atoi(args[0])
			if err != nil || issueNumber < 1 {
				return fmt.Errorf("invalid issue number: %s", args[0])
			}

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			t, err := github.New(cfg.GitHub.Owner, cfg.GitHub.Repo, cfg.GitHub.Token, cfg.GitHub.Host)
			if err != nil {
				return err
			}
			w := &fake.Writer{}

			root, err := gitRepoRoot()
			if err != nil {
				return err
			}

			r := runner.New(root, cfg, t, w)
			result, err := r.Resolve(cmd.Context(), issueNumber)
			if err != nil {
				return err
			}

			fmt.Printf("Issue #%d is now blocked waiting for human input.\n", issueNumber)
			fmt.Printf("Run ID: %s\n", result.RunID)
			fmt.Printf("Use `kogoto status %d` to view the blocked state.\n", issueNumber)
			return nil
		},
	}
}

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status [issue-number]",
		Short: "Show the status of the current run",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadBase()
			if err != nil {
				return err
			}
			root, err := gitRepoRoot()
			if err != nil {
				return err
			}
			r := runner.New(root, cfg, nil, nil)
			if len(args) == 0 {
				return r.StatusActive()
			}
			issueNumber, err := strconv.Atoi(args[0])
			if err != nil || issueNumber < 1 {
				return fmt.Errorf("invalid issue number: %s", args[0])
			}
			return r.Status(issueNumber)
		},
	}
}

func gitRepoRoot() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("resolve git repository root: %w", err)
	}
	return strings.TrimRight(string(out), "\n"), nil
}
