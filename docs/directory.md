# ディレクトリ構成

## 概要

```
cmd/
  kogoto/
    main.go          # CLIエントリポイント

internal/            # ソースコード
  config/            # 設定の読み込み・管理
  tracker/           # Issue Tracker の抽象層（インタフェース・ドメイン型）
    github/          # tracker.Tracker の GitHub 実装
  git/               # Git 操作（worktree の作成・切り替えなど）
  agent/             # エージェントアダプタ（writer / reviewer の起動と通信）
  runner/            # Kogoto 実行ライフサイクルの管理
  state/             # 実行状態の永続化・読み込み
  shell/             # 外部プロセス実行のユーティリティ

docs/                # ドキュメント
  internal/          # Kogoto開発用の内部文書. 設計・用語定義など
  usage/             # Kogotoユーザー向けの外部文書. 利用方法など
```

## 各パッケージの責務

### `cmd/kogoto`

CLIエントリポイント。cobra でサブコマンドを登録し、`internal/runner` に処理を委譲する。
ビジネスロジックはここに書かない。

### `internal/`

#### `config`

設定ファイルの読み込みと構造体へのマッピング。
GitHub トークンやエージェントの設定など、起動時に決まるパラメータを扱う。

#### `tracker`

Issue Tracker に関するインタフェースとドメイン型を定義する抽象層。
`Issue`, `Comment`, `PullRequest` などの型と `Tracker` インタフェースをここに置く。
`runner` はこの層にのみ依存し、特定の Issue Tracker 実装を直接参照しない。

#### `tracker/github`

`tracker.Tracker` の GitHub 実装。GitHub REST API（または `gh` CLI）を使って
Issue の取得・コメント投稿・PR 作成などを行う。
将来 GitLab や Linear などに対応する場合、同じ階層に別の実装パッケージを追加する。

#### `git`

Git コマンドのラッパー。worktree の作成・削除、ブランチ操作など。
エージェントの作業を分離した worktree で行うための操作を提供する。

### `agent`

writer エージェントと reviewer エージェントの起動・通信を抽象化するアダプタ層。
特定のエージェント実装（Claude Code など）に依存せず、ロールベースのインタフェースで表現する。

### `runner`

Kogoto 実行（1 Issue に対する 1 Run）のライフサイクルを管理する。
setup → write → review → (revise) → resolve の各フェーズを状態遷移として扱う。

### `state`

実行状態をローカルに永続化する。中断・再開（`resume`）や `status` 表示のために使用する。
ファイルベースのシンプルなストレージを想定している。

### `shell`

外部プロセスの起動・出力取得などの共通ユーティリティ。
`agent` や `git` から利用する。

## 方針

- `internal/` 配下のパッケージはアルファベット順ではなく、依存の方向で考える。
  `cmd/kogoto` → `runner` → `tracker`, `git`, `agent`, `state` → `tracker/github`, `shell`, `config`
- 上位層（`runner`）は抽象（`tracker`）にのみ依存し、実装（`tracker/github`）を直接参照しない。
- 具体的な実装への依存は `cmd/kogoto` でインジェクションする。
- パッケージはスキャフォールドに必要な段階で作成する。空ディレクトリはコミットしない。

### `docs/`

### `adr/` 

Kogotoの[内部ADR](./internal/adr-types.md]を配置する。
`docs/adr` のドキュメントは、リポジトリ内部にある任意のファイルを参照して良い。

### `concept/`

Kogotoのコンセプトについて説明するドキュメントを配置する。
`docs/concept` のドキュメントは、 以下に列挙したパッケージのファイルのみ参照して良い。

- `docs/usage`
- `docs/concept`

#### `internal/`

Kogoto開発用の内部ドキュメント。
`docs/internal/` のドキュメントは、 以下に列挙したパッケージのファイルのみ参照して良い。

- `internal/`
- `docs/internal`

### `usage/`

Kogotoユーザー向けの内部ドキュメント。
`docs/usage/` のドキュメントは、 以下に列挙したパッケージのファイルのみ参照して良い。

- `docs/internal`
- `docs/usage`
- `docs/concept`
