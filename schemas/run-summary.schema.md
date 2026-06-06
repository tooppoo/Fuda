# run-summary.json — Semantic Rules and Recovery Policy

Schema definition: [run-summary.schema.json](run-summary.schema.json)

## Purpose

`run-summary.json` は close / abort / fail 後に残す監査用・振り返り用の軽量サマリである。

以下は含めない: 詳細ログ、プロンプト全文、巨大な diff、private token、環境変数。

## `final_run_state` / `completion_result` 対応表

`final_run_state` と `completion_result` は意味論上、次の組み合わせを許容する。

| `final_run_state` | 許容される `completion_result` |
|---|---|
| `closed` | `pr_created`, `closed_without_pr` |
| `aborted` | `aborted_by_user` |
| `failed` | `failed_due_to_invalid_state`, `failed_due_to_agent_error`, `failed_due_to_git_error` |

この組み合わせ制約は JSON Schema で完全に表現しづらいため、semantic validation として扱う。

## `pull_request`

- `completion_result = "pr_created"` の場合、`pull_request` は required
- `completion_result != "pr_created"` の場合、`pull_request` は absent

## `review_rounds`

- close / abort / fail 時点の `run.json.review_loop.completed_review_rounds` の値を記録する
- raw output の生成回数や invalid review output の試行回数ではない

## `finished_at`

`finished_at` は Run が `closed` / `aborted` / `failed` になった時点の RFC3339 UTC timestamp を記録する。`final_run_state` によらず同一フィールドを使う。

## Semantic rules

- `final_run_state` と `completion_result` の組み合わせは上記対応表に従う
- `completion_result = "pr_created"` の場合、`pull_request` は required
- `completion_result != "pr_created"` の場合、`pull_request` は absent
- `review_rounds` は `run.json.review_loop.completed_review_rounds` の終端時点の値
- `started_at` / `finished_at` は UTC `Z` 表記を要求する

## Recovery policy

close 後の記録であるため、実行安全性への影響は比較的小さい。

```
run-summary.json invalid
→ warn
→ do not delete
→ if run.json or PR information remains, allow explicit regeneration by human command
```
