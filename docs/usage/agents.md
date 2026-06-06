# 対応Agent

## 概要

Fudaは writer と reviewer の2種類のagent roleを持つ。
v0では `claude` のみを実行可能なagent backendとして扱う。
`codex` は将来対応予定の既知backendとして認識するが、v0では設定・実行できない。

---

## Agent の Role

### Writer

Issueに基づいて変更を書くagent。

writerは以下を扱う:

* コード実装
* ドキュメント
* 設定ファイル
* テスト
* CI設定
* その他リポジトリ内成果物

writerは不明点・曖昧点があれば推測で進めず、blocked状態を構造化して返す。

### Reviewer

writerが作成した変更を検査するagent。

reviewerは以下を確認する:

* 差分の内容
* テスト結果
* Issueの受け入れ基準への適合
* scope逸脱
* 設計上の問題

reviewerは原則として変更を書かず、指摘を構造化して返す（参照: [コマンドリファレンス](commands.md)のreviewer出力形式）。

---

## v0 対応Agent 種別

| 種別 | 識別子 | 状態 |
|------|--------|------|
| Claude Code | `claude` | v0で実行可能 |
| OpenAI Codex CLI | `codex` | 既知だがv0では未対応 |

### Agent support policy

v0では `claude` のみを設定可能なagent backendとする。

`codex` は将来対応予定の既知backendだが、v0では明示的に拒否する。
未知のagent名は、既知だが未対応のagentとは区別して拒否する。

```txt
agent=claude
→ accepted

agent=codex
→ rejected as unsupported-in-v0

agent=gemini
→ rejected as unknown-agent
```

`codex` を指定した場合のメッセージ例:

```txt
codex is a planned agent backend, but it is not supported in v0.
Supported agents in v0: claude.
```

未知のagent名を指定した場合のメッセージ例:

```txt
unknown agent: gemini.
Supported agents in v0: claude.
Planned but unsupported: codex.
```

### 実装優先度

1. `claude` writer
2. `claude` reviewer
3. `claude` reviewer subagent指定
4. `codex` reviewer
5. `codex` writer

v0.1の最小到達点は `claude` writer + `claude` reviewer。

---

## Claude Code

### 設定

```bash
fuda writer claude
fuda reviewer claude
```

reviewer でsubagentを指定する場合:

```bash
fuda reviewer claude code-reviewer
```

### 設定ファイル

```toml
[agents.writer]
type = "claude"

[agents.reviewer]
type = "claude"
subagent = "code-reviewer"  # 省略可
```

### Subagentの指定

`fuda reviewer claude <reviewer-name>` では、Claude Code側のreviewer subagent名を指定できる。

指定したsubagentがレビューを担当する。省略した場合はデフォルトのClaude Codeがレビューを行う。

---

## Codex CLI

v0では対応しない。
将来バージョン（v0.5以降）での対応を予定する。

v0の実装では、`codex` をunknown agentとして扱わない。
既知だが未対応のagentとして分類し、unsupported-in-v0として拒否する。

### Codex対応の前提条件

Codex対応は以下を安定して扱えることを確認したうえで段階的に進める:

* CLI入出力仕様
* JSON出力
* exit code
* 権限制御

### 設定（将来）

```bash
fuda writer codex
fuda reviewer codex
```

---

## Agent 非依存の設計方針

Fudaは特定のagentに依存しない設計を採用している。

* `internal/agent` パッケージがwriter / reviewer roleをインタフェースとして定義する
* 各agent種別（Claude, Codex等）はそのインタフェースの実装として追加する
* 上位層（`runner`, `app`）はインタフェースにのみ依存し、特定実装を参照しない

将来のagent追加時は、`internal/agent` に新しいアダプタを実装することで対応できる。
