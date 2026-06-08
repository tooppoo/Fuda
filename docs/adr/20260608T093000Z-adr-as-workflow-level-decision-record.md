# ADRをワークフローレベルの意思決定記録として扱う

Status: Accepted
DateTime: 2026-06-08T18:30:00+09:00

## 背景

Issue #40 において、Kogoto の human-in-the-loop ワークフローに ADR を知識蓄積の一部として組み込む方針が定められた。

その後、ADR が Kogoto のワークフローの中でどのような位置付けにあるかを具体化する必要が生じた。特に、次の問いを整理する必要があった。

- ADR は `run.json` スキーマや agent protocol と同列のものか
- ADR の記録対象をどの粒度で定義するか
- ADR candidate の選定を誰が行うか（Kogoto が自動判断するか、人間が判断するか）

Kogoto の知識の流れは次のように整理されている。

```
判断過程: Issue comments
判断記録: ADR
現在仕様: docs
```

ADR はこの流れの中間に位置し、Issue comments の審議過程から確定した重要判断を抽出して記録するものである。

## 決定

ADR を、Kogoto のワークフローレベルの意思決定記録として扱う。

具体的には、次の通りとする。

- ADR は `run.json` スキーマ・agent protocol・run state とは独立した、ワークフローの一成果物として位置付ける
- ADR candidate の選定は人間が行う。Kogoto は候補を提示するが、記録するかどうかは人間が判断する
- Kogoto のワークフローは、ADR に関して次の操作を行う
  - **ADR candidate extraction** — Issue・コメント・diff から ADR 候補を抽出する
  - **ADR candidate selection** — 抽出した候補のうち記録に値するものを人間と協調して選定する
  - **ADR save** — 選定した ADR をリポジトリに保存する

## 検討した代替案

### ADR 作成を自動化する

Kogoto が ADR 候補を検出したとき、人間の確認なしに自動的に ADR を作成・保存する案。

この案は、Kogoto のコアコンセプトである「強い HITL（Human-in-the-loop first）」に反する。ADR は「記録するに値する判断」を選定することに意義がある。自動化すると、人間が内容を評価する機会が失われ、知識蓄積の質が下がる。

採用しない。

### ADR を run state の一部として扱う

ADR の保存を run state（`run.json`）のフィールドとして管理する案。

ADR はワークフローの成果物であり、run の実行状態とは異なる関心事である。run state に含めると、run state スキーマの変更が ADR 処理の変更に波及しやすくなる。関心事を分離する観点から採用しない。

### ADR を Kogoto のワークフローに含めない

ADR の記録を Kogoto の workflow の外側に置き、開発者が手動で管理する案。

Kogoto が知識蓄積を支援しないとすれば、Issue comments は検討過程として残るが、重要判断を後続の Issue・実装・agent が参照しやすい形で記録する手段がなくなる。Kogoto の human-in-the-loop ワークフローの価値の一部が失われるため、採用しない。

## 判断理由

ADR をワークフローレベルの成果物と位置付けることで、次の整合性が保たれる。

- run state スキーマ・agent protocol の変更が ADR 処理に波及しない
- ADR candidate selection の判断を人間に委ねることで、強い HITL の原則と一致する
- Issue comments（審議過程）→ ADR（判断記録）→ docs（現在仕様）という知識フローが明確になる

## 結果

- Kogoto のワークフローは ADR candidate extraction・selection・save を明示的に含む
- ADR candidate の選定は常に人間が行う
- `run.json` スキーマおよび agent protocol の変更は、ADR 処理の変更を伴わない
- ADR の位置付けが明確になったことで、ユーザー ADR と内部 ADR の区別も自然に成立する（[docs/internal/adr-types.md](../internal/adr-types.md) 参照）

## 関連

- Issue #43: Kogoto による ADR 運用の具体化
- Issue #40: ADR を human-in-the-loop ワークフローに組み込む（概念的方針）
- [docs/concept/human-in-the-loop-first.md](../concept/human-in-the-loop-first.md) — 強い HITL の定義と知識蓄積方針
- [docs/internal/adr-types.md](../internal/adr-types.md) — ユーザー ADR と内部 ADR の区別
