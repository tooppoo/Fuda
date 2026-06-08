# ADR CLI支援をv0から先送りにする

Status: Accepted
DateTime: 2026-06-08T18:32:00+09:00

## 背景

Issue #43 では、Kogoto における ADR 運用の具体化として、次の4項目を検討対象として挙げていた。

1. ファイル命名規則
2. テンプレート
3. 保存ディレクトリ
4. **作成支援 CLI**

作成支援 CLI とは、ADR candidate を対話的に選定・作成するための CLI 機能（複数候補の表示・選択、ADR ファイルの自動生成など）を指す。

v0 の実装スコープを決定するにあたり、この CLI 支援を v0 に含めるかどうかを判断する必要があった。

## 決定

ADR 作成支援 CLI の実装を v0 から先送りにする。

v0 では、ADR に関して次の範囲のみを対象とする。

- ADR のワークフロー上の位置付けの定義（ドキュメント）
- save strategy の設計方針の定義（ドキュメント）
- 内部 ADR の管理ルール（ドキュメント）

v0 における ADR 対応は、ドキュメントとワークフロー定義に限定する。CLI の実装は行わない。

CLI 実装は Issue #54 で扱う。

## 検討した代替案

### ADR 作成支援 CLI を v0 に含める

v0 の段階で CLI による ADR candidate 選定・ファイル生成機能を実装する案。

機能としての価値はある。しかし、v0 は Kogoto の core workflow（Issue scope → writer agent → reviewer agent → run state → human checkpoint）の実装を優先する段階である。

ADR 作成支援 CLI は便利な機能だが、core workflow が動作しない段階では実際に使われることがなく、仕様も固まっていない。v0 で実装することは、scope の拡大と仕様の早期固定につながるリスクがある。

採用しない。

## 判断理由

v0 の目的は、Kogoto の core workflow を動かし、human-in-the-loop の基本的な流れを検証することである。

ADR 作成支援 CLI は、workflow が一定程度動作してから実際の使用パターンが見えてくる機能である。v0 の段階で先行実装すると、実際の使用パターンに基づいた設計修正が困難になる可能性がある。

まずドキュメントと workflow 定義によって ADR の役割と save strategy の方針を固め、CLI 実装は core workflow の動作後に着手することが適切である。

## 結果

- v0 では ADR 作成支援 CLI を実装しない
- v0 における ADR 対応は、ワークフロー定義とドキュメントに限定する
- CLI 実装は v0 完了後に Issue #54 で着手する
- v0 段階では、ADR の作成は開発者が手動で行う

## 関連

- Issue #43: Kogoto による ADR 運用の具体化（本 ADR の直接の起点）
- Issue #54: ADR 作成支援 CLI の実装（先送り先）
