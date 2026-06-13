package fake_test

import (
	"context"
	"testing"
	"time"

	"github.com/tooppoo/Kogoto/internal/agent"
	"github.com/tooppoo/Kogoto/internal/agent/fake"
)

func TestWriterPlanReturnsBlockedByAmbiguity(t *testing.T) {
	w := &fake.Writer{}
	result, err := w.Plan(context.Background(), agent.PlanInput{
		RunID:       "run-1",
		Repository:  "owner/repo",
		IssueNumber: 1,
		IssueTitle:  "Test Issue",
		IssueBody:   "Test body",
		Branch:      "kogoto/issue-1",
		Worktree:    "/tmp/worktrees/repo-issue-1",
		Now:         time.Now().UTC(),
	})

	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if result.Plan.PlanningResult != agent.PlanningResultBlockedByAmbiguity {
		t.Errorf("PlanningResult: got %q, want %q", result.Plan.PlanningResult, agent.PlanningResultBlockedByAmbiguity)
	}
	if len(result.Plan.Questions) == 0 {
		t.Error("Questions should not be empty")
	}
	for _, q := range result.Plan.Questions {
		if !q.Blocking {
			t.Errorf("question %q should be blocking", q.ID)
		}
	}
}
