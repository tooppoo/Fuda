# Architecture Decision Records（内部 ADR）

このディレクトリは、Kogoto 自身の設計判断を **内部 ADR** として管理するための場所である。

> Kogoto の開発では ADR を「ユーザー ADR」と「内部 ADR」の 2 種類に区別する。  
> このディレクトリが対象とするのは **内部 ADR** のみである。  
> 2 種類の区別については [docs/internal/adr-types.md](../internal/adr-types.md) を参照。

ADR は、Kogoto の現在仕様そのものではなく、「なぜその判断をしたか」を残すための記録である。現在有効な仕様・利用方法・実装方針は、必要に応じて `docs/` 配下の通常文書にも反映する。

---

## 目的

ADR は、後続の Issue・実装・レビュー・設計判断に影響する重要な意思決定を記録する。

ADR に残す内容は、単なる結論ではない。少なくとも、判断時点の背景、検討した代替案、採用理由、結果として発生する制約・影響を記録する。

これにより、後から同じ論点を検討するときに、過去の判断を再利用・検査・更新できるようにする。

---

## ADR と docs の役割分担

| 種類 | 役割 |
|---|---|
| Issue / Issue comments | 検討過程、質問、回答、暫定判断を記録する |
| ADR | 後続に影響する重要判断を、理由・代替案・影響とともに記録する |
| docs | ADR によって安定した現在仕様・利用方法・運用方針を記述する |

ADR は「なぜそうしたか」を記録する。

docs は「現在どう扱うか」を記述する。

---

## 保存場所

Kogoto の内部 ADR は、このディレクトリに保存する。

```text
docs/adr/
```

ADR は Markdown ファイルとして保存する。

---

## ファイル命名規則

ADR ファイル名は、timestamp prefix と kebab-case title を組み合わせる。

```text
YYYYMMDDTHHMMSSZ-kebab-case-title.md
```

例:

```text
20260608T081500Z-use-go-as-implementation-language.md
20260608T083000Z-keep-distribution-method-undecided.md
```

### timestamp prefix

- timestamp は UTC を使う
- 形式は `YYYYMMDDTHHMMSSZ` とする
- 秒まで含める
- `Z` は UTC を表す
- timestamp は原則として ADR 作成時刻を表す

### title 部分

- lowercase の kebab-case とする
- 英数字と hyphen を使う
- 日本語タイトルを本文見出しに使う場合でも、ファイル名の title 部分は英語の短い識別子にする
- ファイル拡張子は `.md` とする

### 採番を使わない理由

Kogoto では ADR ファイル名の prefix として連番ではなく timestamp を使う。

理由は次の通りである。

- 複数の ADR が並行して作成されても衝突しにくい
- Git 上での作成順を把握しやすい
- branch / PR をまたいだ作業で採番競合が起きにくい
- ADR の安定参照をファイル名に集約できる

---

## ADR 本文の基本構成

ADR は、原則として次の構成を持つ。

```markdown
# ADR: <title>

Status: Proposed | Accepted | Deprecated | Superseded  
DateTime: YYYY-MM-DDTHH:mm:ss+09:00

## 背景

## 決定

## 検討した代替案

## 判断理由

## 結果

## 関連
```

必要に応じて、節を追加してよい。ただし、背景・決定・代替案・判断理由・結果の区別は保つ。

現時点では、Kogoto 用 template をこの README に記載する。独立した template file は作らない。

将来的に Kogoto default template として利用者向けにも提供する場合は、ファイル分離を別途検討する。

---

## DateTime

ADR 本文には `Date` ではなく `DateTime` を記載する。

```markdown
DateTime: 2026-06-08T17:15:00+09:00
```

要件:

- ISO 8601 形式で記載する
- timezone offset を必ず含める
- 日本時間で記録する場合は `+09:00` を付ける
- UTC で記録する場合は `Z` を使ってよい

ファイル名の timestamp prefix は UTC compact form とし、本文の `DateTime` は timezone 付きの human-readable な形式とする。

---

## Status

ADR の `Status` は、次のいずれかとする。

| Status | 意味 |
|---|---|
| `Proposed` | 提案中。まだ確定判断ではない |
| `Accepted` | 採用済み。現在有効な判断である |
| `Deprecated` | 非推奨。過去判断として残すが、新規には使わない |
| `Superseded` | 後続 ADR によって置き換えられた |

`Superseded` にする場合は、置き換え先の ADR を `関連` に明記する。

---

## ADR を作成する基準

すべての判断を ADR にする必要はない。

ADR を作成するべき判断:

- 複数の実質的な選択肢がある
- 後続 Issue・実装・レビューに影響する
- 一度決めると戻しにくい
- 複数のファイル・機能・workflow に波及する
- 将来同じ判断を繰り返す可能性がある
- agent が今後参照すべき設計判断である
- なぜその判断になったかを残さないと、後から誤解されやすい

ADR を作成しなくてもよい判断:

- typo 修正
- 局所的な refactor
- 既存方針に従っただけの実装
- 一時的な作業メモ
- 実験的な spike の途中経過
- Issue comments で完結する軽微な判断

---

## ADR の書き方

ADR では、次の区別を明確にする。

- 事実として確認できること
- 判断時点の前提
- 検討した代替案
- 採用した決定
- 決定の理由
- 決定によって発生する制約・影響
- 未解決の論点

ADR は結論を正当化するための文章ではなく、後から検査・更新できる判断記録である。

そのため、採用しなかった案も、単に否定するのではなく、どの条件では有力で、なぜ今回は採用しなかったのかを記録する。

---

## ADR の更新

原則として、採用済み ADR の本文は、誤字修正やリンク修正を除いて大きく書き換えない。

判断が変わった場合は、次のいずれかで扱う。

- 既存 ADR を `Deprecated` にする
- 既存 ADR を `Superseded` にし、新しい ADR を作成する
- 現在仕様だけが変わる場合は、ADR ではなく docs を更新する

過去の判断を消すのではなく、判断がどのように変わったかを追跡できる状態に保つ。

---

## 関連 Issue / PR の扱い

ADR には、関連する Issue / PR / docs を `関連` 節に記載する。

例:

```markdown
## 関連

- Issue #50: Go製CLIアプリケーションとしての配布方式を検討する
- PR #xx: docs: add ADR management README
- docs/concept/human-in-the-loop-first.md
```

ADR は Issue comments の代替ではない。Issue comments で検討過程を残し、ADR では確定した重要判断を要約して保存する。

---

## 例

```markdown
# ADR: Kogoto の実装言語として Go を採用する

Status: Accepted  
DateTime: 2026-06-08T17:15:00+09:00

## 背景

Kogoto は local-first / CLI-first な開発支援 runner として設計する。

## 決定

Kogoto の実装言語として Go を採用する。

## 検討した代替案

- TypeScript / Node.js
- Rust
- Python
- Bun / Deno

## 判断理由

Go は軽量な native binary CLI を作りやすく、local filesystem、process execution、Git operation を扱いやすい。

## 結果

Kogoto は Go 製 CLI application として実装する。

## 関連

- Issue #50
```
