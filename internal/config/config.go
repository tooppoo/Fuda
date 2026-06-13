package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Config struct {
	GitHub    GitHubConfig
	Workspace WorkspaceConfig
	Review    ReviewConfig
}

type GitHubConfig struct {
	Host  string
	Owner string
	Repo  string
	Token string
}

type WorkspaceConfig struct {
	Root         string
	BranchPrefix string
}

type ReviewConfig struct {
	MaxLoops int
}

// LoadBase loads config without requiring a GitHub token.
func LoadBase() (*Config, error) {
	owner, repo, err := resolveRepo()
	if err != nil {
		return nil, fmt.Errorf("cannot determine repository: %w", err)
	}

	root := os.Getenv("KOGOTO_WORKSPACE_ROOT")
	if root == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("cannot determine home directory: %w", err)
		}
		root = filepath.Join(home, "kogoto-worktrees")
	}

	return &Config{
		GitHub: GitHubConfig{
			Host:  "github.com",
			Owner: owner,
			Repo:  repo,
		},
		Workspace: WorkspaceConfig{
			Root:         root,
			BranchPrefix: "kogoto/issue-",
		},
		Review: ReviewConfig{
			MaxLoops: 3,
		},
	}, nil
}

// Load loads config and requires GITHUB_TOKEN to be set.
func Load() (*Config, error) {
	cfg, err := LoadBase()
	if err != nil {
		return nil, err
	}

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN is not set")
	}
	cfg.GitHub.Token = token

	return cfg, nil
}

func resolveRepo() (owner, repo string, err error) {
	if r := os.Getenv("KOGOTO_REPO"); r != "" {
		return parseOwnerRepo(r)
	}

	out, cmdErr := exec.Command("git", "remote", "get-url", "origin").Output()
	if cmdErr != nil {
		return "", "", fmt.Errorf("KOGOTO_REPO is not set and git remote origin not found")
	}
	return parseRemoteURL(strings.TrimSpace(string(out)))
}

func parseOwnerRepo(s string) (owner, repo string, err error) {
	parts := strings.SplitN(s, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid repository format %q, expected owner/repo", s)
	}
	return parts[0], parts[1], nil
}

func parseRemoteURL(u string) (owner, repo string, err error) {
	u = strings.TrimSuffix(u, ".git")

	if strings.HasPrefix(u, "git@") {
		// git@github.com:owner/repo
		parts := strings.SplitN(u, ":", 2)
		if len(parts) != 2 {
			return "", "", fmt.Errorf("cannot parse git remote URL: %q", u)
		}
		return parseOwnerRepo(parts[1])
	}

	// https://github.com/owner/repo
	parts := strings.Split(u, "/")
	if len(parts) < 5 {
		return "", "", fmt.Errorf("cannot parse git remote URL: %q", u)
	}
	return parts[len(parts)-2], parts[len(parts)-1], nil
}
