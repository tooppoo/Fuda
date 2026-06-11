# Agent Adapter Contract

## Scope

このドキュメントは `internal/agent` パッケージが runner に提供する入出力契約を定義する。

対象は次の境界である。

- writer / reviewer role の Go interface
- agent adapter に渡す input
- agent adapter から返す result
- Claude Code v0 adapter の CLI 起動形式
- prompt template の必須構成
- stdout / stderr / raw output / JSON 抽出の扱い
- agent 起動失敗・出力不正・中断・timeout の扱い

正規化済み永続 JSON の schema は `schemas/` を正とする。このドキュメントでは schema の全 field を再定義しない。

---

## 基本方針

- v0 で実行可能な backend は `claude` のみとする。
- `codex` は known backend だが v0 executable backend ではない。設定や run state 上では既知値として扱えるが、実行要求は semantic validation で拒否する。
- runner は agent backend 固有の CLI 引数・stdout 形式・stderr 形式を知らない。これらは adapter が吸収する。
- adapter は raw output を失わず runner に返す。永続化の責務は runner / state 層が持つ。
- adapter は agent の自然文を直接 workflow 判断に使わない。workflow 判断に使う値は、parse / schema validation / normalization 済みの構造化 result に限る。
- agent が不明点を返した場合、runner は推測で継続せず `blocked` または `human_review_required` へ進む。

---

## Go Interface

`internal/agent` は role ごとに interface を分ける。

```go
type Writer interface {
	Plan(ctx context.Context, input PlanInput) (PlanResult, error)
	Write(ctx context.Context, input WriteInput) (WriteResult, error)
	Fix(ctx context.Context, input FixInput) (FixResult, error)
}

type Reviewer interface {
	Review(ctx context.Context, input ReviewInput) (ReviewResult, error)
}
```

### Interface の責務

| Method | Phase | Responsibility |
|---|---|---|
| `Writer.Plan` | `planning` | Issue と既存 context を読み、作業可能性・task list・blocking question を返す |
| `Writer.Write` | `writing` | plan に基づいて worktree 内のファイルを変更し、実行概要を返す |
| `Writer.Fix` | `fixing` | verification failure または reviewer finding に基づいて worktree 内の変更を修正する |
| `Reviewer.Review` | `reviewing` | diff・test result・Issue scope を確認し、findings / human review request を返す |

adapter は GitHub Issue 取得、worktree 作成、test / lint / typecheck、commit、PR 作成を行わない。これらは runner の責務である。

---

## Backend Support

| Backend | v0 behavior |
|---|---|
| `claude` | executable |
| `codex` | known-but-unsupported。実行要求は user-facing validation error として拒否する |
| other | unknown。設定段階で user-facing validation error として拒否する |

unsupported backend / unknown backend は agent process の launch failure ではない。run 開始後にこの状態が検出された場合は、`last_error.code = runner_error` として停止する。

---

## Common Input

すべての agent input は次の情報を持つ。

| Field | Meaning |
|---|---|
| `RunID` | 対象 run の ID |
| `Repository` | `owner/name` 形式の repository |
| `IssueNumber` | 対象 GitHub Issue number |
| `Issue` | Issue title / body / relevant comments を含む snapshot |
| `Worktree` | agent process の working directory |
| `Branch` | Kogoto が作成した作業 branch |
| `RunState` | 現在の `run_state` |
| `ArtifactsDir` | run artifact を置く directory |
| `Now` | adapter invocation 時点の UTC timestamp |

`Issue` は runner が取得して adapter に渡す。adapter は tracker API を直接呼ばない。

### `PlanInput`

| Field | Meaning |
|---|---|
| `Common` | common input |
| `RepositorySummary` | 任意。runner が事前に集めた repository 概要 |
| `ExistingPlan` | resume / re-plan 時のみ。通常は absent |
| `HumanAnswers` | blocked から再開する場合の回答一覧 |

### `WriteInput`

| Field | Meaning |
|---|---|
| `Common` | common input |
| `Plan` | 正規化済み `plan.json` 相当の plan |
| `HumanAnswers` | blocked から再開する場合の回答一覧 |

### `FixInput`

| Field | Meaning |
|---|---|
| `Common` | common input |
| `Plan` | 正規化済み plan |
| `Reason` | `verification_failed` または `review_findings` |
| `VerificationFailure` | verification failure による修正時のみ。command / exit code / log artifact ref / summary |
| `Review` | reviewer finding による修正時のみ。正規化済み `review-N.json` 相当 |
| `HumanAnswers` | human review 後に再開する場合の回答一覧 |

### `ReviewInput`

| Field | Meaning |
|---|---|
| `Common` | common input |
| `Plan` | 正規化済み plan |
| `ReviewNumber` | 次に作成する review number |
| `DiffSummary` | runner が取得した diff summary |
| `Diff` | review 対象 diff。大きすぎる場合は artifact ref と要約 |
| `Verification` | 直近の test / lint / typecheck 結果 |
| `PreviousReviews` | 必要な場合のみ。過去の正規化済み review |

---

## Common Result

すべての result は次の情報を持つ。

| Field | Meaning |
|---|---|
| `Backend` | 実行した backend。v0 では `claude` |
| `RawOutput` | agent stdout から得た raw response。parse 失敗時も保持する |
| `Stderr` | stderr summary または artifact ref。workflow 判断には使わない |
| `ExitCode` | process exit code |
| `StartedAt` / `FinishedAt` | UTC timestamp |
| `Duration` | 実行時間 |

`RawOutput` は debugging artifact として保存するための値である。`run-summary.json` には含めない。

### `PlanResult`

`PlanResult` は正規化済み plan を持つ。

| Field | Meaning |
|---|---|
| `Plan` | `schemas/plan.schema.json` に適合する normalized plan |

`Plan.PlanningResult = blocked_by_ambiguity` の場合、runner は `run_state = blocked` に遷移し、Issue に質問コメントを投稿する。

### `WriteResult`

`WriteResult` は writer 実行結果を表す。

| Field | Meaning |
|---|---|
| `Completed` | writer が作業を完了した場合 true |
| `Summary` | 変更概要 |
| `ChangedFiles` | writer が報告した変更ファイル一覧。正本は git diff |
| `Blocked` | 継続不能な不明点がある場合のみ |

`ChangedFiles` は参考情報である。runner は必ず git status / diff で実際の変更を確認する。

### `FixResult`

`FixResult` は `WriteResult` と同じ構造を使う。`Blocked` が存在する場合、runner は `run_state = blocked` に遷移する。

### `ReviewResult`

`ReviewResult` は正規化済み review を持つ。

| Field | Meaning |
|---|---|
| `Review` | `schemas/review.schema.json` に適合する normalized review |

`runner_decision` は adapter または runner の normalization step で導出してよい。ただし導出ルールは `schemas/review.schema.md` を正とし、runner が最終検証する。

---

## Claude Adapter Invocation

v0 の Claude adapter は Claude Code の print mode を使う。

### 基本形式

```bash
claude -p \
  --output-format json \
  --json-schema '<schema>' \
  --no-session-persistence \
  --allowed-tools '<tools>' \
  '<prompt>'
```

`--json-schema` が使える場合、adapter は method ごとの expected output schema を渡す。Claude Code の JSON response から `structured_output` を抽出し、schema validation / normalization を行う。

`--json-schema` が利用できない環境では、adapter は `--output-format json` の `result` から JSON object を抽出してよい。ただしこの fallback は v0 では degraded mode とし、JSON 抽出失敗時は invalid output として扱う。

### Tool Permission

Claude Code に渡す tool permission は method ごとに分ける。

| Method | Tool policy |
|---|---|
| `Plan` | read-only。`Read` と repository inspection に必要な安全な `Bash` のみ |
| `Write` | `Read`, `Edit`, `Write` と repository inspection に必要な安全な `Bash` |
| `Fix` | `Read`, `Edit`, `Write` と repository inspection に必要な安全な `Bash` |
| `Review` | read-only。`Read` と diff / status inspection に必要な安全な `Bash` のみ |

v0 adapter は `--dangerously-skip-permissions` を使わない。

`Bash` permission は必要な command prefix に限定する。例:

```text
Bash(git status *),Bash(git diff *),Bash(go test *),Bash(make test *)
```

runner が verification を実行するため、writer に広い test command 実行権限を与える必要はない。writer / reviewer が確認のために test を実行する場合も、許可済み command に限定する。

### Subagent 指定

reviewer に Claude subagent が指定されている場合、adapter は prompt の先頭で対象 subagent の利用を明示する。

```text
Use the configured Claude Code reviewer subagent: <subagent-name>.
```

Claude Code の subagent 起動方法が CLI flag として安定して利用できる場合は、adapter 実装でその flag に置き換えてよい。ただし runner に subagent 固有の CLI 仕様を漏らさない。

### Working Directory

agent process の working directory は必ず `input.Worktree` とする。

agent に repository root 以外の directory で作業させてはならない。runner は process 起動前に `input.Worktree` が Kogoto 管理 worktree であることを検証する。

### Environment Variables

adapter は環境変数を最小限にする。

| Variable | Policy |
|---|---|
| `PATH` | claude executable の解決に必要な値を継承する |
| `HOME` | Claude Code の認証・設定に必要なため継承可 |
| `CLAUDE_CONFIG_DIR` | 設定で指定された場合のみ設定する |
| `ANTHROPIC_API_KEY` | 呼び出し元環境に存在する場合のみ継承する。ログや summary に出さない |
| `KOGOTO_RUN_ID` | adapter が設定してよい |
| `KOGOTO_ISSUE_NUMBER` | adapter が設定してよい |

token / credential / API key は prompt、raw output、log summary、run summary に含めてはならない。

### Timeout

adapter invocation には timeout を設定する。

| Method | Default timeout |
|---|---|
| `Plan` | 10 minutes |
| `Write` | 60 minutes |
| `Fix` | 45 minutes |
| `Review` | 20 minutes |

timeout は設定で上書きできるが、実行時に採用した値は log に記録する。timeout 発生時は process を終了し、対応する launch / execution error として runner に返す。

### stdout / stderr

| Stream | Policy |
|---|---|
| stdout | Claude Code JSON response として parse する。raw response は artifact 用に保持する |
| stderr | diagnostic として保持する。JSON 抽出元にしない |

stderr に JSON らしい文字列が含まれていても workflow 判断には使わない。

---

## JSON Extraction

### Preferred path

`--output-format json --json-schema` を使った場合:

1. stdout 全体を Claude Code response JSON として parse する。
2. `structured_output` を取り出す。
3. method ごとの expected schema に対して validation する。
4. Kogoto normalized schema に変換する。
5. normalized result を再度 `schemas/*.schema.json` に対して validation する。

### Fallback path

`structured_output` がない場合:

1. stdout 全体を Claude Code response JSON として parse する。
2. `result` を text として取り出す。
3. `result` 内から最初の JSON object を抽出する。
4. markdown fence がある場合は fence 内だけを候補にする。
5. 複数の JSON object がある場合は invalid output とする。
6. method ごとの expected schema に対して validation する。

fallback path は自然文の解釈で補正してはならない。欠落 field、未知 field、不正 enum は invalid output とする。

---

## Prompt Templates

prompt は method ごとに生成する。各 prompt は次の構造を持つ。

```text
<role instruction>

<Kogoto workflow rules>

<repository context>

<issue context>

<method-specific input>

<required output contract>
```

### 共通 workflow rules

すべての prompt に次を含める。

```text
You are running inside Kogoto, a human-in-the-loop-first issue runner.

Follow the GitHub Issue scope. Do not expand scope without an explicit human answer.
If the task is ambiguous or unsafe to continue, return the structured blocked output instead of guessing.
Do not merge pull requests, push to main, delete unrelated files, or perform destructive cleanup.
Return only the requested structured output.
```

### Writer Plan Template

```text
Role: writer planner.

Read the Issue context and decide whether the work is ready to implement.
Return a plan with tasks if ready.
If acceptance criteria, target files, or expected behavior are ambiguous, return blocked_by_ambiguity with concrete questions.

Output must match the PlanResult contract.
```

### Writer Write Template

```text
Role: writer.

Implement the approved plan in the current worktree.
Keep changes inside the Issue scope.
Run only commands needed to inspect or edit the repository.
If you cannot continue safely, return blocked with questions.

Output must match the WriteResult contract.
```

### Writer Fix Template

```text
Role: writer fixing a previous attempt.

Use the provided verification failure or reviewer findings as the required correction target.
Do not introduce unrelated refactors.
If the requested fix conflicts with the Issue scope or requires a human decision, return blocked with questions.

Output must match the FixResult contract.
```

### Reviewer Template

```text
Role: reviewer.

Review the diff against the Issue scope, plan, and verification result.
Do not edit files.
Report findings as structured data.
Use human_review_required for decisions that require human judgment.

Output must match the ReviewResult contract.
```

---

## Error Handling

adapter error は runner が `run_state` と `last_error.code` に変換する。

| Case | Adapter error | Runner error code |
|---|---|---|
| executable not found | `LaunchFailed` | `writer_launch_failed` or `reviewer_launch_failed` |
| process start failed | `LaunchFailed` | `writer_launch_failed` or `reviewer_launch_failed` |
| timeout | `ExecutionTimeout` | `writer_launch_failed` or `reviewer_launch_failed` |
| context canceled by user | `ExecutionCanceled` | `aborted` if user initiated; otherwise `runner_error` |
| exit code non-zero with no parseable valid output | `ExecutionFailed` | `writer_launch_failed` or `reviewer_launch_failed` |
| stdout is not Claude JSON | `InvalidOutput` | `invalid_writer_output` or `invalid_reviewer_output` |
| `structured_output` / extracted JSON is schema-invalid | `InvalidOutput` | `invalid_writer_output` or `invalid_reviewer_output` |
| natural language only | `InvalidOutput` | `invalid_writer_output` or `invalid_reviewer_output` |
| multiple JSON objects in fallback extraction | `InvalidOutput` | `invalid_writer_output` or `invalid_reviewer_output` |

`Plan` / `Write` / `Fix` の invalid output は `invalid_writer_output` とする。
`Review` の invalid output は `invalid_reviewer_output` とする。

### exit code != 0

exit code が non-zero の場合でも stdout に valid structured output がある場合、adapter はその output を返してよい。ただし stderr / exit code は result に保持する。

stdout に valid structured output がない場合、launch / execution failure として扱う。自然文 error message を推測で blocked / finding に変換してはならない。

### agent が自然文だけ返した場合

自然文だけの output は invalid output である。

- writer: `run_state = failed` + `last_error.code = invalid_writer_output`
- reviewer: `run_state = failed` + `last_error.code = invalid_reviewer_output`

raw output は保存するが、`plan.json` / `review-N.json` は書かない。

`Plan` の invalid output では `plan.raw.txt` のみ保存し、`plan.json` は書かない。
`Write` の invalid output では `implement.log` を保存し、追加の正規化済み JSON は書かない。
`Fix` の invalid output では `fix-N.log` を保存し、追加の正規化済み JSON は書かない。
`Review` の invalid output では `review-N.raw.txt` のみ保存し、`review-N.json` は書かない。

### agent 実行中断

user が明示中断した場合、runner は `run_state = aborted` に遷移する。

OS signal、親 process 終了、context cancellation など user intent が判定できない中断は `runner_error` または launch failure として扱う。adapter は中断理由を error detail に含める。

---

## Artifact Mapping

| Method | Raw output artifact | Normalized artifact |
|---|---|---|
| `Plan` | `plan.raw.txt` | `plan.json` |
| `Write` | `implement.log` | none |
| `Fix` | `fix-N.log` | none |
| `Review` | `review-N.raw.txt` | `review-N.json` |

`Write` / `Fix` は worktree の変更そのものが成果物であり、正規化済み JSON artifact は持たない。runner は git diff / status を正本として扱う。

---

## Runner Responsibilities

runner は adapter result を受け取った後に次を行う。

- raw output / log artifact を保存する
- normalized JSON を schema validation してから保存する
- `run.json.run_state` を状態機械に従って更新する
- verification command を実行する
- git diff / status / commit を管理する
- reviewer output から最終 `runner_decision` を検証する
- blocked / human review required の Issue comment を投稿する
- `last_error` を永続化する

adapter は runner state を直接書き換えない。

---

## 既存仕様との接続

| Spec | Connection |
|---|---|
| `docs/internal/domain-model.md` | writer / reviewer role の意味 |
| `docs/internal/state-machine.md` | agent result 後の `run_state` 遷移 |
| `docs/internal/error-handling.md` | `last_error.code` と raw output 保存方針 |
| `docs/internal/log-spec.md` | artifact file layout |
| `schemas/plan.schema.md` | normalized plan の semantic rules |
| `schemas/review.schema.md` | normalized review と `runner_decision` 導出 |
| `schemas/run.schema.md` | backend enum と v0 executable backend policy |
