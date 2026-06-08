# ユーザーリポジトリに特定のADR形式を強制しない

Status: Accepted
DateTime: 2026-06-08T18:31:00+09:00

## 背景

Kogoto のワークフローは、ADR candidate を選定したあとユーザーのリポジトリに ADR を保存する。この保存処理を「save strategy」として扱う。

save strategy を設計するにあたり、次の問いが生じた。

- Kogoto は特定の ADR フォーマット・保存先・命名規則をユーザーに強制すべきか
- Kogoto が提供する「デフォルト」の save strategy はどのような性格を持つべきか
- デフォルト strategy の名称として「kogoto_standard」を使うべきか

ユーザーリポジトリには、すでに独自の ADR 運用（形式・ディレクトリ構成・命名規則）が存在している場合がある。Kogoto の利用を始めるために既存の運用を変更することを求めるべきかどうかが、判断の核心となった。

## 決定

ユーザーリポジトリに特定の ADR 形式・保存先・命名規則を強制しない。

具体的には、次の通りとする。

- Kogoto が提供する save strategy は、あくまで「オプションとして選択できるデフォルト」であり、使用を強制しない
- ユーザーは自分のリポジトリの ADR 運用に合わせて strategy を選択または設定できる
- デフォルト strategy の名称は `default` とする。`kogoto_standard` という名称は使わない

## 検討した代替案

### 特定のADR形式を強制する（`kogoto_standard` 戦略）

Kogoto が提供する形式・ディレクトリ・命名規則に統一することを前提とし、それを `kogoto_standard` という名称で提供する案。

この案では、Kogoto の ADR 運用が統一されるため、Kogoto 自身がリポジトリをスキャン・解析しやすくなる。

一方で、次の問題がある。

- 既存の ADR 運用を持つプロジェクトが Kogoto を導入する際に摩擦が生じる
- `kogoto_standard` という名称は、Kogoto がユーザーリポジトリの ADR 形式を「規格化」しているかのような印象を与える
- Kogoto の役割は ADR 形式を定めることではなく、ADR を記録する過程を支援することである

採用しない。

### デフォルト戦略を提供しない

完全にユーザーが独自の strategy を定義することを前提とし、Kogoto はデフォルト strategy を持たない案。

新規ユーザーが Kogoto を始める際に、何も設定せずに使い始められないという問題がある。採用しない。

## 判断理由

Kogoto の役割は、ADR をどのフォーマットで書くべきかを規定することではなく、ADR を作成・記録する過程を human-in-the-loop で支援することである。

ユーザーリポジトリの ADR 運用はプロジェクトごとに異なる。Kogoto の導入コストを下げるためには、既存の運用を尊重する設計が必要である。

`kogoto_standard` という名称は、Kogoto が ADR の標準仕様であるかのような誤解を招く。`default` という名称は、「Kogoto が提供するデフォルトの選択肢」という意味に留まり、ユーザーへの強制を示唆しない。

## 結果

- Kogoto は ADR のフォーマット・保存先・命名規則をユーザーリポジトリに強制しない
- 設定なしに使い始められるよう、`default` strategy をデフォルトとして提供する
- `default` strategy は、オプションとして選択できる指針であり、規格ではない
- `kogoto_standard` という名称は Kogoto のコードベース・ドキュメント内で使用しない

## 採用しない方針

- ユーザーリポジトリの ADR 保存先・フォーマット・命名規則を Kogoto が規定する
- `kogoto_standard` を strategy 名として使用する
- Kogoto の利用開始に際して特定の ADR ディレクトリ構成を要求する

## 関連

- Issue #43: Kogoto による ADR 運用の具体化
- Issue #51: ユーザー ADR の candidate extraction / selection / save strategy の詳細
- [docs/internal/adr-types.md](../internal/adr-types.md) — ユーザー ADR と内部 ADR の区別
