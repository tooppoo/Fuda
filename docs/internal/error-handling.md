# Error Handling

## Scope

このドキュメントは Kogoto Run における失敗時の挙動を定義する。

正常系フローは `processing-sequence.md` を参照。`run_state` の定義と状態遷移は `state-machine.md` を参照。

---

## 基本方針

- `run_state` は Run ライフサイクル状態のみを表す。失敗原因の詳細は `last_error.code` で表す。
- Kogoto が安全に判断できない場合は、推測で継続せず停止する。
- 既存 branch / worktree / run artifact は自動上書きしない。
- agent raw output は debugging artifact として保存する。parse / normalization 失敗時も削除しない。
- GitHub Issue / PR / branch / worktree への破壊的操作は、人間の明示判断なしに行わない。

---

## `invalid_agent_output` の扱い

`invalid_agent_output` を `run_state` としては使用しない。`run_state` はライフサイクル状態を表すフィールドであり、`invalid_agent_output` は失敗原因である。

agent 出力不正は、`run_state = failed` と `last_error.code` の組み合わせで表す。

- writer 出力不正: `last_error.code = invalid_writer_output`
- reviewer 出力不正: `last_error.code = invalid_reviewer_output`

```json
{
  "run_state": "failed",
  "last_error": {
    "code": "invalid_reviewer_output",
    "phase": "reviewing",
    "message": "Reviewer output could not be parsed or normalized.",
    "recoverability": "retryable_after_human_confirmation",
    "artifact": "review-1.raw.txt",
    "occurred_at": "2026-06-07T00:00:00Z"
  }
}
```

`last_error` の永続形式は `schemas/run.schema.json` を正とする。error code の意味・復旧方針は本ドキュメントの Error Matrix を正とする。

---

## Verification Failure Policy

test / lint / typecheck 失敗は review loop とは別の **verification failure** として扱う。

### `verification_loop` ストレージ

verification retry の進捗は `run.json` の `verification_loop` オブジェクトに必ず記録する。

```json
{
  "verification_loop": {
    "max_retries": 2,
    "used_retries": 0
  }
}
```

- verification が開始された時点で `verification_loop` を `run.json` に書く。
- `used_retries` は writer への修正依頼回数を表す。初期値は `0`。`testing` → `fixing` 遷移を `run.json` に永続化するタイミングで加算する（writer 呼び出しより前）。
- `used_retries` は run 全体での累積カウントであり、review round をまたいでリセットしない。
- `review_loop.completed_review_rounds` とは独立して管理する。両者を混同してはならない。

### retry 上限

retry 上限は `config.toml` の `[verification] max_retries` で設定する（初期値: `2`、最小値: `0`）。実行時に採用した値は `verification_loop.max_retries` として `run.json` に記録する。

### 失敗時の挙動

- `used_retries < max_retries` の場合: `used_retries` を加算して `run_state = fixing` に遷移し（この永続化を完了させてから）、writer に修正を依頼する。
- `used_retries >= max_retries` でさらに失敗した場合: PR は作成せず `run_state = failed` で停止する。

停止時には、失敗した command、exit code、要約 log、次に取れる操作を表示する。

### ログ保持方針

- verification 失敗時の stdout/stderr 全文は run artifact として別ファイルに保持する。
- `run-summary.json` には stdout/stderr 全文を含めない。`failure_summary.artifact_ref` で参照する。

---

## Raw Output Preservation Policy

- agent の raw output は、成否にかかわらず artifact として保存する。
- parse / schema validation / normalization に失敗した場合、正規化済み JSON は書かない。raw output のみ保存する。
  - writer plan output 失敗: `plan.raw.txt` のみ保存、`plan.json` は書かない
  - writer write / fix output 失敗: `implement.log` または `fix-N.log` を保存し、追加の正規化済み JSON は書かない
  - reviewer output 失敗: `review-N.raw.txt` のみ保存、`review-N.json` は書かない
- raw output は `run-summary.json` に含めない。

---

## Existing Branch / Worktree Conflict Policy

既存 branch / worktree / path に conflict がある場合、Kogoto は自動上書きしない。

**Kogoto-managed run を特定できる場合**: 新しい run は作らず、既存 run の状態を表示して停止する。`run_state` は既存 run のものを維持する。

**Kogoto-managed run を特定できない場合**: 新しい run を `run_state = failed` として作成して停止する。

---

## Error Matrix

Run の各 phase における失敗ケースを以下に定義する。`Run state` 列は失敗時の最終 `run_state` を示す。

| Case | Phase | Error code | Retry | Behavior | Run state |
|---|---|---|---|---|---|
| GitHub認証失敗 | `loading_issue` | `github_auth_failed` | 設定修正後に再実行 | 停止 | `failed` |
| Issueが存在しない | `loading_issue` | `issue_not_found` | no | worktree を作らず停止 | `failed` |
| Issueがclosed | `loading_issue` | `issue_closed` | デフォルトでは no | closed Issue を勝手に再 open せず停止 | `failed` |
| main更新失敗 | `preparing_worktree` | `main_update_failed` | yes | worktree 作成前に停止 | `failed` |
| worktree作成失敗 | `preparing_worktree` | `worktree_create_failed` | 手動修正後に再実行 | 既存ファイルを壊さず停止 | `failed` |
| branch既存 (Kogoto管理runあり) | `preparing_worktree` | `branch_already_exists` | 既存 run を resume | 既存 run の状態を表示して停止 | 既存 run の `run_state` |
| branch既存 (Kogoto管理run不明) | `preparing_worktree` | `branch_already_exists` | 手動確認後に再実行 | 上書きせず停止 | `failed` |
| worktree path既存 (Kogoto管理runあり) | `preparing_worktree` | `worktree_path_already_exists` | 既存 run を resume | 既存 run の状態を表示して停止 | 既存 run の `run_state` |
| worktree path既存 (Kogoto管理run不明) | `preparing_worktree` | `worktree_path_already_exists` | 手動確認後に再実行 | 上書きせず停止 | `failed` |
| writer agent起動失敗 | `planning` / `writing` | `writer_launch_failed` | yes | raw なしで停止 | `failed` |
| writer出力が不正 | `planning` / `writing` / `fixing` | `invalid_writer_output` | 明示的な再実行 | raw 保存、正規化 JSON を書かず停止 | `failed` |
| test / lint / typecheck失敗 | `testing` | `verification_failed` | `max_retries` 回まで writer に修正依頼 | retry 中は `testing` → `fixing` を経由。上限到達後: PR 作成せず停止、失敗 command / exit code / 要約 log を表示 | `failed` |
| commit対象変更なし | `committing` | `nothing_to_commit` | デフォルトでは no | PR 作成せず停止 | `failed` |
| commit失敗 | `committing` | `commit_failed` | 手動修正後に再実行 | diff を保持して停止 | `failed` |
| reviewer agent起動失敗 | `reviewing` | `reviewer_launch_failed` | yes | review round を進めず停止 | `failed` |
| reviewer出力が不正 | `reviewing` | `invalid_reviewer_output` | 明示的な再レビュー | `review-N.raw.txt` のみ保存、`review-N.json` を書かず停止 | `failed` |
| push失敗 | `pushing` | `push_failed` | yes | local commit を保持して停止 | `failed` |
| PR作成失敗 | `creating_pr` | `pr_create_failed` | yes | branch / commit を保持して停止 | `failed` |

---

## `run-summary.json` への失敗原因記録

Run が `failed` で終端した場合、`run-summary.json` に失敗概要を `failure_summary` として記録する。

`failure_summary` は `run.json.last_error` から summary 用に安全な情報のみを転記する。

| `run.json.last_error` | `run-summary.json.failure_summary` |
|---|---|
| `code` | `code` |
| `phase` | `phase` |
| `message` | `message` |
| `recoverability` | `recoverability` |
| `occurred_at` | `occurred_at` |
| `artifact` | `artifact_ref` |

`run-summary.json` には以下を含めない。

- 巨大ログ（詳細 test log など）
- 詳細 diff
- 機密情報（token、credential など）
- agent prompt 全文
- command output 全文

詳細 log は run artifact として別ファイルに保持し、`failure_summary.artifact_ref` で参照する。`run-summary.schema.md` を参照。

`aborted` は失敗ではなく user / policy による中断であるため、`failure_summary` は持たせない。

---

## 既存仕様との接続

| 仕様 | 接続点 |
|---|---|
| `state-machine.md` | `run_state` enum および状態遷移 trigger |
| `processing-sequence.md` | resolve / close / blocked フローと失敗箇所の対応 |
| `schemas/run.schema.json` | `run_state` enum と `last_error` の定義 |
| `schemas/run-summary.schema.md` | `terminal_state = failed` と `completion_result` の対応 |
| `schemas/README.md` | raw output 保存方針と schema validation / error classification policy |
