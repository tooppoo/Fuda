package agent

import (
	"context"
	"time"
)

type Question struct {
	ID       string
	Question string
	Blocking bool
}

type Task struct {
	ID          string
	Description string
}

type Plan struct {
	PlanningResult string
	Summary        string
	Tasks          []Task
	Questions      []Question
}

type PlanResult struct {
	Plan Plan
}

type PlanInput struct {
	RunID        string
	Repository   string
	IssueNumber  int
	IssueTitle   string
	IssueBody    string
	Branch       string
	Worktree     string
	ArtifactsDir string
	Now          time.Time
}

type Writer interface {
	Plan(ctx context.Context, input PlanInput) (PlanResult, error)
}
