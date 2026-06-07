# Run のライフサイクル

`kogoto resolve <issue>` を実行すると、Kogoto は以下の順序で Issue を処理する。

## 通常フロー（happy path）

```mermaid
flowchart TD
    START([kogoto resolve &lt;issue&gt;]) --> A[Issue 取得]
    A --> B[Worktree 準備]
    B --> C[計画]
    C --> D[実装]
    D --> E[検証]
    E --> F[Commit]
    F --> G[レビュー]

    G -->|問題なし| H([PR 作成・完了])
    G -->|要修正| FIX[修正]
    FIX --> E
    G -->|自動判断困難| STOP([停止・人間確認])

    C -->|不明点あり| WAIT([停止・回答待ち])
    D -->|不明点あり| WAIT
    FIX -->|不明点あり| WAIT
    WAIT -->|kogoto resume| C
```

各フェーズの概要:

| フェーズ | 内容 |
|---|---|
| Issue 取得 | GitHub Issue の本文・コメントを取得する |
| Worktree 準備 | Issue 専用の worktree と branch を作成する |
| 計画 | writer agent が実装計画を作成する |
| 実装 | writer agent が変更を実装する |
| 検証 | test / lint / typecheck を実行して検証する |
| Commit | 変更を commit する |
| レビュー | reviewer agent が差分・テスト結果・受け入れ基準を検査する |
| 修正 | reviewer の指摘を受けて writer agent が修正する |
| PR 作成・完了 | PR を作成し、Run が正常終了する |
| 停止・回答待ち | writer が不明点を検出した場合に停止する。`kogoto resume` で再開できる |
| 停止・人間確認 | 自動判断が困難な場合に停止する。人間の判断を求める |

## 主な分岐

- **修正ループ**: reviewer が修正必要と判断した場合、修正 → 検証 を繰り返す（上限あり）
- **回答待ち**: writer が不明点を検出した場合、Kogoto は Issue に質問を投稿して停止する。`kogoto resume` で再開できる
- **人間確認**: 修正ループ上限到達または自動判断が困難な場合、Kogoto は停止して人間の判断を求める

各シナリオの詳細は [scenarios.md](scenarios.md) を参照。

完全な状態遷移図（`run_state` enum・異常系・再開ポリシーを含む）は [内部仕様: Run State Machine](../internal/state-machine.md) を参照。
