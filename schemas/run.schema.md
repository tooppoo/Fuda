# run.json — Semantic Rules and Recovery Policy

Schema definition: [run.schema.json](run.schema.json)

## Purpose

`run.json` は個別 Run の内部状態・再開判断の正本である。

Issue 全体の現在状態と current run 選択は `issue-state.json` が担う。`run.json` は個別 Run に閉じた状態を管理する。詳細は [docs/internal/state-data.md](../docs/internal/state-data.md) を参照。

## Field policy

| Field | Policy |
|---|---|
| `run_id` | UUID string。生成規約として UUIDv7 を使う。JSON Schema では UUID として検証し、v7 であることは semantic rule として扱う |
| `repository` | GitHub `owner/name` 形式。URL は許容しない |
| `issue_number` | integer, minimum `1` |
| `worktree` | absolute path 必須。OS ごとの絶対パス判定は semantic validation で行う |
| `branch` | string, `minLength: 1`。prefix はユーザー指定可能にするため固定 pattern は持たせない |
| `writer.backend` | enum として `claude`, `codex` を含める。v0 で executable backend として許可されるかは semantic validation で判定する |
| `reviewer.backend` | `writer.backend` と同様 |
| `review_loop.completed_review_rounds` | integer, minimum `0`。有効に完了した review round 数。初期値は `0` |
| `review_loop.max_rounds` | integer, minimum `1` |
| `verification_loop.max_retries` | integer, minimum `0`。`config.toml` の `[verification] max_retries` から実行時に採用した上限値。初めて `testing` に遷移した時点で書き込み、以降は変更しない |
| `verification_loop.used_retries` | integer, minimum `0`。writer への修正依頼回数。初期値は `0`。`testing` フェーズ開始時に absent から `{ max_retries: N, used_retries: 0 }` に書く。以降は absent にしない |
| `pull_request` | PR 作成前は absent。作成後は `number` と `url` を持つ object。`null` は使わない |
| `last_error` | `run_state = "failed"` の場合のみ存在する。失敗原因の詳細を持つ object。`run_state != "failed"` の場合は absent とする |
| `created_at` / `updated_at` | RFC3339 UTC timestamp 必須。UTC `Z` 表記を要求する |

## `review_loop`

- `completed_review_rounds` は有効に完了した review round 数を表す
- `review-N.json` の正規化・validation が成功した時点で `N` に更新する
- 次に実行する review number は `completed_review_rounds + 1`
- `completed_review_rounds` は `max_rounds` 以下でなければならない

## `verification_loop`

- `run_state` が初めて `testing` に遷移した時点で `verification_loop: { max_retries: N, used_retries: 0 }` を `run.json` に書く（`N` は `config.toml` の `[verification] max_retries`。未指定時は `2`）
- `testing` → `fixing` 遷移を `run.json` に永続化するタイミングで `used_retries` を加算する（writer 呼び出しより前）
- `used_retries` は run 全体での累積カウントであり、review round をまたいでリセットしない
- `used_retries >= max_retries` で verification がさらに失敗した場合、`run_state = failed` に遷移する
- `review_loop.completed_review_rounds` とは独立して管理する。両者を混同してはならない
- `max_retries = 0` の場合、初回 verification 失敗時に即 `run_state = failed` に遷移する

## `blocked` object

`run_state = "blocked"` の場合のみ存在する。`run_state != "blocked"` の場合は absent とする。

### Field policy

| Field | Policy |
|---|---|
| `questions` | 未解決の blocking question 一覧。required, array, `minItems: 1` |
| `question_comment_id` | Kogoto が投稿した clarification question comment の GitHub comment id。integer, minimum `1` |
| `question_posted_at` | question comment を投稿した RFC3339 UTC timestamp |
| `waiting_since` | run が回答待ちに入った RFC3339 UTC timestamp |
| `last_seen_issue_comment_id` | resume 判定時に参照済みの最新 Issue comment id。integer, minimum `1`。未取得の場合は absent |

`blocked.questions` は履歴ではなく、現在の停止理由の正本とする。

### Resume judgment

resume 時には、少なくとも次を満たす Issue comment を回答候補とする。

- `question_posted_at` より後に投稿されたコメント
- Kogoto 自身の投稿ではないコメント
- bot 投稿ではないコメント
- `kogoto:` marker を含まないコメント

`last_seen_issue_comment_id` が absent の場合でも、`question_posted_at` を基準に回答候補を判定できること。

## `last_error` object

`run_state = "failed"` の場合のみ存在する。`run_state != "failed"` の場合は absent とする。

### Field policy

| Field | Policy |
|---|---|
| `code` | 失敗原因を識別する enum。`schemas/README.md` の error classification を正とする |
| `phase` | 失敗が発生した phase 名 (例: `reviewing`, `committing`)。string, `minLength: 1` |
| `message` | 人間が読める失敗の説明。string, `minLength: 1` |
| `recoverability` | 失敗の復旧可能性を表す enum。値の意味と resume 判断は下記「`recoverability` enum」を正とする |
| `occurred_at` | 失敗が発生した RFC3339 UTC timestamp |
| `artifact` | 関連する debugging artifact のファイル名 (例: `review-1.raw.txt`)。optional |

`run_state = "failed"` かつ `last_error` が absent の state は invalid。
`run_state != "failed"` で `last_error` が存在する state は invalid。

### `recoverability` enum

`recoverability` は表示用情報ではなく、`resume` / retry / abort / manual inspection の判断に使う制御情報である。取りうる値は次の5値に固定する。

| `recoverability` | `resume` 時の挙動 | 意味 |
|---|---|---|
| `retryable` | 自動再試行可能。`kogoto resume` で失敗 phase から再開してよい | transient な失敗（agent 起動失敗・ネットワーク・GitHub API 一時失敗など）。再実行で復旧が見込める |
| `retryable_after_human_confirmation` | 自動再開しない。人間の明示確認後に `kogoto resume` する | 状態自体は壊れていないが、再実行に人間の判断が必要（agent 出力の再生成、closed Issue の再 open など） |
| `retryable_after_manual_fix` | 自動再開しない。人間が Kogoto 管理外の状態を修正した後に `kogoto resume` する | 設定 / git / ファイルシステム等、Kogoto が直せない外部状態を修正しないと再実行が成功しない |
| `manual_inspection_required` | 自動再開しない。人間が artifact / 状態を調査してから対応を決める | 失敗原因や現在状態が不明確で、安全に再開できない |
| `terminal` | `resume` を拒否する | この Run としては復旧不能。再対応は新規 run として扱う |

- `manual_inspection_required` の場合、Kogoto は自動再開してはならない。
- `terminal` の場合、`kogoto resume` は拒否しなければならない。
- `retryable` 以外はいずれも自動再開しない（人間の確認・修正・調査・新規対応のいずれかを要求する）。

各 `last_error.code` の default `recoverability` は下記対応表を正とする。`recoverability` は per-occurrence のフィールドであり、`runner_error` / `environment_error` のような分類外コードでは、runner が occurrence ごとにより適切な値を設定してよい（default は `manual_inspection_required`）。

| `last_error.code` | default `recoverability` |
|---|---|
| `github_auth_failed` | `retryable_after_manual_fix` |
| `issue_not_found` | `terminal` |
| `issue_closed` | `retryable_after_human_confirmation` |
| `main_update_failed` | `retryable` |
| `worktree_create_failed` | `retryable_after_manual_fix` |
| `branch_already_exists` | `manual_inspection_required` |
| `worktree_path_already_exists` | `manual_inspection_required` |
| `writer_launch_failed` | `retryable` |
| `invalid_writer_output` | `retryable_after_human_confirmation` |
| `verification_failed` | `manual_inspection_required` |
| `nothing_to_commit` | `manual_inspection_required` |
| `commit_failed` | `retryable_after_manual_fix` |
| `reviewer_launch_failed` | `retryable` |
| `invalid_reviewer_output` | `retryable_after_human_confirmation` |
| `push_failed` | `retryable` |
| `pr_create_failed` | `retryable` |
| `runner_error` | `manual_inspection_required`（occurrence ごとに上書き可。例: blocked コメント投稿失敗は `retryable`） |
| `environment_error` | `manual_inspection_required`（occurrence ごとに上書き可） |

ケースごとの最終的な対応は `docs/internal/error-handling.md` の Error Matrix を正とする。

---

## Terminal states

`run_state` の終端状態は次の通り。

| `run_state` | 意味 |
|---|---|
| `succeeded` | Kogoto Run が正常終了し、必要な終了処理が完了した。GitHub source issue の close 状態を意味しない |
| `aborted` | ユーザー操作によって中断された |
| `failed` | 不正な状態・agent エラー・git エラーなどにより失敗した。`last_error` に失敗原因の詳細を持つ |

PR 作成済みの正常終了では、状態は次の順に進む。

| From | Trigger | To | Notes |
|---|---|---|---|
| `reviewing` | `runner_decision = pass` and PR creation succeeded | `pr_created` | `run.json.pull_request` に PR number / URL を記録する |
| `reviewing` | `runner_decision = ready_with_minor_findings` and PR creation succeeded | `pr_created` | minor findings は summary に残す |
| `pr_created` | run summary written and finalization completed | `succeeded` | Kogoto Run の正常終了 |

`run_state = "succeeded"` は GitHub source issue が close 済みであることを意味しない。GitHub source issue を close する操作が将来必要な場合は、`source_issue_closed` など対象を明示する event 名を使う。

## Semantic rules

- `repository` は `owner/name` 形式を正とし、URL 形式は受け付けない
- `run_id` は UUIDv7 として生成する
- `worktree` は host OS の path rule に従う absolute path でなければならない
- `branch` の完全な妥当性は JSON Schema ではなく Git ref validation で確認する
- `updated_at` は `run.json` を書き換えるたびに更新する
- `codex` は known backend だが、v0 では executable backend として拒否する
- `run_state = "blocked"` かつ `blocked` が absent の state は invalid
- `run_state = "blocked"` かつ `blocked.questions` が空の state は invalid
- `run_state != "blocked"` で `blocked` が存在する state は invalid
- `run_state = "failed"` かつ `last_error` が absent の state は invalid
- `run_state != "failed"` で `last_error` が存在する state は invalid
- `review_loop.completed_review_rounds` は `max_rounds` 以下でなければならない
- `verification_loop` は `testing` フェーズ開始後は absent にしてはならない
- `verification_loop.used_retries` は `0` から始まり、`max_retries` を超えてはならない
- `verification_loop.max_retries` は `0` 以上の整数でなければならない
- `created_at` / `updated_at` は UTC `Z` 表記を要求する

## Recovery policy

`run.json` は個別 Run の再開判断の正本であるため、壊れていたら `resume` してはならない。

```
run.json parse error / validation error
→ stop
→ do not resume automatically
→ show path and reason
→ retire corrupt file as run.json.corrupt.<timestamp>
→ require human inspection
```
