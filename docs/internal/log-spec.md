# ログ仕様

> JSON フォーマットの正本は [`schemas/`](../../schemas/README.md) 以下の各スキーマファイル。このドキュメントはファイル構成・目的・サンプルを示す。

## Run Log

### 目的

* 後から検証できるようにする
* agentの判断を追跡できるようにする
* 失敗時に再開・調査しやすくする
* PR本文生成に利用する

---

## 作業中のログ構造

作業中は以下のファイルがrunディレクトリに作成される。

```
.fuda/runs/{issue_number}/
  run.json          # run全体のメタデータ・状態
  issue.md          # 取得したIssue本文とコメント
  plan.json         # writerが作成したplan（正規化済み）
  plan.raw.txt      # writerのraw output（debugging artifact）
  implement.log     # writer実行ログ
  test-1.log        # 1回目のtest/lint/typecheck実行ログ
  review-1.json     # 1回目のreviewerレビュー結果（正規化済み）
  review-1.raw.txt  # 1回目のreviewer raw output（debugging artifact）
  fix-1.log         # 1回目の修正ログ
  test-2.log        # 2回目のtest/lint/typecheck実行ログ
  review-2.json     # 2回目のレビュー結果
  review-2.raw.txt  # 2回目のreviewer raw output
  pr.md             # PR本文
```

ループが続く場合は `review-N.json` / `review-N.raw.txt` / `fix-N.log` / `test-N.log` が追加される。

---

## ファイル仕様

### `run.json`

runの状態とメタデータを管理するファイル。スキーマ: [run.schema.json](../../schemas/run.schema.json) / [run.schema.md](../../schemas/run.schema.md)

```json
{
  "schema_version": 1,
  "run_id": "<run-id>",
  "repository": "tooppoo/fuda",
  "issue_number": 7,
  "run_state": "reviewing",
  "branch": "fuda/issue-7",
  "worktree": "/home/user/src/fuda-worktrees/fuda-issue-7",
  "writer": { "backend": "claude" },
  "reviewer": { "backend": "claude" },
  "review_loop": {
    "completed_review_rounds": 1,
    "max_rounds": 3
  },
  "created_at": "2026-06-06T00:00:00Z",
  "updated_at": "2026-06-06T00:20:00Z"
}
```

### `plan.json`

writerが作成したplan（正規化済み）。スキーマ: [plan.schema.json](../../schemas/plan.schema.json) / [plan.schema.md](../../schemas/plan.schema.md)

`planning_result = "ready_to_write"` の場合:

```json
{
  "schema_version": 1,
  "planning_result": "ready_to_write",
  "summary": "Implement schema validation for run.json and plan.json.",
  "tasks": [
    { "id": "task-1", "description": "Define run.json JSON Schema." },
    { "id": "task-2", "description": "Define plan.json JSON Schema." }
  ],
  "questions": []
}
```

`planning_result = "blocked_by_ambiguity"` の場合:

```json
{
  "schema_version": 1,
  "planning_result": "blocked_by_ambiguity",
  "summary": "Issue acceptance criteria are ambiguous.",
  "tasks": [],
  "questions": [
    {
      "id": "q1",
      "question": "Should Fuda create a PR automatically, or only generate a PR body?",
      "blocking": true
    }
  ]
}
```

### `review-N.json`

N回目のレビュー結果（正規化済み）。スキーマ: [review.schema.json](../../schemas/review.schema.json) / [review.schema.md](../../schemas/review.schema.md)

passの場合:

```json
{
  "schema_version": 1,
  "review_number": 1,
  "reviewer_assessment": "pass",
  "findings": [],
  "human_review_required": [],
  "runner_decision": "pass"
}
```

指摘がある場合:

```json
{
  "schema_version": 1,
  "review_number": 1,
  "reviewer_assessment": "needs_fix",
  "findings": [
    {
      "id": "r1",
      "severity": "blocking",
      "category": "bug",
      "file": "src/runner.ts",
      "line": 120,
      "message": "The runner may treat its own comment as a user answer.",
      "required_fix": "Ignore comments authored by the configured GitHub actor and comments containing fuda markers."
    }
  ],
  "human_review_required": [],
  "runner_decision": "needs_revision"
}
```

---

## close後のログ構造

`fuda close` 実行後は詳細ファイルを削除し、summaryのみ残す。

```
~/.local/state/fuda/runs/{repo_owner}-{repo_name}/{issue_number}/
  run-summary.json
```

### 削除対象ファイル（中間ファイル）

| ファイル | 内容 |
|---------|------|
| `issue.md` | Issueの内容 |
| `plan.json` | writerのplan |
| `plan.raw.txt` | writer raw output |
| `implement.log` | writer実行ログ |
| `test-*.log` | テスト実行ログ |
| `review-*.json` | レビュー結果 |
| `review-*.raw.txt` | reviewer raw output |
| `fix-*.log` | 修正ログ |
| `pr.md` | PR本文 |

### 残すファイル

| ファイル | 内容 |
|---------|------|
| `run-summary.json` | run全体のサマリ |

---

## `run-summary.json`

close後に残すsummary。スキーマ: [run-summary.schema.json](../../schemas/run-summary.schema.json) / [run-summary.schema.md](../../schemas/run-summary.schema.md)

```json
{
  "schema_version": 1,
  "repository": "tooppoo/fuda",
  "issue_number": 7,
  "run_id": "<run-id>",
  "branch": "fuda/issue-7",
  "worktree": "/home/user/src/fuda-worktrees/fuda-issue-7",
  "pull_request": {
    "number": 12,
    "url": "https://github.com/tooppoo/fuda/pull/12"
  },
  "final_run_state": "closed",
  "completion_result": "pr_created",
  "started_at": "2026-06-06T00:00:00Z",
  "finished_at": "2026-06-06T01:00:00Z",
  "review_rounds": 2,
  "minor_findings": [
    {
      "id": "r3",
      "category": "maintainability",
      "message": "Consider renaming this helper."
    }
  ],
  "forced": false
}
```

### summaryを残す目的

* 作業履歴を最低限追跡できるようにする
* 後からどのIssueをFudaが処理したか確認できるようにする
* 詳細ログを残し続けることによるローカル肥大化を避ける
* AI判断の詳細は残さず、運用履歴だけを残す
* v0で修正ループを起動しなかったminor findingsを追跡できるようにする

---

## ストレージパス

| 用途 | パス |
|------|------|
| 作業中ログ | `.fuda/runs/{issue_number}/` |
| close後summary | `~/.local/state/fuda/runs/{repo_owner}-{repo_name}/{issue_number}/` |

---

## Issueコメントの管理

Fudaが投稿するIssueコメントには、Fuda管理用markerを含める。

Fuda自身のコメントとユーザーコメントを識別するために使用する。

### markerフォーマット

```
<!-- fuda:{type} {attributes} -->
```

| type | 用途 |
|------|------|
| `question` | blocked時の質問コメント |
| `run-status` | runの状態報告 |

### question markerの例

```
<!-- fuda:question run=<run-id> issue=7 question=q1 -->
```

### run-status markerの例

```
<!-- fuda:run-status issue=7 run=<run-id> -->
```
