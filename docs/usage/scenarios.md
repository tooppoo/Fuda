# 利用シナリオ

## シナリオ 1: 基本的なIssue解決フロー

最もシンプルな利用フロー。Issueが明確で、レビュー指摘なしにPRが作成される場合。

### 手順

```bash
# 1. 初回のみ: セットアップ
kogoto setup

# 2. Issue #7 の解決を開始する
kogoto resolve 7
```

Kogotoが自動で以下を実行する:

1. Issue #7 の内容を取得する
2. `kogoto-issue-7` worktreeを作成する
3. writer agentが実装する
4. test / lint / typecheck を実行する
5. commitする
6. reviewer agentがレビューする
7. 指摘なし → PRを作成する
8. PR URLを表示する

```bash
# 3. PRをレビューしてmergeする（人間が行う）

# 4. merge後に終了処理
kogoto close 7
```

---

## シナリオ 2: レビュー指摘ありの修正フロー

reviewer agentが指摘を出し、修正ループが発生する場合。

### 手順

```bash
kogoto resolve 7
```

Kogotoが自動で以下を実行する:

1. Issue #7 の内容を取得する
2. worktreeを作成する
3. writer agentが実装する
4. test / lint / typecheck を実行する
5. commitする（`commit 1: resolve issue #7`）
6. reviewer agentがレビューする
7. `blocking` 指摘あり → writer agentが修正する
8. test / lint / typecheck を再実行する
9. commitする（`commit 2: address review findings round 1`）
10. reviewer agentが再レビューする
11. 指摘なし → PRを作成する
12. PR URLを表示する

```bash
# PRをレビューしてmergeする（人間が行う）

kogoto close 7
```

---

## シナリオ 3: 不明点ありのblocked フロー

writerが作業中に不明点を検出し、一時停止する場合。

### 手順

```bash
kogoto resolve 7
```

Kogotoが以下を実行して停止する:

1. Issue #7 の内容を取得する
2. worktreeを作成する
3. writer agentがplanを作成する
4. planに不明点あり → Issueに質問コメントを投稿して停止する

Issueに投稿されるコメント例:

```markdown
<!-- kogoto:question run=<run-id> issue=7 question=q1 -->

## Kogoto blocked: clarification needed

作業中に次の不明点が見つかりました。

1. Should Kogoto create a PR automatically, or only generate a PR body?

このコメントへの返信、または `kogoto answer 7` で回答してください。
```

### 回答して再開する

```bash
# 回答を投稿する（エディタが開く）
kogoto answer 7

# または回答本文を直接指定する
kogoto answer 7 --body "PRは自動作成してください。PR bodyのみの生成は不要です。"

# runを再開する
kogoto resume 7
```

再開後はKogotoが通常フローを継続し、最終的にPRを作成する。

---

## シナリオ 4: minor 指摘のみでPR作成へ進むフロー

reviewer agentが `minor` 指摘のみを出した場合。

### 手順

```bash
kogoto resolve 7
```

Kogotoが自動で以下を実行する:

1. Issue #7 の内容を取得する
2. worktreeを作成する
3. writer agentが実装する
4. test / lint / typecheck を実行する
5. commitする
6. reviewer agentがレビューする
7. `minor` 指摘のみ → 修正ループには入らない
8. minor findingsをPR本文とrun summaryに記録する
9. PRを作成する
10. PR URLを表示する

v0では、`minor` 指摘はデフォルトで `comment-only` として扱う。
`blocking` / `major` 指摘がある場合のみ修正ループに入る。

---

## シナリオ 5: max_review_loops 到達時の停止フロー

修正ループが上限回数（デフォルト: 3）に達した場合。

### 手順

```bash
kogoto resolve 7
```

3回の修正ループ後も `blocking` / `major` 指摘が残ると、KogotoはIssueにコメントを投稿して停止する。

```markdown
<!-- kogoto:run-status issue=7 run=<run-id> -->

## Kogoto blocked: max review loops reached

修正ループが上限（3回）に達しました。

残存する指摘を確認し、方針を決定してください。

- `kogoto resume 7` で手動再開できます
- Issueに追加の指示を記入して `kogoto resume 7` を実行してください
```

### 人間が確認して再開する

```bash
# 状態を確認する
kogoto status 7

# Issueに追加指示をコメントし、再開する
kogoto answer 7 --body "指摘のr3は許容します。現状のままPRを作成してください。"
kogoto resume 7
```

---

## シナリオ 6: runを中止する

作業を途中で中止する場合。

```bash
kogoto abort 7
```

v0では `abort` はworktree・branch・logを削除しない。状態が `aborted` になるだけで、調査はあとから可能。

### 中止後に状態を確認する

```bash
kogoto status 7
```

---

## シナリオ 7: 初期設定フロー

Kogotoを初めて使う場合の設定手順。

```bash
# 1. 初期設定
kogoto setup
# → GitHub repositoryを設定する
# → GitHub認証状態を確認する
# → writer / reviewer agentを設定する
# → worktree rootを設定する
# → test / lint / typecheckコマンドを設定する

# 2. writerをClaudeに設定する
kogoto writer claude

# 3. reviewerをClaude（code-reviewerサブエージェント）に設定する
kogoto reviewer claude code-reviewer
```

設定は `~/.config/kogoto/config.toml` に保存される。

v0で実行可能なagent backendは `claude` のみ。
`codex` は将来対応予定の既知backendだが、v0では未対応として拒否される。

---

## シナリオ 8: 既存worktreeがある場合の対応

`kogoto resolve 7` 実行時に、Issue #7 のworktreeが既に存在する場合。

| runの状態 | Kogotoの対応 |
|----------|------------|
| active | `kogoto resume 7` を促す |
| blocked | 回答待ちとして表示し、`kogoto answer 7` を促す |
| done / pr-created | `kogoto close 7` を促す |
| 状態不明 | 停止し、人間確認を求める |

Kogotoは既存worktreeを上書きしない。

---

## シナリオ 9: 完全なEnd-to-Endフロー

初回セットアップから終了処理まで。

```bash
# 初回のみ
kogoto setup
kogoto writer claude
kogoto reviewer claude code-reviewer

# Issue解決
kogoto resolve 7
# → Kogotoが自動的に実行し、PR URLを表示する

# 状態確認（任意）
kogoto status 7

# PRをGitHub上でレビューしてmergeする（人間）

# 終了処理
kogoto close 7
# → Issue close, main更新, worktree削除, summary保存
```

`kogoto resolve 7` の1コマンドで、KogotoはIssue取得・worktree作成・実装・レビュー・修正・PR作成まで自動で行う。

`kogoto close 7` の1コマンドで、後片付けをすべて行う。

人間は以下に集中できる:

* Issue作成
* blocked時の回答
* PRの最終レビューとmerge判断
