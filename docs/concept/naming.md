# Kogoto — Fuda からの名称変更方針

Fuda のツール名を `Fuda` から `Kogoto` へ変更する方針を整理する。

---

## 決定

ツール名を `Fuda` から `Kogoto` へ変更することを採用する。

`Kogoto` は「小言」に由来し、Fuda が目指す [Strong HITL](./human-in-the-loop-first.md) の性格を直接表す名称である。

---

## 名称の意味づけ

### Fuda とは

`Fuda`（「札」）は、Issue を「作業のラベル・札」として扱う issue-driven workflow に由来する。各 Issue をひとつの「札」と見なし、その札を処理していく動作を指す。

Issue-driven workflow としての表現としては適切だが、ツールのコア・コンセプトが `Strong HITL` として確立された結果、`Fuda` という名称はツールの本質ではなく workflow artifact 側を指す名称になった。

### Kogoto とは

`Kogoto`（「小言」）は「小うるさい」という意味を持つ。これは、Kogoto が目指す Strong HITL の性格を直接表現している。

Kogoto は「少し小うるさい」runner として:

- **黙って推測しない** — 不明点は人間に問いかけ、推測で先に進まない
- **実行前に確認する** — 重要なステップで止まって確認を求める
- **不確実性を隠さず報告する** — agent の推測を確定判断として提示しない
- **判断が必要な箇所では人間へ戻す** — `blocked` / `human_review_required` で停止
- **重要な判断を Issue / ADR / docs に残す** — 知識を蓄積する

「黙って動く効率的な AI runner」ではなく、「適切な箇所で小うるさく確認する、人間と協調する runner」として位置づける。

---

## Tagline

```
Kogoto — a little fussy runner for the space between human judgment and AI execution.
```

日本語説明:

> Kogoto は、人間の判断と AI の実行のあいだで働く、少し小うるさい AI runner である。

---

## Kogoto と Strong HITL の関係

[Human-in-the-loop first](./human-in-the-loop-first.md) で定義した Strong HITL は次のように整理できる。

> 人間と AI がインクリメンタルに、ステップバイステップで相互作用する。AI は実作業を担うが、思考・判断・評価については人間と協力する。AI は与えられていない権限を自ら取らない。

`Kogoto` という名前は、この Strong HITL の振る舞いを「小うるさい」という一言で直接表現している。

| Kogoto の振る舞い | Strong HITL との対応 |
|---|---|
| 不明点で止まって問いかける | `blocked` フロー — agent は推測で先に進まない |
| 自動判断できない場合に停止する | `human_review_required` — 人間が方針を決定してから再開 |
| 無制限に自己修正を続けない | 検証 retry 上限 — 上限に達したら停止 |
| PR を作成するが merge しない | merge・release 判断は人間が担う |
| 明示的な close コマンドで終了 | run ライフサイクルは人間のコマンドで制御 |

「小言を言う runner」は、ツールとしての完全自律を目指さず、人間との協調を前提とした runner である。

---

## Awai Project との関係

将来的に Awai Project の一部として位置づける場合、次の構成が考えられる。

```
Awai Project
└─ Awai Workflow
   └─ Kogoto
```

`Awai Workflow` は workflow / concept 名、`Kogoto` はその CLI implementation / runner 名として扱う。

---

## Rename の影響範囲

rename を実施する際、以下の箇所が影響を受ける。実装 rename は後続 Issue で行う。

### コード・バイナリ

| 対象 | 現在の値 | 変更後 |
|------|---------|--------|
| CLI binary name | `fuda` | `kogoto` |
| cmd ディレクトリ | `cmd/fuda/` | `cmd/kogoto/` |
| Go module path | `github.com/tooppoo/Fuda` | `github.com/tooppoo/Kogoto` |
| cobra root `Use` | `"fuda"` | `"kogoto"` |
| cobra root `Short` | `"...Fuda..."` 表記 | `"...Kogoto..."` 表記 |

### 設定・状態ファイル

| 対象 | 現在の値 | 変更後 |
|------|---------|--------|
| config file path | `~/.config/fuda/config.toml` | `~/.config/kogoto/config.toml` |
| workspace `branch_prefix` | `fuda/issue-` | `kogoto/issue-` |
| state ディレクトリ | `.fuda/` (想定) | `.kogoto/` |

### GitHub・リポジトリ

| 対象 | 現在の値 | 変更後 |
|------|---------|--------|
| GitHub repository name | `tooppoo/Fuda` | `tooppoo/Kogoto` |
| Go module path (GitHub) | `github.com/tooppoo/Fuda` | `github.com/tooppoo/Kogoto` |
| release artifact names | `fuda-*` | `kogoto-*` |

### ドキュメント

| 対象 | 内容 |
|------|------|
| `README.md` | `Fuda` 表記をすべて `Kogoto` に更新 |
| `docs/` 配下 | `Fuda` 表記をすべて `Kogoto` に更新 |
| `schemas/` 配下 | `Fuda` 表記をすべて `Kogoto` に更新 |
| GitHub Issues / PR 参照 | repository rename 後に更新が必要な箇所を確認 |

### 後方互換性

| 対象 | 検討事項 |
|------|---------|
| config file migration | `~/.config/fuda/` → `~/.config/kogoto/` への移行方針 |
| state file migration | 既存の state を Kogoto が読み込めるか |
| binary alias | `fuda` コマンドを一定期間 `kogoto` へ転送するか |
| release migration guide | ユーザー向け移行ガイドの提供 |

---

## 後続 Issue

rename の実装は以下の後続 Issue に分割して行う。

- [Issue #44](https://github.com/tooppoo/Fuda/issues/44): 実装: CLI binary name / Go module path の rename
- [Issue #45](https://github.com/tooppoo/Fuda/issues/45): 実装: config file / state file の rename と後方互換性設計
- [Issue #46](https://github.com/tooppoo/Fuda/issues/46): ドキュメント: README / docs の `Fuda` → `Kogoto` 表記更新
- [Issue #47](https://github.com/tooppoo/Fuda/issues/47): 運用: GitHub repository name の rename

## 関連

- [Human-in-the-loop first](./human-in-the-loop-first.md) — Kogoto のコアコンセプト
- [GitHub Issue #41](https://github.com/tooppoo/Fuda/issues/41) — 本方針の出発点
- [GitHub Issue #40](https://github.com/tooppoo/Fuda/issues/40) — Strong HITL コンセプトの確立
