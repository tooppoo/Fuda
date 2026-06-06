# Run のライフサイクル

`fuda resolve <issue>` を実行すると、Fuda は以下の順序で Issue を処理する。

## 通常フロー（happy path）

```mermaid
stateDiagram-v2
    [*] --> loading_issue : fuda resolve &lt;issue&gt;

    loading_issue --> preparing_worktree : Issue 取得完了
    preparing_worktree --> planning : worktree 作成完了
    planning --> writing : 計画完了
    writing --> testing : 変更完了
    testing --> committing : テスト通過
    committing --> reviewing : commit 作成
    reviewing --> pr_created : レビュー通過・PR 作成
    pr_created --> succeeded : 終了処理完了

    succeeded --> [*]
```

各フェーズの概要:

| フェーズ | 内容 |
|---|---|
| `loading_issue` | GitHub Issue の本文・コメントを取得する |
| `preparing_worktree` | Issue 専用の worktree と branch を作成する |
| `planning` | writer agent が実装計画を作成する |
| `writing` | writer agent が変更を実装する |
| `testing` | test / lint / typecheck を実行して検証する |
| `committing` | 変更を commit する |
| `reviewing` | reviewer agent が差分・テスト結果・受け入れ基準を検査する |
| `pr_created` | PR 作成済み。`run.json` に PR number / URL を記録した状態 |
| `succeeded` | Run が正常終了し、終了処理が完了した状態 |

## 補足

通常フローには含まれないが、次のケースが発生することがある。

- **修正ループ**: reviewer が `blocking` / `major` 指摘を出した場合、`reviewing → fixing → testing` を繰り返す
- **blocked**: writer が不明点を検出した場合、Fuda は Issue にコメントして人間の回答を待つ
- **human_review_required**: 修正ループ上限到達または自動判断困難な場合、人間の判断を求めて停止する

各シナリオの詳細は [scenarios.md](scenarios.md) を参照。

完全な状態遷移（異常系・再開ポリシーを含む）は [内部仕様: Run State Machine](../internal/state-machine.md) を参照。
