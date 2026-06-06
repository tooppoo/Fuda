# ログ仕様

## Run Log

Fudaは各runのログをローカルに保存する。

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
  plan.json         # writerが作成したplan
  implement.log     # writer実行ログ
  test-1.log        # 1回目のtest/lint/typecheck実行ログ
  review-1.json     # 1回目のreviewerレビュー結果
  fix-1.log         # 1回目の修正ログ
  test-2.log        # 2回目のtest/lint/typecheck実行ログ
  review-2.json     # 2回目のレビュー結果
  pr.md             # PR本文
```

ループが続く場合は `review-N.json` / `fix-N.log` / `test-N.log` が追加される。

---

## ファイル仕様

### `run.json`

runの状態とメタデータを管理するファイル。

```json
{
  "run_id": "<run-id>",
  "issue": 7,
  "repository": "tooppoo/fuda",
  "branch": "fuda/issue-7",
  "worktree": "~/src/fuda-worktrees/fuda-issue-7",
  "status": "reviewing",
  "writer": "claude",
  "reviewer": "claude/code-reviewer",
  "review_loop": 1,
  "max_review_loops": 3,
  "created_at": "2026-06-06T00:00:00Z",
  "updated_at": "2026-06-06T00:20:00Z"
}
```

### `plan.json`

writerが作成したplan。blockedの場合は `status: "blocked"` を含む。

passの場合:

```json
{
  "status": "ready",
  "steps": [...]
}
```

blockedの場合:

```json
{
  "status": "blocked",
  "reason": "acceptance criteria are ambiguous",
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

N回目のレビュー結果。

passの場合:

```json
{
  "status": "pass",
  "findings": [],
  "human_review_required": []
}
```

指摘がある場合:

```json
{
  "status": "needs_fix",
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
  "human_review_required": []
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
| `implement.log` | writer実行ログ |
| `test-*.log` | テスト実行ログ |
| `review-*.json` | レビュー結果 |
| `fix-*.log` | 修正ログ |
| `pr.md` | PR本文 |

### 残すファイル

| ファイル | 内容 |
|---------|------|
| `run-summary.json` | run全体のサマリ |

---

## `run-summary.json`

close後に残すsummary。

```json
{
  "issue": 7,
  "repository": "tooppoo/fuda",
  "branch": "fuda/issue-7",
  "worktree": "~/src/fuda-worktrees/fuda-issue-7",
  "pull_request": 12,
  "status": "closed",
  "writer": "claude",
  "reviewer": "claude/code-reviewer",
  "review_loops": 2,
  "minor_findings": [
    {
      "id": "r3",
      "category": "maintainability",
      "message": "Consider renaming this helper."
    }
  ],
  "created_at": "2026-06-06T00:00:00Z",
  "pr_created_at": "2026-06-06T00:30:00Z",
  "closed_at": "2026-06-06T01:00:00Z",
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
