package state_test

import (
	"testing"
	"time"

	"github.com/tooppoo/Kogoto/internal/state"
)

func TestIssueStateRoundTrip(t *testing.T) {
	dir := t.TempDir()
	now := time.Now().UTC().Truncate(time.Second)

	path := state.IssueStatePath(dir, "github.com", "owner", "repo", 42)

	original := &state.IssueState{
		SchemaVersion:      1,
		Repository:         "owner/repo",
		IssueNumber:        42,
		IssueWorkflowState: "waiting_for_human",
		CurrentRunID:       "abc123",
		Runs:               []state.RunRecord{{RunID: "abc123", RunResult: "active"}},
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

	if got.IssueWorkflowState != "waiting_for_human" {
		t.Errorf("IssueWorkflowState: got %q, want %q", got.IssueWorkflowState, "waiting_for_human")
	}
	if got.CurrentRunID != "abc123" {
		t.Errorf("CurrentRunID: got %q, want %q", got.CurrentRunID, "abc123")
	}
	if len(got.Runs) != 1 || got.Runs[0].RunID != "abc123" {
		t.Errorf("Runs mismatch: %+v", got.Runs)
	}
}

func TestRunStateRoundTrip(t *testing.T) {
	dir := t.TempDir()
	now := time.Now().UTC().Truncate(time.Second)

	path := state.RunStatePath(dir, "github.com", "owner", "repo", 42, "run-001")

	original := &state.RunState{
		SchemaVersion: 1,
		RunID:         "run-001",
		Repository:    "owner/repo",
		IssueNumber:   42,
		RunStateValue: "blocked",
		Branch:        "kogoto/issue-42",
		Worktree:      "/tmp/worktrees/repo-issue-42",
		Writer:        state.Backend{BackendType: "claude"},
		Reviewer:      state.Backend{BackendType: "claude"},
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

	if got.RunStateValue != "blocked" {
		t.Errorf("RunStateValue: got %q, want %q", got.RunStateValue, "blocked")
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
