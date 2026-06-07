# Fuda

**Human-in-the-loop-first AI runner for issue-driven development.**

Fuda is a lightweight, local-first, CLI-first runner for issue-driven development.

It treats each GitHub Issue as a small work ticket — a **fuda** — coordinates writer and reviewer agents, and helps move the task from implementation to review to pull request.

Fuda is not a fully autonomous development agent. It is designed to keep scope, judgment, review, and release control in human hands while automating the repetitive parts of the development loop.

## Core concept

**Fuda is Human-in-the-loop first.**

This does not mean humans merely approve at the end. Human-in-the-loop first means that the human remains in control of scope, judgment, and direction at every significant step.

### Weak HITL vs Strong HITL

Fuda distinguishes between two kinds of human-in-the-loop:

- **Weak HITL**: AI executes most of the work; the human approves at key gates. The human is a gatekeeper, confirming AI results after the fact.
- **Strong HITL**: Human and AI interact incrementally, step by step. AI handles execution; humans and AI collaborate on analysis, judgment, and evaluation. AI does not assume authority it was not given.

Fuda aims for **strong HITL**.

In Fuda, agents may analyze, propose, revise, and flag uncertainty — but they do not make unilateral decisions. When a decision point is unclear or significant, the run stops and asks the human. Fuda keeps judgment in human hands not because AI cannot guess, but because judgment is better formed through human-agent interaction than through automated substitution.

### What humans control

Humans retain control over:

- **Issue scope** — what the Issue covers, what it does not cover, and whether the scope should change
- **Blocked questions** — ambiguous points raised by the writer agent during planning or implementation; Fuda stops and waits for a human answer rather than guessing
- **Reviewer-requested human confirmation** — cases where the reviewer agent cannot decide automatically and signals that a human must decide before the run continues
- **Merge and release** — Fuda creates pull requests but does not merge them or push to the main branch
- **Abort, resume, and close** — the run lifecycle is driven by explicit human commands

### What agents do

Writer and reviewer agents handle the repetitive parts of the development loop:

- Writer agents implement, revise, and document according to the Issue scope
- Reviewer agents inspect diffs, test results, and acceptance criteria
- Both agents report their findings, decision rationale, and uncertainty — they do not treat guesses as settled conclusions

Agents support human judgment. They do not replace it.

### Connection to MVP v0 design

Human-in-the-loop first is not a safety feature added on top of Fuda — it is the design foundation. Several MVP v0 behaviors express this directly:

- **Blocked flow**: When the writer agent encounters ambiguity during planning or implementation, the run stops. Fuda posts a question to the Issue and waits for a human answer. The agent does not guess and proceed.
- **`human_review_required`**: When the reviewer agent flags a case requiring human judgment, the run stops. The human decides whether to continue, revise, or accept as-is.
- **Verification retry limit**: Verification failures trigger a fix cycle, but the retry count is capped. When the limit is reached, the run stops rather than continuing to loop. Fuda does not allow unlimited self-correction.
- **No PR merge**: Fuda creates pull requests but does not merge them. Merge and release decisions belong to the human.
- **No direct push to main**: Fuda works in isolated worktrees and branches. It does not push directly to the main branch.
- **Explicit close and cleanup**: Closing a completed run is a separate, explicit command. Fuda does not automatically close Issues, delete branches, or clean up worktrees.

For the full state machine including `blocked` and `human_review_required` transitions, see [Run State Machine](docs/internal/state-machine.md).
For the detailed HITL concept and knowledge workflow, see [Human-in-the-loop first](docs/concept/human-in-the-loop-first.md).

## Design principles

### Human-controlled

Humans remain responsible for selecting work, defining scope, answering unclear points, accepting or rejecting changes, and deciding whether work should proceed.

AI agents may write, review, or revise, but Fuda keeps the workflow checkpoint-driven. Agents do not expand scope, override acceptance criteria, or continue past ambiguity without human input.

### Local-first

Fuda is designed to run from the developer's local environment.

Workspaces, run state, logs, and summaries should be inspectable locally by default.
Remote services such as GitHub are providers, not Fuda's control plane.

### CLI-first

Fuda is primarily a command-line tool.

The initial workflow is intentionally small and explicit, centered around commands such as resolving an Issue, checking status, answering blocked questions, resuming a run, and closing a completed run.

### Lightweight

Fuda should remain small enough for individual developers and small teams.

It should not require a hosted dashboard, a multi-tenant service, or a general-purpose workflow engine.

### Agent-agnostic

Fuda is not intended to be tied to one AI coding agent.

The initial implementation may prioritize specific agents, but the core model distinguishes roles such as writer and reviewer from any particular agent backend.

## What Fuda is

Fuda is:

* an issue-driven AI runner
* a controlled handoff tool
* a local-first developer tool
* a checkpoint-driven workflow around writer agents, reviewer agents, and human decisions

## What Fuda is not

Fuda is not:

* an autonomous backlog runner
* a hosted agent control plane
* a general-purpose AI agent framework
* a project management tool
* a replacement for human review
* a tool for letting agents merge or land changes without human decision

## Current direction

The initial version focuses on GitHub Issues and a single repository.

A typical run starts from one Issue, creates an isolated worktree, asks a writer agent to work on it, asks a reviewer agent to inspect the diff, loops through revisions when needed, and prepares a pull request when the work is ready.

If the Issue is unclear, Fuda should stop, ask a question, and wait for a human answer rather than guessing.

Future versions may support other issue trackers, repository providers, and agent backends, but the core concept remains the same:

> Issue trackers describe work.
> AI agents assist with work.
> Fuda controls the handoff.
