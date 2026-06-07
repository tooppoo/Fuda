# Fuda から Kogoto への移行

## 概要

v0 は正式リリース前であり、Kogoto は旧 Fuda の config / state / binary alias を自動移行しない。

既存の Fuda 設定・状態ファイルがある場合は、必要に応じて手動で内容を確認し、新しい Kogoto 設定として作り直す。

## ファイルパスの変更

| 種別 | 旧パス（Fuda） | 新パス（Kogoto） |
|------|--------------|----------------|
| config ファイル | `~/.config/fuda/config.toml` | `~/.config/kogoto/config.toml` |
| run state ディレクトリ | `.fuda/` | `.kogoto/` |
| state ログ | `~/.local/state/fuda/` | `~/.local/state/kogoto/` |

## binary

`fuda` binary alias（`kogoto` へのリダイレクト）は v0 では提供しない。

`kogoto` コマンドを直接使用する。

## 移行手順

既存の `~/.config/fuda/config.toml` がある場合:

1. 内容を確認する
2. `~/.config/kogoto/config.toml` として新たに作成する（`kogoto setup` を使うか、手動でコピー・編集する）
3. 旧 `~/.config/fuda/` は自動削除されないため、不要であれば手動で削除する

既存の `.fuda/` run state がある場合:

1. `.fuda/` の状態は `.kogoto/` へ自動移行されない
2. 進行中の run がある場合は、旧 Fuda で完了・abort させるか、手動で状態を確認した上で破棄する

## 注意

正式リリース（v1 以降）でユーザーが存在する時点で移行が必要になった場合は、その時点で別途移行ガイドを設計する。

## 関連

- [コマンドリファレンス](./commands.md)
- [後方互換性方針](../concept/naming.md#後方互換性方針)
