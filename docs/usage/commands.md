# コマンドリファレンス

## 概要

Kogotoは `kogoto` コマンドを通じて操作する。

```
kogoto <subcommand> [arguments] [options]
```

---

## `kogoto setup`

初期設定を行う。

```bash
kogoto setup
```

設定する内容:

* 対象GitHub repositoryの指定
* GitHub認証状態の確認
* writer agentの設定
* reviewer agentの設定
* worktree rootの設定
* test / lint / typecheck コマンドの設定

設定ファイルの保存先: `~/.config/kogoto/config.toml`

---

## `kogoto writer`

writer agentを設定する。

```bash
kogoto writer claude
```

writerは、Issueに基づいて変更を書くagentである。

| subcommand | agent | v0での扱い |
|------------|-------|------------|
| `claude` | Claude Code | 設定可能 |
| `codex` | OpenAI Codex CLI | 既知だが未対応。v0では明示的に拒否する |

v0では実行可能なwriter agentは `claude` のみ。
`codex` は将来対応予定のagentとして認識するが、v0では設定・実行できない。
未知のagent名は、`codex` のような既知未対応agentとは区別して拒否する。

---

## `kogoto reviewer`

reviewer agentを設定する。

```bash
kogoto reviewer claude
kogoto reviewer claude <reviewer-name>
```

reviewerは、writerの変更を検査するagentである。

| 形式 | 説明 |
|------|------|
| `kogoto reviewer claude` | Claude Codeをreviewer agentに設定する |
| `kogoto reviewer claude <reviewer-name>` | Claude Code側のreviewer subagent名を指定する |

例:

```bash
kogoto reviewer claude code-reviewer
```

---

## `kogoto resolve <issue-number>`

Issue解決のためのrunを開始する。

```bash
kogoto resolve 7
```

実行内容:

1. GitHub Issue #7を取得する
2. Issue専用worktreeを作成する
3. writer agentにIssue内容を渡して作業させる
4. 変更を検証し、commitする
5. reviewer agentに差分レビューさせる
6. レビュー指摘があればwriter agentに修正させる
7. 修正後に再検証・再commit・再レビューする
8. `blocking` / `major` 指摘がなくなるまで修正ループを回す
9. 指摘がない、または `minor` 指摘のみになったらPRを作成する
10. 作成したPR URLを表示する

v0では `minor` 指摘だけでは修正ループを起動しない。
`minor` 指摘はPR本文とrun summaryに記録する。
`human_review_required` が空でない場合は、安全側に倒してrunを停止する。

---

## `kogoto status`

runの状態を表示する。

```bash
kogoto status
kogoto status 7
```

| 形式 | 説明 |
|------|------|
| `kogoto status` | 現在進行中のrunの状態を表示する |
| `kogoto status 7` | Issue #7 のrunの状態を表示する |

---

## `kogoto answer <issue-number>`

blocked状態になったIssueに対して、ユーザー回答をIssueコメントとして投稿する。

```bash
kogoto answer 7
kogoto answer 7 --body "回答本文"
```

| オプション | 説明 |
|-----------|------|
| `--body` | 回答本文を直接指定する。省略した場合はエディタが開く |

---

## `kogoto resume <issue-number>`

中断中のrunを再開する。

```bash
kogoto resume 7
```

`answer` 後の再開や、手動で中断したrunを再開する場合に使う。

---

## `kogoto abort <issue-number>`

runを中止する。

```bash
kogoto abort 7
```

v0では、`abort` はworktree・branch・logを削除しない。調査可能性を残すため、状態を `aborted` にするだけとする。

---

## `kogoto close <issue-number>`

作業完了後の終了処理を行う。

```bash
kogoto close 7
```

実行内容:

1. Issueがopenならcloseする（closedならskip）
2. mainを更新する
3. 対象worktreeを削除する
4. `git worktree prune` を実行する
5. 中間ファイルを削除する
6. summaryを保存する
7. run状態をsucceededにする

### 安全条件

以下の場合、`kogoto close` は停止する:

* 対応PRが存在しない
* 対応PRが未mergeである
* worktreeに未commit変更がある
* worktree branchに未push commitがある
* run状態が不明である
* 対象worktreeがKogoto管理下であることを確認できない

---

## 設定ファイル

### `~/.config/kogoto/config.toml`

```toml
[github]
repo = "tooppoo/Kogoto"
default_base = "main"

[workspace]
root = "~/src/kogoto-worktrees"
branch_prefix = "kogoto/issue-"
worktree_name_template = "{repo}-issue-{issue_number}"

[agents.writer]
type = "claude"

[agents.reviewer]
type = "claude"
subagent = "code-reviewer"

[commands]
test = ["yarn test"]
lint = ["yarn lint"]
typecheck = ["yarn typecheck"]

[verification]
max_retries = 2

[review]
max_loops = 3
fail_on = ["blocking", "major"]
minor_policy = "comment-only"

[commit]
strategy = "checkpoint"

[pr]
create = true
draft = false

[close]
keep_summary = true
delete_intermediate_files = true
require_merged_pr = true
```

### 設定項目

| セクション | キー | 説明 |
|-----------|------|------|
| `github` | `repo` | 対象repositoryの `owner/name` |
| `github` | `default_base` | PRのベースブランチ（通常 `main`） |
| `workspace` | `root` | worktreeを作成するルートディレクトリ |
| `workspace` | `branch_prefix` | 作業ブランチのプレフィックス |
| `workspace` | `worktree_name_template` | worktree名のテンプレート |
| `agents.writer` | `type` | writer agentの種別。v0で実行可能なのは `claude` のみ |
| `agents.reviewer` | `type` | reviewer agentの種別。v0で実行可能なのは `claude` のみ |
| `agents.reviewer` | `subagent` | Claude reviewer subagent名 |
| `commands` | `test` | テスト実行コマンド |
| `commands` | `lint` | lint実行コマンド |
| `commands` | `typecheck` | 型チェック実行コマンド |
| `verification` | `max_retries` | verification retryの上限回数（初期値: `2`）。run全体に対する上限。`0` は初回verification失敗時に即 `failed` で停止する。整数で `0` 以上を指定すること |
| `review` | `max_loops` | 修正ループの上限回数（初期値: 3） |
| `review` | `fail_on` | 修正ループを発動するseverity |
| `review` | `minor_policy` | `minor` 指摘の扱い（`comment-only` など） |
| `commit` | `strategy` | commitの戦略（`checkpoint` など） |
| `pr` | `create` | PR自動作成の有効/無効 |
| `pr` | `draft` | draftPRとして作成するか |
| `close` | `keep_summary` | summaryを保持するか |
| `close` | `delete_intermediate_files` | 中間ファイルを削除するか |
| `close` | `require_merged_pr` | PR未mergeで停止するか |
