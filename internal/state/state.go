package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/santhosh-tekuri/jsonschema/v6"
	kogotoschemas "github.com/tooppoo/Kogoto/schemas"
)

// RunStateValue is the run_state field of run.json.
type RunStateValue string

const (
	RunStateInitialized         RunStateValue = "initialized"
	RunStateLoadingIssue        RunStateValue = "loading_issue"
	RunStatePreparingWorktree   RunStateValue = "preparing_worktree"
	RunStatePlanning            RunStateValue = "planning"
	RunStateBlocked             RunStateValue = "blocked"
	RunStateWriting             RunStateValue = "writing"
	RunStateTesting             RunStateValue = "testing"
	RunStateCommitting          RunStateValue = "committing"
	RunStateReviewing           RunStateValue = "reviewing"
	RunStateFixing              RunStateValue = "fixing"
	RunStateHumanReviewRequired RunStateValue = "human_review_required"
	RunStatePrCreated           RunStateValue = "pr_created"
	RunStateSucceeded           RunStateValue = "succeeded"
	RunStateAborted             RunStateValue = "aborted"
	RunStateFailed              RunStateValue = "failed"
)

// IssueWorkflowState is the issue_workflow_state field of issue-state.json.
type IssueWorkflowState string

const (
	IssueWorkflowStateNotStarted      IssueWorkflowState = "not_started"
	IssueWorkflowStateActive          IssueWorkflowState = "active"
	IssueWorkflowStateWaitingForHuman IssueWorkflowState = "waiting_for_human"
	IssueWorkflowStatePrCreated       IssueWorkflowState = "pr_created"
	IssueWorkflowStateCompleted       IssueWorkflowState = "completed"
	IssueWorkflowStateAborted         IssueWorkflowState = "aborted"
	IssueWorkflowStateFailed          IssueWorkflowState = "failed"
)

// RunResult is the run_result field of a RunRecord.
type RunResult string

const (
	RunResultActive    RunResult = "active"
	RunResultSucceeded RunResult = "succeeded"
	RunResultFailed    RunResult = "failed"
	RunResultAborted   RunResult = "aborted"
	RunResultRetried   RunResult = "retried"
)

// BackendType identifies an agent backend.
type BackendType string

const (
	BackendTypeClaude BackendType = "claude"
	BackendTypeCodex  BackendType = "codex"
)

// ErrorCode is the last_error.code field of run.json.
type ErrorCode string

const (
	ErrorCodeWriterLaunchFailed  ErrorCode = "writer_launch_failed"
	ErrorCodeInvalidWriterOutput ErrorCode = "invalid_writer_output"
	ErrorCodeRunnerError         ErrorCode = "runner_error"
	ErrorCodeIssueNotFound       ErrorCode = "issue_not_found"
	ErrorCodeIssueIsPR           ErrorCode = "issue_is_pr"
	ErrorCodeVerificationFailed  ErrorCode = "verification_failed"
)

// IssueState is the content of issue-state.json.
type IssueState struct {
	SchemaVersion      int                `json:"schema_version"`
	Repository         string             `json:"repository"`
	IssueNumber        int                `json:"issue_number"`
	IssueWorkflowState IssueWorkflowState `json:"issue_workflow_state"`
	CurrentRunID       string             `json:"current_run_id,omitempty"`
	Runs               []RunRecord        `json:"runs"`
	SourceIssue        SourceIssue        `json:"source_issue"`
	CreatedAt          time.Time          `json:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at"`
}

type RunRecord struct {
	RunID     string    `json:"run_id"`
	RunResult RunResult `json:"run_result"`
}

type SourceIssue struct {
	URL               string    `json:"url"`
	LastSeenCommentID *int64    `json:"last_seen_comment_id,omitempty"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// RunState is the content of run.json.
type RunState struct {
	SchemaVersion    int               `json:"schema_version"`
	RunID            string            `json:"run_id"`
	Repository       string            `json:"repository"`
	IssueNumber      int               `json:"issue_number"`
	RunStateValue    RunStateValue     `json:"run_state"`
	Branch           string            `json:"branch"`
	Worktree         string            `json:"worktree"`
	Writer           Backend           `json:"writer"`
	Reviewer         Backend           `json:"reviewer"`
	ReviewLoop       ReviewLoop        `json:"review_loop"`
	VerificationLoop *VerificationLoop `json:"verification_loop,omitempty"`
	Blocked          *BlockedInfo      `json:"blocked,omitempty"`
	LastError        *LastError        `json:"last_error,omitempty"`
	PullRequest      *PullRequest      `json:"pull_request,omitempty"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
}

type Backend struct {
	BackendType BackendType `json:"backend"`
}

type ReviewLoop struct {
	CompletedReviewRounds int `json:"completed_review_rounds"`
	MaxRounds             int `json:"max_rounds"`
}

type VerificationLoop struct {
	MaxRetries  int `json:"max_retries"`
	UsedRetries int `json:"used_retries"`
}

type BlockedInfo struct {
	Questions              []BlockedQuestion `json:"questions"`
	QuestionCommentID      int64             `json:"question_comment_id"`
	QuestionPostedAt       time.Time         `json:"question_posted_at"`
	WaitingSince           time.Time         `json:"waiting_since"`
	LastSeenIssueCommentID *int64            `json:"last_seen_issue_comment_id,omitempty"`
}

type BlockedQuestion struct {
	ID       string `json:"id"`
	Question string `json:"question"`
	Blocking bool   `json:"blocking"`
}

type LastError struct {
	Code           ErrorCode `json:"code"`
	Phase          string    `json:"phase"`
	Message        string    `json:"message"`
	Recoverability string    `json:"recoverability"`
	OccurredAt     time.Time `json:"occurred_at"`
	Artifact       string    `json:"artifact,omitempty"`
}

type PullRequest struct {
	Number int    `json:"number"`
	URL    string `json:"url"`
}

// IssueDir returns the directory for a given issue's state files.
func IssueDir(base, host, owner, repo string, issueNumber int) string {
	return filepath.Join(base, ".kogoto", "repositories", host, owner, repo, "issues", fmt.Sprintf("%d", issueNumber))
}

// IssueStatePath returns the path to issue-state.json.
func IssueStatePath(base, host, owner, repo string, issueNumber int) string {
	return filepath.Join(IssueDir(base, host, owner, repo, issueNumber), "issue-state.json")
}

// RunDir returns the directory for a specific run's artifacts.
func RunDir(base, host, owner, repo string, issueNumber int, runID string) string {
	return filepath.Join(IssueDir(base, host, owner, repo, issueNumber), "runs", runID)
}

// RunStatePath returns the path to run.json for a specific run.
func RunStatePath(base, host, owner, repo string, issueNumber int, runID string) string {
	return filepath.Join(RunDir(base, host, owner, repo, issueNumber, runID), "run.json")
}

func ReadIssueState(path string) (*IssueState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := validateJSON(compiledIssueStateSchema, data); err != nil {
		return nil, fmt.Errorf("schema validation %s: %w", path, err)
	}
	var s IssueState
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return &s, nil
}

func WriteIssueState(path string, s *IssueState) error {
	return writeJSON(path, s, compiledIssueStateSchema)
}

func ReadRunState(path string) (*RunState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := validateJSON(compiledRunStateSchema, data); err != nil {
		return nil, fmt.Errorf("schema validation %s: %w", path, err)
	}
	var s RunState
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return &s, nil
}

func WriteRunState(path string, s *RunState) error {
	return writeJSON(path, s, compiledRunStateSchema)
}

var (
	compiledRunStateSchema   *jsonschema.Schema
	compiledIssueStateSchema *jsonschema.Schema
)

func init() {
	must := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	c := jsonschema.NewCompiler()
	c.AssertFormat()

	var runSchemaDoc, issueStateSchemaDoc any
	must(json.Unmarshal(kogotoschemas.RunSchemaJSON, &runSchemaDoc))
	must(json.Unmarshal(kogotoschemas.IssueStateSchemaJSON, &issueStateSchemaDoc))

	must(c.AddResource("run.schema.json", runSchemaDoc))
	var err error
	compiledRunStateSchema, err = c.Compile("run.schema.json")
	must(err)

	must(c.AddResource("issue-state.schema.json", issueStateSchemaDoc))
	compiledIssueStateSchema, err = c.Compile("issue-state.schema.json")
	must(err)
}

func writeJSON(path string, v interface{}, schema *jsonschema.Schema) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	if err := validateJSON(schema, data); err != nil {
		return fmt.Errorf("schema validation: %w", err)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, append(data, '\n'), 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func validateJSON(schema *jsonschema.Schema, data []byte) error {
	var doc any
	if err := json.Unmarshal(data, &doc); err != nil {
		return err
	}
	return schema.Validate(doc)
}
