# ADR: Kogotoの状態データをIssue単位とRun単位で管理する

Status: Accepted
DateTime: 2026-06-08T21:00:00+09:00

## 背景

Kogoto は Issue 駆動の workflow tool である。主要なユーザー向け操作は GitHub Issue を対象に行われる。

```bash
kogoto resolve <issue-number>
kogoto status <issue-number>
kogoto answer <issue-number>
kogoto resume <issue-number>
kogoto close <issue-number>
```

従来の状態関連ドキュメントでは、`run.json` が Run の再開・停止・検査のための状態ファイルとして中心に置かれていた。

しかし、Kogoto の操作モデルは Run 中心ではない。Run は、ある Issue に対する1回の実行試行にすぎない。このため、次の不一致があった。

- ユーザーは Issue を対象に操作する
- Kogoto 内部では実行試行を Run として追跡する
- blocking question・人間レビュー・close 処理は Issue 文脈に属する
- retry / failed / aborted の実行履歴は個別 Run に属する

したがって、Kogoto には二層の状態モデルが必要である。

## 決定

Kogoto の状態データを次の二層構成で管理する。

```text
Issue = 管理単位
Run   = Issue処理の1回の実行試行・実行記録
```

状態の責務を次のように分離する。

```text
issue-state.json = Issue全体の現在状態・current run選択の正本
run.json         = 個別Runの内部状態・再開判断の正本
```

状態データのパスは、global workspace と multiple repositories に対応できる構成とする。

```text
.kogoto/
  __global__/
    workspace-state.json
    drafts/
      <draft-id>/
        issue-draft.json
        discussion.md

  repositories/
    <host>/
      <owner>/
        <repo>/
          issues/
            <issue-number>/
              issue-state.json
              runs/
                <run-id>/
                  run.json
                  plan.json
                  review-1.json
                  review-1.raw.txt
                  run-summary.json
```

## 検討した代替案

### Run を最上位の管理単位とする

`run.json` のみで状態管理を完結させる案。Issue との対応は `run.json.issue_number` で管理する。

採用しなかった理由:

- ユーザーの操作モデル（Issue 単位）と内部管理単位（Run）が一致しない
- 複数の Run（failed / retried）をまたいだ Issue レベルの状態管理が難しくなる
- `kogoto status <issue>` / `kogoto resume <issue>` の実装で、どの Run が current かを判断するロジックが複雑になる

### 単一ファイルですべての状態を管理する

`issue-run-state.json` のような単一ファイルで Issue と Run の両方の状態を保持する案。

採用しなかった理由:

- Run 履歴が蓄積するにつれてファイルが肥大化する
- Issue 単位の操作と Run 単位の操作で更新タイミングが異なるため、競合・整合性管理が複雑になる
- Run のアーカイブ・削除の粒度が粗くなる

### `.kogoto/issues/<issue-number>/` をルートとする（単一 repository 前提）

当初案として `.kogoto/issues/<issue-number>/...` のパス構成が検討された。

採用しなかった理由:

- 複数 repository や global workspace の状態を管理できない
- `tooppoo/Kogoto` と `someone/Kogoto` のような同名 repository が衝突する
- 将来 GitHub Enterprise や他の host に対応する場合に拡張できない

## 判断理由

### Issueを主要な管理単位にする理由

ユーザーの操作単位と Kogoto 内部の管理単位を揃えることで、`kogoto status`・`kogoto resume` のような Issue 単位のコマンドが自然に実装できる。

### Runの概念を維持する理由

1回の Issue 処理が複数の実行試行（failed → retry → succeeded）を経る場合、各試行の履歴を追跡できる必要がある。Run を個別ファイルとして保持することで、過去の実行記録を失わずに再試行できる。

### `<host>/<owner>/<repo>` を repository key にする理由

repository name だけでは同名の別 repository が衝突する。host まで含めることで一意性を保証する。

### `__global__/` を分離する理由

Issue 作成前の draft workflow は GitHub Issue number を持たない。Issue 文脈に属さない状態を `__global__/` に分離することで、repository-scoped な状態と明確に区別できる。

## 結果

### 正の影響

- Kogoto の内部状態管理がユーザーの操作モデルと一致する
- Run 履歴（failed / aborted / retried）を失わずに保持できる
- パス設計が multiple repositories と global workspace に最初から対応できる

### トレードオフ

- 単一の `run.json` だけで管理する設計よりも状態モデルが複雑になる
- `issue-state.json.current_run_id` と対応する `runs/<run-id>/run.json` の整合性を維持する必要がある

### リスク

- `issue_workflow_state` と `run_state` の違いが実装で混同される可能性がある。特に `run_state = succeeded` を Issue 完了として誤って扱わないよう注意する
- 旧案のパス（`.kogoto/issues/<issue-number>/...`）がドキュメントや実装に残ると、誤った構造で状態データが作られる可能性がある

## 関連

- Issue #51: 状態データ保持を Issue 単位 + Run 単位の二層構成として文書化する
- Issue #43
- docs/internal/state-data.md
- docs/internal/state-machine.md
- schemas/run.schema.md
