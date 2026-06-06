# Fuda

**Human-controlled AI handoff for issue-driven development.**

Fuda is a lightweight, local-first, CLI-first runner for handing issue-based software work to AI coding agents while keeping humans in control.

Fuda treats an Issue as a **fuda** — a work ticket prepared for AI-assisted execution.

It helps developers move a small unit of work through writing, review, revision, and pull request preparation with explicit human checkpoints.

## Core concept

Fuda is not an autonomous development system.

It does not treat an issue tracker as a backlog queue to be continuously consumed by agents.
It does not aim to maximize unattended automation.
It does not let AI agents change the purpose, scope, or acceptance criteria of an Issue on their own.

Instead, Fuda focuses on **controlled handoff**.

A human selects an Issue.
Fuda prepares an isolated working context.
A writer agent proposes or applies changes.
A reviewer agent inspects the result.
If the work is unclear, Fuda stops and asks for clarification.
If the result is acceptable, Fuda helps prepare a pull request.

The goal is not to remove humans from the workflow.
The goal is to make AI-assisted development **small, inspectable, resumable, and stoppable**.

## Design principles

### Human-controlled

Humans remain responsible for selecting work, answering unclear points, accepting or rejecting changes, and deciding whether work should proceed.

AI agents may write, review, or revise, but Fuda keeps the workflow checkpoint-driven.

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
