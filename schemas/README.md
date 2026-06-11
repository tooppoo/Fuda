# JSON Schemas and Recovery Policy

永続ファイルのスキーマ定義・semantic rules・復旧方針の正本は各ファイルである。このドキュメントはスコープ・共通方針・エラー分類を定める。

## 1. Scope

Kogoto が永続管理する JSON ファイルと対応スキーマファイルは以下の通り。

| ファイル | JSON Schema | Semantic rules / Recovery policy |
|---|---|---|
| `issue-state.json` | [issue-state.schema.json](issue-state.schema.json) | [issue-state.schema.md](issue-state.schema.md) |
| `run.json` | [run.schema.json](run.schema.json) | [run.schema.md](run.schema.md) |
| `plan.json` | [plan.schema.json](plan.schema.json) | [plan.schema.md](plan.schema.md) |
| `review-N.json` | [review.schema.json](review.schema.json) | [review.schema.md](review.schema.md) |
| `run-summary.json` | [run-summary.schema.json](run-summary.schema.json) | [run-summary.schema.md](run-summary.schema.md) |

また、以下の raw output ファイルは debugging artifact であり、Kogoto-managed persisted JSON ではない。JSON Schema validation の対象外とする。

```
plan.raw.txt
review-1.raw.txt
review-2.raw.txt
```

---

## 2. Common Policy

全 Kogoto-managed persisted JSON に共通して適用される方針。

### 2.1 JSON Schema draft

JSON Schema Draft 2020-12 を採用する。

Go 用 validator ライブラリ選定、atomic write、migration handling、state read/write の実装は後続 Issue で扱う。

### 2.2 `schema_version`

すべての Kogoto-managed persisted JSON は `schema_version` を required field として持つ。

```json
{ "schema_version": 1 }
```

- `schema_version` は integer
- 各スキーマが独立して `const` 値を管理する。全スキーマが同一バージョンである必要はない
- 破壊的変更（フィールド名変更・削除・型変更）が生じた場合、該当スキーマのみ `const` 値を加算する。他スキーマの `const` は変更しない
- v0 初版ではすべて `const: 1` から開始する
- `schema_version` が存在しないファイルは invalid
- 未知の `schema_version` は `unsupported_schema_version` または `migration_required` として停止する

### 2.3 Unknown fields

Kogoto-managed persisted JSON では unknown field を原則拒否する。

```json
"additionalProperties": false
```

この方針は nested object にも適用する。`writer`、`reviewer`、`review_loop`、`blocked`、`pull_request`、`findings[]`、`human_review_required[]` など、schema 内の各 object にも `additionalProperties: false` を指定する。

### 2.4 汎用 `status` フィールドを使わない

各 JSON ファイルで `status` という同一フィールド名は使用しない。役割を表す名前に分ける。

| ファイル | フィールド名 | 意味 |
|---|---|---|
| `issue-state.json` | `issue_workflow_state` | Issue 全体のワークフロー状態 |
| `run.json` | `run_state` | 個別 Run の内部状態 |
| `plan.json` | `planning_result` | writer による計画段階の結果 |
| `review-N.json` | `reviewer_assessment` | reviewer が自己申告した評価。runner の状態遷移の正本ではない |
| `review-N.json` | `runner_decision` | normalized findings から runner が導出した制御判断 |
| `run-summary.json` | `terminal_state` | succeed / abort / fail 後の終端 Run 状態 |
| `run-summary.json` | `completion_result` | 実行結果の要約 |

### 2.5 Timestamps

日時は RFC3339 UTC timestamp とする。UTC `Z` 表記を要求する。

JSON Schema では `"format": "date-time"` を使い、UTC 必須は semantic rule として文書化する。

### 2.6 JSON Schema と semantic validation の境界

以下は JSON Schema だけで完全に扱わず、semantic validation または文書上の意味論として扱う。

- `run_id` が UUIDv7 として生成されていること
- `worktree` が host OS 上の absolute path であること
- `branch` が Git ref として妥当であること
- `run_state = "blocked"` の場合だけ `blocked` object が存在すること
- `run_state = "failed"` の場合だけ `last_error` object が存在すること
- `terminal_state` と `completion_result` の組み合わせ制約
- `run-summary.json.pull_request` の required / absent 条件
- `run-summary.json.review_rounds` と `run.json.review_loop.completed_review_rounds` の対応関係
- `run-summary.json.failure_summary` が `terminal_state = "failed"` の場合のみ存在すること
- `run-summary.json.completion_result` と `run-summary.json.failure_summary.code` の対応関係
- `run-summary.json.failure_summary` が `run.json.last_error` から安全に転記されていること
- `run-summary.json.failure_summary.artifact_ref` が run-local な artifact reference として妥当であること
- `codex` が known backend だが v0 executable backend ではないこと

---

## 3. Raw output files

agent の raw output は debugging artifact として扱う。

| ファイル | 対応する正規化済み JSON |
|---|---|
| `plan.raw.txt` | `plan.json` |
| `review-N.raw.txt` | `review-N.json` |

- raw output は Kogoto-managed persisted JSON ではなく、schema validation の対象ではない
- raw output だけが存在し対応 `.json` が存在しない場合、parse / validation / normalization に失敗した可能性を示す
- corruption / invalid output の調査に必要な場合は既存 raw output を削除しない

---

## 4. Error classification

### Schema / file-level errors

| Error code | 説明 |
|---|---|
| `json_parse_error` | ファイルを JSON として読めない |
| `schema_validation_error` | JSON だがスキーマに合わない（missing required / invalid enum / unknown field） |
| `unsupported_schema_version` | 未知の `schema_version` |
| `migration_required` | 既知だが現在サポートしていない旧 `schema_version` |
| `corrupted_run_state` | `run.json` が破損しており resume できない状態 |

### `completion_result` / `failure_summary.code` — 失敗分類の対応

`terminal_state = "failed"` の場合、`completion_result` と `failure_summary.code` は以下の対応を持つ。この対応は semantic validation として扱う。

| `completion_result` | `failure_summary.code` |
|---|---|
| `failed_due_to_github_error` | `github_auth_failed`, `issue_not_found`, `issue_closed`, `pr_create_failed` |
| `failed_due_to_git_error` | `main_update_failed`, `worktree_create_failed`, `commit_failed`, `push_failed` |
| `failed_due_to_invalid_state` | `branch_already_exists`, `worktree_path_already_exists`, `nothing_to_commit` |
| `failed_due_to_agent_error` | `writer_launch_failed`, `invalid_writer_output`, `reviewer_launch_failed`, `invalid_reviewer_output` |
| `failed_due_to_verification_error` | `verification_failed` |
| `failed_due_to_runner_error` | `runner_error` |
| `failed_due_to_environment_error` | `environment_error` |

### `last_error.code` — Run failure codes

`run_state = "failed"` の場合、失敗原因は `run.json.last_error.code` で表す。`invalid_agent_output` は `run_state` 値としては使用しない。

| Error code | Phase | 説明 |
|---|---|---|
| `github_auth_failed` | `loading_issue` | GitHub 認証失敗 |
| `issue_not_found` | `loading_issue` | Issue が存在しない |
| `issue_closed` | `loading_issue` | Issue が closed |
| `main_update_failed` | `preparing_worktree` | main ブランチ更新失敗 |
| `worktree_create_failed` | `preparing_worktree` | worktree 作成失敗 |
| `branch_already_exists` | `preparing_worktree` | branch が既存（Kogoto 管理 run 不明） |
| `worktree_path_already_exists` | `preparing_worktree` | worktree path が既存（Kogoto 管理 run 不明） |
| `writer_launch_failed` | `planning` / `writing` | writer agent 起動失敗 |
| `invalid_writer_output` | `planning` / `writing` / `fixing` | writer 出力が不正。planning では `plan.raw.txt` のみ保存し、`plan.json` は書かれない。writing / fixing では `implement.log` または `fix-N.log` を保存し、追加の正規化済み JSON は書かれない |
| `verification_failed` | `testing` | test / lint / typecheck 失敗（retry 上限到達後） |
| `nothing_to_commit` | `committing` | commit 対象変更なし |
| `commit_failed` | `committing` | commit 失敗 |
| `reviewer_launch_failed` | `reviewing` | reviewer agent 起動失敗 |
| `invalid_reviewer_output` | `reviewing` | reviewer 出力が不正。`review-N.raw.txt` のみ保存、`review-N.json` は書かれない |
| `push_failed` | `pushing` | push 失敗 |
| `pr_create_failed` | `creating_pr` | PR 作成失敗 |
| `runner_error` | any | 分類外の runner 例外。上記 code に分類できない runner 内部エラー |
| `environment_error` | any | 分類外の環境エラー。OS / ファイルシステム / ネットワーク等の非分類障害 |

### JSON parse error

JSON として読めない場合:

- 対象ファイルを信用しない
- 自動補修しない
- run を停止する
- 対象ファイルを `*.corrupt.<timestamp>` として退避する
- CLI には recover / manual inspection が必要であることを表示する

### Schema validation error

JSON ではあるがスキーマに合わない場合:

- missing required / invalid enum / unknown field を区別して表示する
- 自動でフィールドを捨てない
- version mismatch は `migration_required` または `unsupported_schema_version` として扱う
- agent output の validation error は `run_state = "failed"` + `last_error.code = "invalid_writer_output"` または `"invalid_reviewer_output"` として停止する

ファイルごとの recovery policy は `*.schema.md` を参照。

---

## 5. v0 Acceptance Criteria

- [ ] `schemas/run.schema.json` が定義されている
- [ ] `schemas/plan.schema.json` が定義されている
- [ ] `schemas/review.schema.json` が定義されている
- [ ] `schemas/run-summary.schema.json` が定義されている
- [ ] すべての Kogoto-managed persisted JSON に `schema_version` が required として定義されている
- [ ] `schema_version` は integer として定義されており、各スキーマが独立して `const` 値を管理している（v0 初版はすべて `const: 1`）
- [ ] nested object を含め、unknown field を原則拒否する方針が明記されている
- [ ] `status` という汎用フィールド名を使わず、`run_state` / `planning_result` / `reviewer_assessment` / `runner_decision` / `terminal_state` / `completion_result` に分離されている
- [ ] JSON Schema で扱う制約と semantic validation で扱う制約が分けて書かれている
- [ ] 各対象ファイルについて JSON Schema / Semantic rules / Recovery policy が分けて書かれている
- [ ] `run.json` の `blocked` object schema と semantic rules が定義されている
- [ ] `human_review_required` の要素型が object array として定義されている
- [ ] `reviewer_assessment` と `runner_decision` の関係が明記されている
- [ ] `findings` / `human_review_required` から `runner_decision` を導出する正規化ルールが明記されている
- [ ] `run-summary.json.pull_request` の required / absent 条件が明記されている
- [ ] `run-summary.json.review_rounds` が `run.json.review_loop.completed_review_rounds` の終端時点の値であることが明記されている
- [ ] `terminal_state` と `completion_result` の対応表が明記されている
- [ ] raw output が Kogoto-managed persisted JSON ではなく debugging artifact であることが明記されている
- [ ] JSON parse error の扱いが明記されている
- [ ] schema validation error の扱いが明記されている
- [ ] corrupted file の退避・停止・再開可否がファイルごとに明記されている
- [ ] `run-summary.json` に詳細ログや機密情報を残さない方針が明記されている
