# issue-state.json — Semantic Rules and Recovery Policy

Schema definition: [issue-state.schema.json](issue-state.schema.json)

## Purpose

`issue-state.json` は Issue 全体の現在状態・current run 選択の正本である。

個別 Run の内部状態・再開判断の正本は `run.json` が担う。二層構造の詳細は [docs/internal/state-data.md](../docs/internal/state-data.md) を参照。

## Field policy

| Field | Policy |
|---|---|
| `schema_version` | integer, `const: 1` |
| `repository` | GitHub `owner/name` 形式。URL は許容しない |
| `issue_number` | integer, minimum `1` |
| `issue_workflow_state` | Issue 全体のワークフロー状態。下記 enum を参照 |
| `current_run_id` | UUID string。`issue_workflow_state = not_started` の場合は absent。それ以外は required |
| `runs` | Run 履歴の配列。空配列は `issue_workflow_state = not_started` の場合のみ許容 |
| `source_issue.url` | GitHub Issue URL。URI 形式 |
| `source_issue.last_seen_comment_id` | resume 判定時に参照済みの最新 Issue comment id。未取得の場合は absent |
| `source_issue.updated_at` | 最後に取得した Issue の `updated_at`（RFC3339 UTC） |
| `created_at` / `updated_at` | RFC3339 UTC timestamp 必須。UTC `Z` 表記を要求する |

## `issue_workflow_state` Enum

| 値 | 意味 |
|---|---|
| `not_started` | まだ Run が開始されていない |
| `active` | Run が実行中である |
| `waiting_for_human` | 人間の回答・判断を待っている（`run_state = blocked` または `human_review_required` に対応） |
| `pr_created` | PR が作成済みであり、`kogoto close` を待っている |
| `completed` | Issue 処理が完了した（GitHub Issue close 済み） |
| `aborted` | Issue 処理が中断された |
| `failed` | Run が失敗し、Issue 処理が止まっている |

`completed` は GitHub source issue が close 済みであることを表す。`pr_created` とは区別する。

## `runs` Array

- `runs` は Issue に対して実行されたすべての Run の履歴を保持する
- failed / aborted / retried の Run も削除せず、履歴として残す
- `run_result` は Run の最終的な結果分類を表す

| `run_result` | 意味 |
|---|---|
| `active` | Run が現在進行中である |
| `succeeded` | Run が正常終了した |
| `failed` | Run が失敗した |
| `aborted` | Run がユーザー操作によって中断された |
| `retried` | Run が新しい試行に置き換えられた（リトライによって supersede された） |

`run_result = active` の要素は `runs` 配列内で最大1つでなければならない。

## `current_run_id`

`current_run_id` は、現在アクティブな Run または最後に実行された Run の `run_id` を指す。

- `issue_workflow_state = not_started` の場合: absent
- それ以外の場合: required
- `current_run_id` は `runs` 配列内のいずれかの `run_id` と一致しなければならない

## Semantic rules

- `repository` は `owner/name` 形式を正とし、URL 形式は受け付けない
- `current_run_id` は `issue_workflow_state = not_started` の場合は absent
- `current_run_id` は `issue_workflow_state != not_started` の場合は required
- `current_run_id` は `runs[].run_id` のいずれかと一致しなければならない
- `runs` 内で `run_result = active` の要素は最大1つ
- `issue_workflow_state = active` または `waiting_for_human` の場合、`runs` 内に `run_result = active` のエントリが存在する
- `issue_workflow_state = completed` は GitHub source issue が close されたことを意味する。`pr_created` と混同してはならない
- `updated_at` は `issue-state.json` を書き換えるたびに更新する
- `created_at` / `updated_at` は UTC `Z` 表記を要求する

## Recovery policy

`issue-state.json` は Issue 全体の現在状態の正本であるため、壊れていたら Issue 単位のコマンドを実行してはならない。

```
issue-state.json parse error / validation error
→ stop
→ do not execute issue-level commands automatically
→ show path and reason
→ retire corrupt file as issue-state.json.corrupt.<timestamp>
→ require human inspection
```
