# ADR ワークフロー

Kogoto は ADR（Architecture Decision Record）を開発 workflow 上の判断記録レイヤーとして扱う。

ADR の概念的な位置付けは [Human-in-the-loop first](../concept/human-in-the-loop-first.md) を参照。ここでは、Kogoto の workflow における ADR の具体的な扱い・方針を説明する。

---

## ADR の役割

Kogoto の workflow では、知識の流れを次のように整理する。

```
判断過程: Issue comments
判断記録: ADR
現在仕様: docs
```

ADR は、後続 Issue・実装・設計・agent の判断に影響する重要決定を記録するレイヤーである。

---

## Kogoto は ADR を自動確定しない

Kogoto が行うのは、Issue comments / PR discussion / run summary などから設計判断候補を抽出し、ユーザーに提示するところまでである。

ADR として記録する / 記録しない / 後で判断する / 編集して保存する、という判断はユーザーに残す。

---

## ADR の形式・保存先・命名規則

ADR の形式・保存先・命名規則はユーザーの repository 側で管理する。Kogoto が特定の形式・保存先・命名規則を強制することはない。

どのような保存処理を行うかは、ADR save strategy によって選択できる。

---

## ADR save strategy

Kogoto は次の ADR save strategy を概念上の候補として定義する。

| strategy | 説明 |
|---|---|
| `disabled` | ADR 支援を行わない |
| `manual` | ADR candidates を表示するだけで、ファイル保存は行わない |
| `default` | Kogoto が提供する標準形式を使って保存する。ただし、利用者 repository への強制ではない |
| `external_command` | ユーザー定義コマンドに ADR candidate を渡す（将来案） |

`default` は Kogoto が提供する標準形式であるが、利用者 repository の構造・命名規則・テンプレートを上書きするものではない。

v0 では CLI 実装を行わず、workflow と方針の文書化に留める。CLI 実装は別 Issue とする。

---

## ADR candidate 抽出

Kogoto は、Issue comments / PR discussion / run summary などから設計判断候補を抽出する。

---

## ADR candidate selection

ADR candidate selection では、複数候補をまとめて選択できるようにする想定とする。

選択 UI では ADR title だけを一覧表示する。候補本文・判断理由・詳細内容の閲覧や編集は、Kogoto の選択 UI の範囲外とする。ユーザーが任意の方法で確認・編集する。

Kogoto の責務は次に限定する。

- ADR candidate の title 一覧を提示する
- ユーザーが記録対象候補を複数選択できるようにする
- 選択結果を保存・出力処理へ渡す

---

## Evidence

Evidence は Issue URL / PR URL が記載されていればよい。

Issue comment URL / PR review comment URL / excerpt 単位の厳密な evidence tracking は要求しない。

---

## 将来の標準導線

将来的に、`kogoto close <issue-number>` 実行時に ADR candidate review を行うことを標準導線として想定する。

```bash
# 将来: kogoto close 時に ADR candidate review を含む想定
kogoto close 7
```

また、独立したコマンドとして次のようなものも検討対象とする。

```bash
# 将来案
kogoto adr review <issue-number>
```

v0 では、これらの CLI 実装は行わない。

---

## スコープ外（別 Issue）

| 項目 | Issue |
|---|---|
| ADR candidate の状態データ保持方式 | [Issue #51](https://github.com/tooppoo/Kogoto/issues/51) |
| Kogoto 全体のキャッシュ機構の設計 | [Issue #52](https://github.com/tooppoo/Kogoto/issues/52) |
| `external_command` strategy の実装 | 別途検討 |
| `kogoto adr review` CLI の実装 | 別途検討 |
| `kogoto close` 時の ADR candidate review CLI | 別途検討 |

---

## 関連

- [Human-in-the-loop first](../concept/human-in-the-loop-first.md)
- [ADR の種類](../internal/adr-types.md)
- [内部 ADR 管理ルール](../adr/README.md)
- [Issue #51](https://github.com/tooppoo/Kogoto/issues/51) — ユーザー ADR candidate extraction / selection / save strategy の実装
- [Issue #52](https://github.com/tooppoo/Kogoto/issues/52) — Kogoto 全体のキャッシュ機構
