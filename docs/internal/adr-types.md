# ADR の種類

Kogoto の開発では、ADR を次の 2 種類に区別する。

---

## ユーザー ADR

Kogoto の workflow が取り扱う対象の ADR である。

Kogoto が Issue を処理する過程で、次の操作を行う対象となる。

- **ADR candidate extraction** — Issue・コメント・diff から ADR 候補を抽出する
- **ADR candidate selection** — 抽出した候補のうち、記録に値するものを人間と協調して選定する
- **ADR save strategy** — 選定した ADR をユーザーのリポジトリに保存する

ユーザー ADR は、Kogoto を使って開発しているプロジェクト（ユーザーのリポジトリ）の設計判断記録である。

保存先・フォーマット・命名規則はユーザーのプロジェクトに依存する。Kogoto は save strategy に従って処理するが、保存先は Kogoto 自身のリポジトリではない。

> ユーザー ADR の candidate extraction / selection / save strategy の詳細は [Issue #51](https://github.com/tooppoo/Kogoto/issues/51) で扱う。

---

## 内部 ADR

Kogoto 自身の設計判断を記録する ADR である。

Kogoto の開発者が、Kogoto の設計・実装・方針に関する重要な判断を内部文書として残すためのものである。

内部 ADR は Kogoto リポジトリの `docs/adr/` に保存する。

保存ルール・命名規則・テンプレートは [docs/adr/README.md](../adr/README.md) を参照。

---

## 対比

| 項目 | ユーザー ADR | 内部 ADR |
|---|---|---|
| 対象 | ユーザーのプロジェクトの設計判断 | Kogoto 自身の設計判断 |
| 作成者 | Kogoto workflow（人間と協調） | Kogoto 開発者 |
| 保存先 | ユーザーのリポジトリ（save strategy による） | `docs/adr/` |
| Kogoto との関係 | Kogoto が処理・操作する対象 | Kogoto 開発の内部文書 |

---

## 関連

- [docs/adr/README.md](../adr/README.md) — 内部 ADR の管理ルール
- [Issue #51](https://github.com/tooppoo/Kogoto/issues/51) — ユーザー ADR の扱い（candidate extraction / selection / save strategy）
- [Human-in-the-loop first](../concept/human-in-the-loop-first.md) — Kogoto のコアコンセプト（ADR を知識蓄積の一部として位置付ける）
