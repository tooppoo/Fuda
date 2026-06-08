# Kogoto の実装言語として Go を採用する

Status: Accepted
DateTime: 2026-06-08T17:25:00+09:00

## 背景

Kogoto は、Issue driven development のための Human-in-the-loop-first な CLI application である。

Kogoto は、自律的に backlog を処理し続ける agent platform ではなく、GitHub Issue を作業単位として扱い、人間の判断、writer agent、reviewer agent、local workspace、Git operation、run state を明示的に接続する開発支援 runner として設計する。

Kogoto の実装言語は、次の要件に影響する。

* CLI としての起動速度
* local filesystem 操作
* Git repository / worktree 操作
* 外部 agent process の実行・制御
* run state の読み書き
* JSON / Markdown / 設定ファイルの扱い
* cross-platform binary の作成
* dependency 管理
* 個人開発・小規模チームでの保守性
* 将来的な配布方式の選択肢

Kogoto は、local-first / CLI-first / lightweight な tool として育てる。そのため、実装言語には、軽量な CLI application を作りやすく、追加 runtime への依存を小さくし、Linux / macOS / Windows 向けの配布物を作りやすいことが求められる。

## 決定

Kogoto の実装言語として Go を採用する。

Kogoto は、Go 製 CLI application として実装する。実行本体は native binary として扱うことを基本方針とする。

現時点では、配布方式そのものはこの ADR では確定しない。npm、GitHub Releases、Homebrew、Scoop、Winget、`go install`、install script などの具体的な配布方式は、別 Issue で検討し、必要に応じて別 ADR として記録する。

## 検討した代替案

### TypeScript / Node.js

TypeScript / Node.js は、CLI 実装、JSON 処理、GitHub API 連携、AI tooling との接続に強い。npm 配布とも自然に接続できる。

しかし、Kogoto では実行本体を Node script として配布することを前提にしない。Kogoto は、追加 runtime への依存を小さくした native binary CLI として扱えることが望ましい。

Node.js を実行時基盤にすると、次の懸念がある。

* 利用環境に Node.js runtime を要求する
* dependency tree が大きくなりやすい
* package manager 差異の影響を受けやすい
* local-first CLI としての単純な配布からやや離れる

したがって、TypeScript / Node.js は採用しない。

### Rust

Rust は、native binary、cross-platform distribution、型安全性、性能、堅牢性の点で有力な候補である。

一方で、Kogoto の現時点の主要課題は、低レベル性能やメモリ安全性ではなく、human checkpoint、Issue scope、agent handoff、run state、review loop、Git operation の設計である。

Rust を採用すると、実装の厳密性は高まるが、MVP段階での実装速度と設計変更への追従性が下がる可能性がある。Kogoto の初期段階では、複雑な所有権設計よりも、明示的で読みやすく、変更しやすい CLI implementation を優先する。

したがって、Rust は採用しない。

### Python

Python は、agent orchestration、プロトタイピング、外部API連携には適している。

しかし、Kogoto は開発者がローカル環境で直接使う CLI application である。Python を実行時基盤にすると、Python version、virtual environment、dependency installation、OS差異への対応が配布上の論点になりやすい。

Kogoto では、利用者が追加の runtime / environment 管理を意識せずに実行できる配布物を目指すため、Python は採用しない。

### Bun / Deno による single executable

Bun / Deno は、TypeScript で実装しつつ executable を生成できる選択肢である。

ただし、これは基本的には JavaScript / TypeScript runtime を同梱した executable であり、Go / Rust のような native CLI tool とは性質が異なる。TypeScript で実装できる利点はあるが、Kogoto の中核は frontend や web runtime ではなく、local workspace、process execution、Git operation、state management である。

そのため、Bun / Deno は採用しない。

## 判断理由

Go は、Kogoto の制約に対して最も均衡がよい。

Go を採用する理由は次の通りである。

* 小さな CLI application を作りやすい
* 起動が速い
* native binary として配布しやすい
* cross-platform build と相性がよい
* 標準ライブラリで filesystem / process / JSON / HTTP / concurrency を扱いやすい
* dependency を抑えた実装にしやすい
* error handling が明示的で、停止条件や失敗理由を表現しやすい
* 個人開発・小規模チームで読みやすく保守しやすい
* Git / local workspace / external process を扱う CLI runner という性質に合っている

Kogoto は、hosted service ではなく local-first CLI として設計する。Go は、この方針に対して過剰な runtime dependency や framework dependency を持ち込みにくい。

また、Kogoto は agent framework そのものではなく、人間の判断と agent execution の間にある handoff / checkpoint / review / state を制御する runner である。Go の明示的で単純な制御構造は、この用途に適している。

## 結果

### 良い影響

* Kogoto を軽量な CLI application として実装しやすい
* native binary を配布しやすい
* npm 以外の配布方式にも展開しやすい
* local filesystem / process execution / Git operation を扱いやすい
* dependency surface を抑えやすい
* run state や stop reason を明示的に扱いやすい
* 初期MVPを比較的速く実装しやすい

### 悪い影響・制約

* AI tooling 周辺の一部ecosystemは TypeScript / Python の方が豊富である
* Rust と比べると、型による状態遷移の厳密な表現力は弱い
* rich TUI / GUI を重視する場合は、別途 library selection が必要になる
* error handling が冗長になりやすい
* package layout を誤ると、過剰分割または巨大 package に寄りやすい

## 採用しない方針

この ADR では、次の方針は採用しない。

* Kogoto を Node script CLI として実装する
* Kogoto を Python script CLI として実装する
* Kogoto を hosted service 前提で設計する
* 実装初期から汎用 agent framework として過度に一般化する
* 外部の巨大 project layout を機械的に採用する
* 配布方式をこのADRで確定する

## 配布方式について

配布方式は、このADRでは確定しない。

現時点では、次の候補を検討対象とする。

* GitHub Releases による platform 別 native binary archive
* npm package による installation
* `go install` による Go 開発者向け fallback
* Homebrew
* Scoop
* Winget
* deb / rpm / apk
* install script
* Docker / OCI image

配布方式は、別 Issue で整理し、必要に応じて別 ADR として記録する。

## 関連

* Issue #50: Go製CLIアプリケーションとしての配布方式を検討する
* `go.mod`
* `cmd/kogoto/main.go`
