# Kogoto と ADR

ADR（Architecture Decision Record）は、開発上の重要な判断を記録するドキュメントである。

Kogoto は ADR を、人間と agent の相互作用から生まれる判断を蓄積するレイヤーとして位置づける。

---

## Kogoto における ADR の位置づけ

Kogoto の workflow では、知識の流れを次のように整理する。

```
判断過程: Issue comments
判断記録: ADR
現在仕様: docs
```

Issue を処理する過程で、人間と agent のやりとりから設計上の判断が生まれる。その判断を後続の Issue・実装・設計に活かすために、ADR として記録する。

ADR を使うことで、次のことができるようになる。

- 過去の判断を再利用・検査・更新できる
- 将来の agent が過去の判断を参照できる
- なぜその設計になったかを、後から追跡できる

---

## Kogoto は ADR を自動確定しない

Kogoto が行うのは、Issue comments / PR discussion / run summary などから設計判断の候補を抽出し、ユーザーに提示するところまでである。

- 「この候補を ADR として記録するかどうか」はユーザーが判断する
- 記録する / しない / 後で判断する / 編集してから保存するといった判断は、Kogoto が代わりに行わない

---

## ADR の形式・保存先はユーザーが管理する

ADR の形式・保存先・命名規則は、ユーザーの repository 側で管理する。Kogoto がこれらを強制することはない。

Kogoto が ADR をどう扱うかは、ADR save strategy によって設定できる。

| strategy | 説明 |
|---|---|
| `disabled` | ADR 支援を行わない |
| `manual` | ADR 候補を表示するだけで、ファイル保存は行わない |
| `default` | Kogoto が提供する標準形式を使って保存する |
| `external_command` | ユーザー定義コマンドに ADR 候補を渡す（将来案） |

`default` を選んでも、ユーザーの repository の既存の構造・命名規則・テンプレートが上書きされるわけではない。

---

## 将来の利用イメージ

将来的に、`kogoto close` 実行時に ADR candidate review が含まれる想定である。

```bash
kogoto close 7
# → PR merge を確認して cleanup を実行
# → ADR 候補を提示し、記録するかどうかをユーザーが選択
```

v0 では ADR に関する CLI 操作は実装されていない。

---

## 関連

- [Human-in-the-loop first](../concept/human-in-the-loop-first.md) — ADR を知識蓄積の一部として位置付ける Kogoto のコアコンセプト
- [ADR の種類](../internal/adr-types.md) — ユーザー ADR と内部 ADR の区別
- [ADR ワークフロー（内部設計）](../internal/adr.md) — save strategy・candidate selection・evidence 要件の詳細
