# 状態データ設計

## 基本方針

```text
Issue = 管理単位
Run   = Issue処理の1回の実行試行・実行記録
```

Kogoto の基本操作は Issue 単位で行われる。

```bash
kogoto resolve <issue-number>
kogoto status <issue-number>
kogoto answer <issue-number>
kogoto resume <issue-number>
kogoto close <issue-number>
```

そのため、Kogoto の永続状態も Issue を管理単位とする。

ただし、Run の概念は捨てない。Run は、特定の Issue に対する1回の実行試行・実行記録として保持する。

---

## 状態データの責務分担

```text
issue-state.json = Issue全体の現在状態・current run選択の正本
run.json         = 個別Runの内部状態・再開判断の正本
```

### `issue-state.json` の責務

- Issue 全体の現在状態（`issue_workflow_state`）を保持する
- `current_run_id` を保持する
- failed / aborted / retried を含む run 履歴を保持する

Issue 単位の操作（`status`、`resume`、`answer`、`close`）は、まず `issue-state.json` を参照して現在の `current_run_id` を特定する。

### `run.json` の責務

- 個別 Run の内部状態（`run_state`）を保持する
- planning / writing / testing / committing / reviewing / fixing など、1回の Run 内部の状態遷移を扱う
- 特定 Run 内部の blocked state・human review required state を保持する
- Run レベルの失敗情報・再開判断の正本とする

`run.json` は GitHub source issue の完了状態を表さない。`run_state = succeeded` は、Kogoto Run が正常終了したことを意味する。GitHub source issue が close されたことは意味しない。Issue レベルの完了状態は `issue_workflow_state` で表す。

---

## 状態の二層化

```text
issue_workflow_state = Issue全体の現在状態
run_state            = 個別Runの内部状態
```

### `issue_workflow_state` の候補値

| 値 | 意味 |
|---|---|
| `not_started` | まだ Run が開始されていない |
| `active` | Run が実行中である |
| `waiting_for_human` | 人間の回答・判断を待っている |
| `pr_created` | PR が作成済みである |
| `completed` | Issue 処理が完了した（GitHub Issue close 済み） |
| `aborted` | Issue 処理が中断された |
| `failed` | Run が失敗し、Issue 処理が止まっている |

### `run_state` について

`run_state` の定義は [state-machine.md](state-machine.md) を参照。

`issue_workflow_state` と `run_state` は混同してはならない。

- `issue_workflow_state` は Issue 全体のワークフロー状態を表す
- `run_state` は個別 Run の内部状態遷移を表す

---

## failed / aborted / retried run の履歴保持

failed・aborted・retried の run は削除せず、`issue-state.json` の `runs` 配列に履歴として保持する。

```json
{
  "runs": [
    { "run_id": "<run-id-1>", "run_result": "retried" },
    { "run_id": "<run-id-2>", "run_result": "aborted" },
    { "run_id": "<run-id-3>", "run_result": "active" }
  ],
  "current_run_id": "<run-id-3>"
}
```

`current_run_id` は、現在アクティブな Run または最後に実行された Run の `run_id` を指す。

---

## `issue-state.json` の例

```json
{
  "schema_version": 1,
  "repository": "tooppoo/Kogoto",
  "issue_number": 43,
  "issue_workflow_state": "waiting_for_human",
  "current_run_id": "<run-id>",
  "runs": [
    {
      "run_id": "<run-id>",
      "run_result": "active"
    }
  ],
  "source_issue": {
    "url": "https://github.com/tooppoo/Kogoto/issues/43",
    "last_seen_comment_id": 4646615488,
    "updated_at": "2026-06-08T00:00:00Z"
  },
  "created_at": "2026-06-08T00:00:00Z",
  "updated_at": "2026-06-08T00:00:00Z"
}
```

---

## 状態データのパス

### 最終的なディレクトリ構造

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

例:

```text
.kogoto/
  repositories/
    github.com/
      tooppoo/
        Kogoto/
          issues/
            51/
              issue-state.json
              runs/
                <run-id>/
                  run.json
```

### ディレクトリの責務

| ディレクトリ | 責務 |
|---|---|
| `.kogoto/__global__/` | 特定の repository / Issue に属さない workspace-level state を置く。Issue 作成前の draft workflow はここに置く |
| `.kogoto/repositories/<host>/<owner>/<repo>/` | repository に属する状態データを置く |
| `issues/<issue-number>/issue-state.json` | Issue 全体の現在状態・current run・run 履歴を保持する |
| `issues/<issue-number>/runs/<run-id>/run.json` | 個別 Run の内部状態・再開判断を保持する |

### repository 識別子に host / owner / repo を含める理由

repository name だけでは repository を一意に識別できない。

例えば、`tooppoo/Kogoto` と `someone/Kogoto` は repository name が同じでも別物である。

そのため、repository の状態データは `<host>/<owner>/<repo>` を key とする。これにより、GitHub Enterprise や `github.com` 以外の host にも将来対応しやすくなる。

### ディレクトリ名に複数形を使う

集合を表すディレクトリには複数形を使う。

使用する: `repositories/`, `issues/`, `runs/`, `drafts/`

使用しない: `repository/`, `issue/`, `run/`, `draft/`

---

## Issue draft workflow

Issue 作成支援 workflow は `resolve` とは別の workflow として扱う。

Issue 作成前には GitHub Issue number が存在しない。そのため、draft state は `issues/` 配下には保存しない。

draft state は次に保存する。

```text
.kogoto/__global__/drafts/<draft-id>/
  issue-draft.json
  discussion.md
```

GitHub Issue 作成後、その draft と次のパスを接続できる。

```text
.kogoto/repositories/<host>/<owner>/<repo>/issues/<issue-number>/
```

---

## 各コマンドが参照する状態データ

| コマンド | 参照する正本 | 参照する目的 |
|---|---|---|
| `kogoto resolve <issue>` | `issue-state.json` | Issue の現在状態確認、新 Run 開始 |
| `kogoto status <issue>` | `issue-state.json` | Issue 全体の状態表示 |
| `kogoto status <issue>` | `run.json`（`current_run_id` 経由） | 実行中 Run の詳細表示 |
| `kogoto resume <issue>` | `issue-state.json` | `current_run_id` の特定 |
| `kogoto resume <issue>` | `run.json`（`current_run_id` 経由） | Run の再開判断 |
| `kogoto answer <issue>` | `issue-state.json` | `current_run_id` の特定 |
| `kogoto answer <issue>` | `run.json`（`current_run_id` 経由） | blocked 状態の確認・更新 |
| `kogoto close <issue>` | `issue-state.json` | Issue 状態を `completed` に更新 |
