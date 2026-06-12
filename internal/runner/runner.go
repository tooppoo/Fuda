package runner

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/tooppoo/Kogoto/internal/agent"
	"github.com/tooppoo/Kogoto/internal/config"
	"github.com/tooppoo/Kogoto/internal/state"
	"github.com/tooppoo/Kogoto/internal/tracker"
)

type Runner struct {
	base    string
	cfg     *config.Config
	tracker tracker.Tracker
	writer  agent.Writer
}

func New(base string, cfg *config.Config, t tracker.Tracker, w agent.Writer) *Runner {
	return &Runner{base: base, cfg: cfg, tracker: t, writer: w}
}

type ResolveResult struct {
	RunID string
}

func (r *Runner) Resolve(ctx context.Context, issueNumber int) (*ResolveResult, error) {
	now := time.Now().UTC()
	host := r.cfg.GitHub.Host
	owner := r.cfg.GitHub.Owner
	repo := r.cfg.GitHub.Repo
	repository := owner + "/" + repo

	issueStatePath := state.IssueStatePath(r.base, host, owner, repo, issueNumber)

	// Refuse to overwrite an in-progress run.
	if existing, err := state.ReadIssueState(issueStatePath); err == nil {
		switch existing.IssueWorkflowState {
		case "active", "waiting_for_human":
			return nil, fmt.Errorf("issue #%d is already %s (run %s); use `kogoto resume`",
				issueNumber, existing.IssueWorkflowState, existing.CurrentRunID)
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("read issue-state.json: %w", err)
	}

	// Fetch issue (loading_issue phase).
	issue, err := r.tracker.GetIssue(ctx, issueNumber)
	if err != nil {
		return nil, fmt.Errorf("load issue: %w", err)
	}

	runID, err := newUUID()
	if err != nil {
		return nil, fmt.Errorf("generate run ID: %w", err)
	}

	branch := fmt.Sprintf("%s%d", r.cfg.Workspace.BranchPrefix, issueNumber)
	worktree := filepath.Join(r.cfg.Workspace.Root, fmt.Sprintf("%s-issue-%d", repo, issueNumber))

	// Create issue-state.json.
	issueState := &state.IssueState{
		SchemaVersion:      1,
		Repository:         repository,
		IssueNumber:        issueNumber,
		IssueWorkflowState: "active",
		CurrentRunID:       runID,
		Runs:               []state.RunRecord{{RunID: runID, RunResult: "active"}},
		SourceIssue: state.SourceIssue{
			URL:       issue.URL,
			UpdatedAt: issue.UpdatedAt,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := state.WriteIssueState(issueStatePath, issueState); err != nil {
		return nil, fmt.Errorf("write issue-state.json: %w", err)
	}

	runStatePath := state.RunStatePath(r.base, host, owner, repo, issueNumber, runID)
	runState := &state.RunState{
		SchemaVersion: 1,
		RunID:         runID,
		Repository:    repository,
		IssueNumber:   issueNumber,
		RunStateValue: "initialized",
		Branch:        branch,
		Worktree:      worktree,
		Writer:        state.Backend{BackendType: "claude"},
		Reviewer:      state.Backend{BackendType: "claude"},
		ReviewLoop: state.ReviewLoop{
			CompletedReviewRounds: 0,
			MaxRounds:             r.cfg.Review.MaxLoops,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := state.WriteRunState(runStatePath, runState); err != nil {
		return nil, fmt.Errorf("write run.json (initialized): %w", err)
	}

	// Advance through pre-planning states.
	for _, s := range []string{"loading_issue", "preparing_worktree", "planning"} {
		runState.RunStateValue = s
		runState.UpdatedAt = time.Now().UTC()
		if err := state.WriteRunState(runStatePath, runState); err != nil {
			return nil, fmt.Errorf("write run.json (%s): %w", s, err)
		}
	}

	// Call planner.
	planResult, err := r.writer.Plan(ctx, agent.PlanInput{
		RunID:        runID,
		Repository:   repository,
		IssueNumber:  issueNumber,
		IssueTitle:   issue.Title,
		IssueBody:    issue.Body,
		Branch:       branch,
		Worktree:     worktree,
		ArtifactsDir: state.RunDir(r.base, host, owner, repo, issueNumber, runID),
		Now:          time.Now().UTC(),
	})
	if err != nil {
		planErr := fmt.Errorf("plan: %w", err)
		cleanupErr := markFailed(issueStatePath, issueState, runStatePath, runState, "planning", "writer_launch_failed", planErr)
		return nil, errors.Join(planErr, cleanupErr)
	}

	if planResult.Plan.PlanningResult != "blocked_by_ambiguity" {
		unsupported := fmt.Errorf("planning result %q is not yet supported", planResult.Plan.PlanningResult)
		cleanupErr := markFailed(issueStatePath, issueState, runStatePath, runState, "planning", "invalid_writer_output", unsupported)
		return nil, errors.Join(unsupported, cleanupErr)
	}

	// Build blocked questions (only blocking ones go into blocked.questions).
	// Validate before posting to GitHub to avoid leaving a stale comment on failure.
	var blockedQuestions []state.BlockedQuestion
	for _, q := range planResult.Plan.Questions {
		if q.Blocking {
			blockedQuestions = append(blockedQuestions, state.BlockedQuestion{
				ID:       q.ID,
				Question: q.Question,
				Blocking: true,
			})
		}
	}
	if len(blockedQuestions) == 0 {
		noBlocking := fmt.Errorf("blocked_by_ambiguity returned no blocking questions")
		cleanupErr := markFailed(issueStatePath, issueState, runStatePath, runState, "planning", "invalid_writer_output", noBlocking)
		return nil, errors.Join(noBlocking, cleanupErr)
	}

	// Post clarification comment.
	commentBody := formatBlockedComment(runID, issueNumber, planResult.Plan.Questions)
	commentID, err := r.tracker.PostComment(ctx, issueNumber, commentBody)
	if err != nil {
		cleanupErr := markFailed(issueStatePath, issueState, runStatePath, runState, "posting_comment", "runner_error", err)
		return nil, errors.Join(fmt.Errorf("post blocked comment: %w", err), cleanupErr)
	}

	postedAt := time.Now().UTC()

	// Transition to blocked.
	runState.RunStateValue = "blocked"
	runState.Blocked = &state.BlockedInfo{
		Questions:         blockedQuestions,
		QuestionCommentID: commentID,
		QuestionPostedAt:  postedAt,
		WaitingSince:      postedAt,
	}
	runState.UpdatedAt = time.Now().UTC()
	if err := state.WriteRunState(runStatePath, runState); err != nil {
		return nil, fmt.Errorf("write run.json (blocked): %w", err)
	}

	// Update issue workflow state.
	issueState.IssueWorkflowState = "waiting_for_human"
	issueState.UpdatedAt = time.Now().UTC()
	if err := state.WriteIssueState(issueStatePath, issueState); err != nil {
		return nil, fmt.Errorf("write issue-state.json (waiting_for_human): %w", err)
	}

	return &ResolveResult{RunID: runID}, nil
}

func (r *Runner) Status(issueNumber int) error {
	host := r.cfg.GitHub.Host
	owner := r.cfg.GitHub.Owner
	repo := r.cfg.GitHub.Repo

	issueStatePath := state.IssueStatePath(r.base, host, owner, repo, issueNumber)
	issueState, err := state.ReadIssueState(issueStatePath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Issue #%d: no state found\n", issueNumber)
			return nil
		}
		return fmt.Errorf("read issue-state.json: %w", err)
	}

	fmt.Printf("Issue #%d\n", issueNumber)
	fmt.Printf("  workflow state: %s\n", issueState.IssueWorkflowState)

	if issueState.CurrentRunID == "" {
		return nil
	}

	runStatePath := state.RunStatePath(r.base, host, owner, repo, issueNumber, issueState.CurrentRunID)
	runState, err := state.ReadRunState(runStatePath)
	if err != nil {
		return fmt.Errorf("read run.json: %w", err)
	}

	fmt.Printf("\nRun %s\n", runState.RunID)
	fmt.Printf("  run state: %s\n", runState.RunStateValue)

	if runState.Blocked != nil {
		fmt.Printf("\n  Blocked questions:\n")
		for _, q := range runState.Blocked.Questions {
			fmt.Printf("    [%s] %s\n", q.ID, q.Question)
		}
	}

	return nil
}

func formatBlockedComment(runID string, issueNumber int, questions []agent.Question) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "<!-- kogoto:question run=%s issue=%d -->\n\n", runID, issueNumber)
	sb.WriteString("## Kogoto blocked: clarification needed\n\n")
	sb.WriteString("The following questions need to be answered before work can continue.\n\n")
	for i, q := range questions {
		fmt.Fprintf(&sb, "%d. [%s] %s\n", i+1, q.ID, q.Question)
	}
	fmt.Fprintf(&sb, "\nPlease reply to this comment or run `kogoto answer %d` to provide answers.\n", issueNumber)
	return sb.String()
}

func (r *Runner) StatusActive() error {
	host := r.cfg.GitHub.Host
	owner := r.cfg.GitHub.Owner
	repo := r.cfg.GitHub.Repo

	issuesDir := filepath.Join(r.base, ".kogoto", "repositories", host, owner, repo, "issues")
	entries, err := os.ReadDir(issuesDir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No active runs.")
			return nil
		}
		return fmt.Errorf("read issues directory: %w", err)
	}

	active := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		issueNumber, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}
		issueStatePath := state.IssueStatePath(r.base, host, owner, repo, issueNumber)
		issueState, err := state.ReadIssueState(issueStatePath)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return fmt.Errorf("read issue-state.json for issue #%d: %w", issueNumber, err)
		}
		switch issueState.IssueWorkflowState {
		case "active", "waiting_for_human":
		default:
			continue
		}
		active++
		if err := r.Status(issueNumber); err != nil {
			return err
		}
	}
	if active == 0 {
		fmt.Println("No active runs.")
	}
	return nil
}

func markFailed(
	issueStatePath string, issueState *state.IssueState,
	runStatePath string, runState *state.RunState,
	phase, code string, origErr error,
) error {
	now := time.Now().UTC()
	runState.RunStateValue = "failed"
	runState.LastError = &state.LastError{
		Code:           code,
		Phase:          phase,
		Message:        origErr.Error(),
		Recoverability: "retryable",
		OccurredAt:     now,
	}
	runState.UpdatedAt = now

	issueState.IssueWorkflowState = "failed"
	issueState.UpdatedAt = now
	for i, rec := range issueState.Runs {
		if rec.RunID == runState.RunID {
			issueState.Runs[i].RunResult = "failed"
			break
		}
	}

	return errors.Join(
		state.WriteIssueState(issueStatePath, issueState),
		state.WriteRunState(runStatePath, runState),
	)
}

func newUUID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	// UUIDv7: embed Unix millisecond timestamp in the first 48 bits.
	ms := time.Now().UnixMilli()
	b[0] = byte(ms >> 40)
	b[1] = byte(ms >> 32)
	b[2] = byte(ms >> 24)
	b[3] = byte(ms >> 16)
	b[4] = byte(ms >> 8)
	b[5] = byte(ms)
	b[6] = (b[6] & 0x0f) | 0x70 // version 7
	b[8] = (b[8] & 0x3f) | 0x80 // variant 10xx
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:]), nil
}
