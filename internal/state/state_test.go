package state_test

import (
	"testing"
	"time"

	"github.com/tooppoo/Kogoto/internal/state"
)

// Fixed UUIDs used across tests (valid UUID v4 format).
const (
	testRunID1 = "00000000-0000-4000-8000-000000000001"
	testRunID2 = "00000000-0000-4000-8000-000000000002"
)

func TestIssueStateRoundTrip(t *testing.T) {
	dir := t.TempDir()
	now := time.Now().UTC().Truncate(time.Second)

	path := state.IssueStatePath(dir, "github.com", "owner", "repo", 42)

	original := &state.IssueState{
		SchemaVersion:      1,
		Repository:         "owner/repo",
		IssueNumber:        42,
		IssueWorkflowState: state.IssueWorkflowStateWaitingForHuman,
		CurrentRunID:       testRunID1,
		Runs:               []state.RunRecord{{RunID: testRunID1, RunResult: state.RunResultActive}},
		SourceIssue: state.SourceIssue{
			URL:       "https://github.com/owner/repo/issues/42",
			UpdatedAt: now,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := state.WriteIssueState(path, original); err != nil {
		t.Fatalf("WriteIssueState: %v", err)
	}

	got, err := state.ReadIssueState(path)
	if err != nil {
		t.Fatalf("ReadIssueState: %v", err)
	}

	if got.IssueWorkflowState != state.IssueWorkflowStateWaitingForHuman {
		t.Errorf("IssueWorkflowState: got %q, want %q", got.IssueWorkflowState, state.IssueWorkflowStateWaitingForHuman)
	}
	if got.CurrentRunID != testRunID1 {
		t.Errorf("CurrentRunID: got %q, want %q", got.CurrentRunID, testRunID1)
	}
	if len(got.Runs) != 1 || got.Runs[0].RunID != testRunID1 {
		t.Errorf("Runs mismatch: %+v", got.Runs)
	}
}

func TestRunStateRoundTrip(t *testing.T) {
	dir := t.TempDir()
	now := time.Now().UTC().Truncate(time.Second)

	path := state.RunStatePath(dir, "github.com", "owner", "repo", 42, testRunID2)

	original := &state.RunState{
		SchemaVersion: 1,
		RunID:         testRunID2,
		Repository:    "owner/repo",
		IssueNumber:   42,
		RunStateValue: state.RunStateBlocked,
		Branch:        "kogoto/issue-42",
		Worktree:      "/tmp/worktrees/repo-issue-42",
		Writer:        state.Backend{BackendType: state.BackendTypeClaude},
		Reviewer:      state.Backend{BackendType: state.BackendTypeClaude},
		ReviewLoop:    state.ReviewLoop{CompletedReviewRounds: 0, MaxRounds: 3},
		Blocked: &state.BlockedInfo{
			Questions: []state.BlockedQuestion{
				{ID: "q1", Question: "What is the expected output?", Blocking: true},
			},
			QuestionCommentID: 99,
			QuestionPostedAt:  now,
			WaitingSince:      now,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := state.WriteRunState(path, original); err != nil {
		t.Fatalf("WriteRunState: %v", err)
	}

	got, err := state.ReadRunState(path)
	if err != nil {
		t.Fatalf("ReadRunState: %v", err)
	}

	if got.RunStateValue != state.RunStateBlocked {
		t.Errorf("RunStateValue: got %q, want %q", got.RunStateValue, state.RunStateBlocked)
	}
	if got.Blocked == nil {
		t.Fatal("Blocked should not be nil")
	}
	if len(got.Blocked.Questions) != 1 {
		t.Fatalf("Questions length: got %d, want 1", len(got.Blocked.Questions))
	}
	if got.Blocked.Questions[0].ID != "q1" {
		t.Errorf("Question ID: got %q, want %q", got.Blocked.Questions[0].ID, "q1")
	}
	if got.Blocked.QuestionCommentID != 99 {
		t.Errorf("QuestionCommentID: got %d, want 99", got.Blocked.QuestionCommentID)
	}
}

func TestRecoverabilityFor(t *testing.T) {
	cases := []struct {
		code state.ErrorCode
		want state.Recoverability
	}{
		{state.ErrorCodeWriterLaunchFailed, state.RecoverabilityRetryable},
		{state.ErrorCodeInvalidWriterOutput, state.RecoverabilityRetryableAfterHumanConfirmation},
		{state.ErrorCodeRunnerError, state.RecoverabilityManualInspectionRequired},
		{state.ErrorCodeIssueNotFound, state.RecoverabilityTerminal},
		{state.ErrorCodeIssueIsPR, state.RecoverabilityTerminal},
		{state.ErrorCodeVerificationFailed, state.RecoverabilityManualInspectionRequired},
	}
	for _, c := range cases {
		if got := state.RecoverabilityFor(c.code); got != c.want {
			t.Errorf("RecoverabilityFor(%q): got %q, want %q", c.code, got, c.want)
		}
	}

	// Unknown codes must not become auto-resumable.
	if got := state.RecoverabilityFor(state.ErrorCode("unknown_code")); got != state.RecoverabilityManualInspectionRequired {
		t.Errorf("RecoverabilityFor(unknown): got %q, want %q", got, state.RecoverabilityManualInspectionRequired)
	}
}

func TestStatePaths(t *testing.T) {
	base := "/base"
	host := "github.com"
	owner := "myowner"
	repo := "myrepo"
	issueNumber := 7
	runID := "run-abc"

	issueDir := state.IssueDir(base, host, owner, repo, issueNumber)
	wantIssueDir := "/base/.kogoto/repositories/github.com/myowner/myrepo/issues/7"
	if issueDir != wantIssueDir {
		t.Errorf("IssueDir: got %q, want %q", issueDir, wantIssueDir)
	}

	issueStatePath := state.IssueStatePath(base, host, owner, repo, issueNumber)
	wantIssueStatePath := wantIssueDir + "/issue-state.json"
	if issueStatePath != wantIssueStatePath {
		t.Errorf("IssueStatePath: got %q, want %q", issueStatePath, wantIssueStatePath)
	}

	runDir := state.RunDir(base, host, owner, repo, issueNumber, runID)
	wantRunDir := wantIssueDir + "/runs/run-abc"
	if runDir != wantRunDir {
		t.Errorf("RunDir: got %q, want %q", runDir, wantRunDir)
	}

	runStatePath := state.RunStatePath(base, host, owner, repo, issueNumber, runID)
	wantRunStatePath := wantRunDir + "/run.json"
	if runStatePath != wantRunStatePath {
		t.Errorf("RunStatePath: got %q, want %q", runStatePath, wantRunStatePath)
	}
}
