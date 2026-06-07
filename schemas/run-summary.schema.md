# run-summary.json — Semantic Rules and Recovery Policy

Schema definition: [run-summary.schema.json](run-summary.schema.json)

## Purpose

`run-summary.json` は succeed / abort / fail 後に残す監査用・振り返り用の軽量サマリである。

以下は含めない: 詳細ログ、プロンプト全文、巨大な diff、private token、環境変数。詳細情報は run artifact として別ファイルに保持し、`artifact_ref` で参照する。

## `terminal_state` / `completion_result` 対応表

`terminal_state` と `completion_result` は意味論上、次の組み合わせを許容する。

| `terminal_state` | 許容される `completion_result` |
|---|---|
| `succeeded` | `pr_created`, `no_pr_created` |
| `aborted` | `aborted_by_user` |
| `failed` | `failed_due_to_invalid_state`, `failed_due_to_agent_error`, `failed_due_to_git_error`, `failed_due_to_github_error`, `failed_due_to_verification_error`, `failed_due_to_runner_error`, `failed_due_to_environment_error` |

この組み合わせ制約は JSON Schema で完全に表現しづらいため、semantic validation として扱う。

## `completion_result` / `failure_summary.code` 対応表

| `completion_result` | `failure_summary.code` |
|---|---|
| `failed_due_to_github_error` | `github_auth_failed`, `issue_not_found`, `issue_closed`, `pr_create_failed` |
| `failed_due_to_git_error` | `main_update_failed`, `worktree_create_failed`, `commit_failed`, `push_failed` |
| `failed_due_to_invalid_state` | `branch_already_exists`, `worktree_path_already_exists`, `nothing_to_commit` |
| `failed_due_to_agent_error` | `writer_launch_failed`, `invalid_writer_output`, `reviewer_launch_failed`, `invalid_reviewer_output` |
| `failed_due_to_verification_error` | `verification_failed` |
| `failed_due_to_runner_error` | `runner_error` |
| `failed_due_to_environment_error` | `environment_error` |

この対応は semantic validation として扱う。JSON Schema レベルでは `failure_summary.code` の enum のみ担保する。

## `pull_request`

- `completion_result = "pr_created"` の場合、`pull_request` は required
- `completion_result != "pr_created"` の場合、`pull_request` は absent

## `failure_summary`

- `terminal_state = "failed"` の場合、`failure_summary` は required
- `terminal_state != "failed"` の場合、`failure_summary` は absent

`failure_summary` は `run.json.last_error` から summary 用に安全な情報のみを転記する。

### `failure_summary` フィールド

| フィールド | 必須 | 説明 |
|---|---|---|
| `code` | required | 失敗原因コード。`run.json.last_error.code` から転記する |
| `phase` | required | 失敗が発生した run phase |
| `message` | required | 人間向けの短い要約 (maxLength: 500) |
| `recoverability` | required | 復旧可能性の分類 |
| `occurred_at` | required | 失敗発生時刻 (RFC3339 UTC) |
| `artifact_ref` | optional | 詳細情報を持つ artifact への相対参照 |

### `failure_summary.message`

`message` は人間向けの短い要約に限定する。JSON Schema で `maxLength: 500` を設定する。

詳細ログ、command output 全文、巨大 diff、prompt 本文などは `run-summary.json` に含めない。これらは artifact として別ファイルに保持し、`artifact_ref` で参照する。

### `failure_summary.artifact_ref`

`artifact_ref` は run-local な artifact への相対参照のみを許可する。

以下は禁止する:

- absolute path（`/` 始まり）
- Windows drive path（`C:` 等）
- backslash（`\`）を含む path（Windows/UNC 形式 `\\server\share`、`..\\secret.log` 等による迂回を防ぐ）
- `..` path segment（親ディレクトリ参照）
- log / diff / prompt の本文そのもの

`artifact_ref` は artifact への参照であり、artifact の内容を格納するフィールドではない。

### `run.json.last_error` からの転記ルール

| `run.json.last_error` | `run-summary.json.failure_summary` |
|---|---|
| `code` | `code` |
| `phase` | `phase` |
| `message` | `message` |
| `recoverability` | `recoverability` |
| `occurred_at` | `occurred_at` |
| `artifact` | `artifact_ref` |

`message` は人間向けの短い要約に限る。詳細ログや command output 全文は artifact 側に保持し、`run-summary.json` には含めない。

## `review_rounds`

- succeed / abort / fail 時点の `run.json.review_loop.completed_review_rounds` の値を記録する
- raw output の生成回数や invalid review output の試行回数ではない

## `finished_at`

`finished_at` は Run が `succeeded` / `aborted` / `failed` になった時点の RFC3339 UTC timestamp を記録する。`terminal_state` によらず同一フィールドを使う。

## Semantic rules

- `terminal_state` と `completion_result` の組み合わせは上記対応表に従う
- `completion_result = "pr_created"` の場合、`pull_request` は required
- `completion_result != "pr_created"` の場合、`pull_request` は absent
- `terminal_state = "failed"` の場合、`failure_summary` は required
- `terminal_state != "failed"` の場合、`failure_summary` は absent
- `completion_result` と `failure_summary.code` の対応は上記対応表に従う
- `failure_summary` は `run.json.last_error` から安全な情報のみを転記したものである
- `failure_summary.artifact_ref` は run-local な artifact への相対参照のみ許可する
- `review_rounds` は `run.json.review_loop.completed_review_rounds` の終端時点の値
- `started_at` / `finished_at` は UTC `Z` 表記を要求する

## Recovery policy

終了後の記録であるため、実行安全性への影響は比較的小さい。

```
run-summary.json invalid
→ warn
→ do not delete
→ if run.json or PR information remains, allow explicit regeneration by human command
```
