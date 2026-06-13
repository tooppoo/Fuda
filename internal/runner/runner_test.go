package runner_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/tooppoo/Kogoto/internal/agent/fake"
	"github.com/tooppoo/Kogoto/internal/config"
	"github.com/tooppoo/Kogoto/internal/runner"
	"github.com/tooppoo/Kogoto/internal/state"
	"github.com/tooppoo/Kogoto/internal/tracker"
)

type mockTracker struct {
	issue       tracker.Issue
	commentID   int64
	commentBody string
	postErr     error
}

func (m *mockTracker) GetIssue(_ context.Context, _ int) (tracker.Issue, error) {
	return m.issue, nil
}

func (m *mockTracker) PostComment(_ context.Context, _ int, body string) (int64, error) {
	if m.postErr != nil {
		return 0, m.postErr
	}
	m.commentBody = body
	return m.commentID, nil
}

func testConfig() *config.Config {
	return &config.Config{
		GitHub: config.GitHubConfig{
			Host:  "github.com",
			Owner: "testowner",
			Repo:  "testrepo",
			Token: "test-token",
		},
		Workspace: config.WorkspaceConfig{
			Root:         "/tmp/kogoto-worktrees",
			BranchPrefix: "kogoto/issue-",
		},
		Review: config.ReviewConfig{
			MaxLoops: 3,
		},
	}
}

func TestResolveBlockedFlow(t *testing.T) {
	dir := t.TempDir()
	cfg := testConfig()

	mt := &mockTracker{
		issue: tracker.Issue{
			Number:    59,
			Title:     "Test Issue",
			Body:      "Test body",
			URL:       "https://github.com/testowner/testrepo/issues/59",
			UpdatedAt: time.Now().UTC(),
		},
		commentID: 12345,
	}

	r := runner.New(dir, cfg, mt, &fake.Writer{})
	result, err := r.Resolve(context.Background(), 59)
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	if result.RunID == "" {
		t.Error("RunID should not be empty")
	}

	// Verify issue-state.json.
	issueStatePath := state.IssueStatePath(dir, "github.com", "testowner", "testrepo", 59)
	issueState, err := state.ReadIssueState(issueStatePath)
	if err != nil {
		t.Fatalf("read issue-state.json: %v", err)
	}

	if issueState.IssueWorkflowState != state.IssueWorkflowStateWaitingForHuman {
		t.Errorf("IssueWorkflowState: got %q, want %q", issueState.IssueWorkflowState, state.IssueWorkflowStateWaitingForHuman)
	}
	if issueState.CurrentRunID != result.RunID {
		t.Errorf("CurrentRunID: got %q, want %q", issueState.CurrentRunID, result.RunID)
	}
	if issueState.Repository != "testowner/testrepo" {
		t.Errorf("Repository: got %q, want %q", issueState.Repository, "testowner/testrepo")
	}

	// Verify run.json.
	runStatePath := state.RunStatePath(dir, "github.com", "testowner", "testrepo", 59, result.RunID)
	runSt, err := state.ReadRunState(runStatePath)
	if err != nil {
		t.Fatalf("read run.json: %v", err)
	}

	if runSt.RunStateValue != state.RunStateBlocked {
		t.Errorf("RunStateValue: got %q, want %q", runSt.RunStateValue, state.RunStateBlocked)
	}
	if runSt.Blocked == nil {
		t.Fatal("Blocked info should not be nil")
	}
	if len(runSt.Blocked.Questions) == 0 {
		t.Error("Blocked.Questions should not be empty")
	}
	if runSt.Blocked.QuestionCommentID != 12345 {
		t.Errorf("QuestionCommentID: got %d, want 12345", runSt.Blocked.QuestionCommentID)
	}

	// Verify comment contains kogoto marker.
	if !strings.Contains(mt.commentBody, "<!-- kogoto:question") {
		t.Errorf("comment should contain kogoto marker, got:\n%s", mt.commentBody)
	}
	if !strings.Contains(mt.commentBody, result.RunID) {
		t.Errorf("comment should contain run ID %q, got:\n%s", result.RunID, mt.commentBody)
	}
}

func TestResolveCommentFailureIsRetryable(t *testing.T) {
	dir := t.TempDir()
	cfg := testConfig()

	mt := &mockTracker{
		issue: tracker.Issue{
			Number:    59,
			Title:     "Test Issue",
			Body:      "Test body",
			URL:       "https://github.com/testowner/testrepo/issues/59",
			UpdatedAt: time.Now().UTC(),
		},
		postErr: errors.New("github is down"),
	}

	r := runner.New(dir, cfg, mt, &fake.Writer{})
	if _, err := r.Resolve(context.Background(), 59); err == nil {
		t.Fatal("Resolve should fail when posting the blocked comment fails")
	}

	runStatePath := state.RunStatePath(dir, "github.com", "testowner", "testrepo", 59, mustCurrentRunID(t, dir))
	runSt, err := state.ReadRunState(runStatePath)
	if err != nil {
		t.Fatalf("read run.json: %v", err)
	}

	if runSt.RunStateValue != state.RunStateFailed {
		t.Errorf("RunStateValue: got %q, want %q", runSt.RunStateValue, state.RunStateFailed)
	}
	if runSt.LastError == nil {
		t.Fatal("LastError should not be nil")
	}
	if runSt.LastError.Code != state.ErrorCodeRunnerError {
		t.Errorf("LastError.Code: got %q, want %q", runSt.LastError.Code, state.ErrorCodeRunnerError)
	}
	// posting_comment failure is a transient runner_error and stays retryable,
	// overriding the manual_inspection_required default for runner_error.
	if runSt.LastError.Recoverability != state.RecoverabilityRetryable {
		t.Errorf("LastError.Recoverability: got %q, want %q", runSt.LastError.Recoverability, state.RecoverabilityRetryable)
	}
}

// mustCurrentRunID reads the current run id recorded in issue-state.json for
// issue 59 under the testowner/testrepo fixture.
func mustCurrentRunID(t *testing.T, dir string) string {
	t.Helper()
	issueStatePath := state.IssueStatePath(dir, "github.com", "testowner", "testrepo", 59)
	issueState, err := state.ReadIssueState(issueStatePath)
	if err != nil {
		t.Fatalf("read issue-state.json: %v", err)
	}
	return issueState.CurrentRunID
}

func TestResolveRefusesActiveRun(t *testing.T) {
	dir := t.TempDir()
	cfg := testConfig()
	mt := &mockTracker{
		issue: tracker.Issue{
			Number:    1,
			URL:       "https://github.com/testowner/testrepo/issues/1",
			UpdatedAt: time.Now().UTC(),
		},
		commentID: 1,
	}

	r := runner.New(dir, cfg, mt, &fake.Writer{})

	// First resolve succeeds.
	if _, err := r.Resolve(context.Background(), 1); err != nil {
		t.Fatalf("first Resolve failed: %v", err)
	}

	// Second resolve should fail because the run is waiting_for_human.
	_, err := r.Resolve(context.Background(), 1)
	if err == nil {
		t.Error("second Resolve should fail on waiting_for_human run")
	}
}

func TestStatusNoState(t *testing.T) {
	dir := t.TempDir()
	cfg := testConfig()

	r := runner.New(dir, cfg, nil, nil)
	if err := r.Status(999); err != nil {
		t.Errorf("Status on missing state should not error: %v", err)
	}
}

func TestStatusAfterResolve(t *testing.T) {
	dir := t.TempDir()
	cfg := testConfig()

	mt := &mockTracker{
		issue: tracker.Issue{
			Number:    7,
			Title:     "My Issue",
			URL:       "https://github.com/testowner/testrepo/issues/7",
			UpdatedAt: time.Now().UTC(),
		},
		commentID: 42,
	}

	r := runner.New(dir, cfg, mt, &fake.Writer{})
	if _, err := r.Resolve(context.Background(), 7); err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	if err := r.Status(7); err != nil {
		t.Errorf("Status failed: %v", err)
	}
}
