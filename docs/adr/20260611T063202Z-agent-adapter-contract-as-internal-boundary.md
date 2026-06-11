# ADR: Agent adapter contract を内部境界として定義する

Status: Accepted
DateTime: 2026-06-11T06:32:02Z

## 背景

Issue #10 で、`docs/internal/agent-contract.md` として agent adapter の内部契約を文書化する必要が出た。

Kogoto は writer agent / reviewer agent を扱うが、runner の責務は agent framework そのものになることではない。runner は Issue、worktree、状態遷移、verification、commit、review loop、PR 作成を制御する。一方で agent backend ごとの CLI 仕様、prompt、stdout / stderr、structured output の取り扱いは、runner に直接漏らすと変更に弱くなる。

既存仕様では次の方針がすでにある。

- v0 で実行可能な backend は `claude` のみ
- `codex` は known backend だが v0 executable backend ではない
- `plan.json` / `review-N.json` は raw output ではなく normalized artifact
- raw output は debugging artifact として保存する
- invalid writer / reviewer output は `run_state = failed` と `last_error.code` で表す

これらと矛盾しない形で、agent adapter と runner の境界を固定する必要がある。

## 決定

`internal/agent` の入出力契約を、runner / schema / log とは分離した内部境界として定義する。

具体的には、次の方針を採用する。

- writer / reviewer role ごとに interface を分ける
- runner は backend 固有の CLI 引数、stdout envelope、stderr、subagent 起動方法を知らない
- v0 の executable backend は Claude Code adapter とする
- Claude Code adapter は print mode と structured output を使う
- agent output は raw output と normalized result に分ける
- workflow 判断には normalized result だけを使う
- `plan.json` / `review-N.json` の schema は `schemas/` を正とし、agent contract はそれらを再定義しない
- `Write` / `Fix` は worktree の変更を成果物とし、追加の normalized JSON artifact は持たない
- invalid output は既存 error code に接続する
  - writer: `invalid_writer_output`
  - reviewer: `invalid_reviewer_output`
- backend unsupported / unknown は agent process の launch failure ではなく、設定または semantic validation の失敗として扱う

## 検討した代替案

### runner が backend 固有仕様を直接扱う

runner が Claude Code の CLI 引数、stdout 形式、subagent 指定、JSON 抽出を直接扱う案。

初期実装は短くなるが、runner が backend 実装に密結合する。将来 Codex など別 backend を追加するとき、runner の状態遷移や error handling に backend 差分が混入しやすい。

採用しない。

### agent contract を JSON Schema としてだけ定義する

agent output の形式をすべて JSON Schema で定義し、adapter interface や prompt / CLI 契約は文書化しない案。

`plan.json` / `review-N.json` については schema が正本である。しかし agent adapter には、process 起動、timeout、stdout envelope、stderr、raw output 保存、permission、working directory など schema だけでは表現しにくい責務がある。

schema だけでは runner と adapter の境界が曖昧になるため、採用しない。

### agent raw output をそのまま workflow 判断に使う

agent の自然文や raw output を runner が読み取り、blocked / findings / pass を推測する案。

この案は実装が柔軟に見えるが、Kogoto の状態遷移を不安定にする。既存仕様では raw output は debugging artifact であり、workflow 判断には normalized JSON を使う方針になっている。

採用しない。

### Write / Fix にも normalized JSON artifact を作る

`implement.json` や `fix-N.json` のようなファイルを追加し、writer の作業完了結果を永続 JSON として保存する案。

将来、詳細な監査が必要になった場合は有力である。しかし v0 では `Write` / `Fix` の正本は worktree の diff / git status / verification result であり、artifact を増やすと schema と recovery policy が増える。

v0 では採用しない。

## 判断理由

agent adapter contract を独立した内部境界として定義すると、次の整合性が保てる。

- runner は状態遷移と workflow 制御に集中できる
- backend ごとの差分を adapter 内に閉じ込められる
- raw output と normalized artifact の既存方針を維持できる
- `schemas/` が normalized JSON の正本であることを保てる
- `codex` を known-but-unsupported として扱う既存方針と矛盾しない
- Claude Code CLI の仕様変更があっても、主に adapter と agent contract の更新で対応できる

## 結果

- `docs/internal/agent-contract.md` を agent adapter 境界の正本とする
- `docs/internal/log-spec.md` から agent contract へ参照を追加する
- runner は adapter result を受け取り、artifact 保存、schema validation、state transition、verification、git 操作を行う
- adapter は runner state を直接書き換えない
- v0 では Claude Code adapter を実装対象とし、Codex adapter は将来対応とする

## 関連

- Issue #10: create documents
- [docs/internal/agent-contract.md](../internal/agent-contract.md)
- [docs/internal/state-machine.md](../internal/state-machine.md)
- [docs/internal/error-handling.md](../internal/error-handling.md)
- [docs/internal/log-spec.md](../internal/log-spec.md)
- [schemas/README.md](../../schemas/README.md)
