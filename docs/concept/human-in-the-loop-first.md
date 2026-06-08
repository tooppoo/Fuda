# Human-in-the-loop first

Kogoto のコア・コンセプトである `Human-in-the-loop first` の詳細を定義する。

---

## 概要

Kogoto は `Human-in-the-loop first` を第一原理として設計されている。

これは「最後に人間が承認する」という意味ではない。人間が scope・判断・方向性を、すべての重要なステップにおいて保持し続けることを意味する。

Kogoto は、AIに作業を完全委任するための自律開発ツールではない。開発ループの反復部分を自動化しつつ、scope・判断・レビュー・release control を人間の手元に残すことを目的とする。

---

## 弱い HITL と強い HITL

Kogoto は `Human-in-the-loop` を二種類に区別する。

| 種類 | 説明 |
|---|---|
| **弱い HITL** | AI が大部分の作業を実行し、人間は要所で承認するだけ。人間は gatekeeper として配置され、AI の実行結果を事後確認する |
| **強い HITL** | 人間と AI がインクリメンタルに、ステップバイステップで相互作用する。AI は実作業を担うが、思考・判断・評価については人間と協力する。AI は与えられていない権限を自ら取らない |

Kogoto が目指すのは **強い HITL** である。

強い HITL において、agent は推測・整理・提案を行うことができる。しかし独断はしない。判断が必要な場面では run を停止し、人間に問いかける。Kogoto が判断を人間の手に残すのは、AI が推測できないからではない。自動化や AI による代替よりも、人間と agent の相互作用を通じて判断を形成することに価値を置くからだ。

---

## 人間が保持する判断領域

Kogoto において、次の判断は人間が行う。

| 領域 | 説明 |
|---|---|
| **Issue scope** | Issue が扱う範囲・扱わない範囲・scope 変更の可否 |
| **Blocked questions への回答** | writer agent が計画・実装中に検出した不明点への回答。agent は推測で進まず停止して待つ |
| **reviewerによる意思決定要求** | reviewer agent が自動判断できないと判定した場合の意思決定（継続・修正・現状承認） |
| **Merge と release control** | Kogoto は PR を作成するが、merge・main branch への反映は行わない |
| **Abort / resume / close** | run のライフサイクルは明示的なコマンドで人間が制御する |

---

## Agent の役割

Writer agent と reviewer agent は、開発ループの反復作業を担う。

- Writer agent: Issue scope に従い、実装・修正・ドキュメント作成を行う
- Reviewer agent: diff・テスト結果・受け入れ基準を検査する
- 両 agent は、検討内容の要約・判断根拠・不確実性を報告する。未確定事項を隠蔽したり、推測を確定判断として提示したりしない

Agent は人間の判断を支援する。置き換えない。

---

## MVP v0 との接続

`Human-in-the-loop first` は Kogoto に後付けされた安全装置ではなく、設計の第一原理である。MVP v0 の次の仕様がこれを直接表している。

### Blocked flow

writer agent が計画・実装中に不明点を検出した場合、run は停止する。Kogoto は Issue に質問コメントを投稿し、人間の回答を待つ。agent は推測で先に進まない。

```
run_state: blocked
→ Kogoto posts question to Issue
→ human answers via kogoto answer / Issue comment
→ kogoto resume → run continues
```

参照: [Run State Machine](../internal/state-machine.md), [利用シナリオ - blocked フロー](../usage/scenarios.md)

### 自動判断が困難な場合の停止

reviewer agent が人間判断を必要とする論点を検出した場合、run は `human_review_required` として停止する。自動的に続行しない。人間が方針を決定してから再開する。

参照: [Run State Machine](../internal/state-machine.md)

### 不安定時の自動継続を避ける

検証が不安定な状態のまま自動で走り続けることより、立ち止まることに価値を置く。修正サイクルには上限があり、上限に達した場合は run が停止する。Kogoto は無制限の自己修正を許容しない。

参照: [Run State Machine](../internal/state-machine.md)

### PR merge・main branch 直接 push を行わない

Kogoto は PR を作成するが、merge は行わない。main branch への直接 push も行わない。merge・release の判断と実行は人間が担う。

### 明示的な close と cleanup

run 完了後の後片付け（Issue close, worktree 削除, summary 保存）は、`kogoto close` という明示的なコマンドで行う。Kogoto は自動で cleanup を実行しない。

---

## 強い HITL と知識の蓄積

強い HITL は、人間と agent の相互作用を一過性にしない。Kogoto では、相互作用の過程を記録し、後続の判断資源として蓄積することを方針とする。

Kogoto における知識の流れは次のように整理する。

```
判断過程: Issue comments
判断記録: ADR
現在仕様: docs
```

| 位置 | 役割 |
|---|---|
| **Issue** | scope container — 何を扱うか・どこまで扱うかを定義する |
| **Issue comments** | deliberation log — 人間と agent の質問・回答・検討・暫定判断を記録する |
| **ADR** | decision record — 後続に影響する重要判断を、理由・代替案・影響とともに記録する |
| **docs** | stabilized project knowledge — ADR によって決まった方針を、現在の仕様・利用方法として反映する |

### ADR と docs の役割分担

ADR と docs は役割を分ける。

- ADR は「なぜその判断をしたか」を記録する
- docs は「その判断の結果、現在どう扱うべきか」を記述する

すべての判断を ADR 化するわけではない。ADR は、後続 Issue・実装・設計・agent の判断に影響する重要決定に限定する。

ADR が必要な判断の例:
- 複数の選択肢があり、どれを選ぶかで後続設計に影響する
- 一度決めると戻しにくい
- Issue / PR をまたいで参照される
- agent が今後同じ判断を繰り返す可能性がある

ADR が不要な判断の例:
- 明白な typo 修正
- 局所的な refactor
- 既存方針に従っただけの実装

> **Note**: ADR のファイル命名規則・テンプレート・保存ディレクトリの確定・作成支援 CLI は、この概念文書の対象外である。[Issue #43](https://github.com/tooppoo/Kogoto/issues/43) で扱う。

---

## 自律 agent platform との差別化

Kogoto は、AI が自律的に Issue を消化し続ける自律開発 platform を目指さない。

自律実行を重視する agent runner や orchestration tool は、作業の継続実行や自動化に重点を置く。Kogoto はその方向とは異なり、個々の Issue に対する人間・agent の段階的相互作用と判断記録を中心に置く。

Kogoto の設計の軸:

- 完全自律ではなく、人間との協調を設計の中心に置く
- 個々の Issue を明示的に制御された単位として扱う
- 人間と agent が step-by-step に相互作用し、その過程を記録する
- AI の実行速度より、人間と agent の相互作用を通じた段階的な判断形成を優先する

Kogoto が提供するのは、AI が速く動くための runway ではなく、人間が制御を保ちながら AI の力を借りるための runner である。
