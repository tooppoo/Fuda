# review-N.json — Normalization Rules and Recovery Policy

Schema definition: [review.schema.json](review.schema.json)

## Purpose

`review-N.json` は raw reviewer output の正規化が成功した場合にのみ書き込まれる normalized review output である。

raw reviewer output は `review-N.raw.txt` として常に保存する。

## File numbering

`review-N.json` の `review_number` は `N` と一致する。

```
review-1.json → review_number = 1
review-2.json → review_number = 2
```

## Write condition

`review-N.json` は reviewer output の parse・schema validation・normalization がすべて成功した場合にのみ書き込む。

```
normalization 成功 → review-N.json を書く
normalization 失敗 → review-N.json を書かない。review-N.raw.txt のみ保存。
                     run_state = "failed" + last_error.code = "invalid_reviewer_output" に遷移。
```

`review-N.raw.txt` が存在し `review-N.json` が存在しない場合、normalization に失敗した可能性を示す。

## `reviewer_assessment` vs `runner_decision`

`reviewer_assessment` は reviewer が自己申告した評価であり、runner の状態遷移の正本ではない。

`runner_decision` は `findings` と `human_review_required` から runner が導出した制御判断であり、こちらが review 後の状態遷移の正本となる。

v0 では `reviewer_assessment` に `inconclusive` を採用しない。reviewer が判断不能・出力不完全・構造化に失敗した場合は、`review-N.json` を書かずに `run_state = "failed"` + `last_error.code = "invalid_reviewer_output"` に遷移する。

## Runner decision derivation

`review-N.json` が書かれる = normalization 成功が確定している。

`runner_decision` は `findings` と `human_review_required` から次の優先順で導出する。

| 条件 | `runner_decision` |
|---|---|
| `human_review_required` が非空 | `human_review_required` |
| `findings` に `blocking` または `major` がある | `needs_revision` |
| `findings` が `minor` のみ | `ready_with_minor_findings` |
| `findings` が空 | `pass` |

normalization 失敗時は `runner_decision` をこのファイルに書かず、runner が `run_state = "failed"` + `last_error.code = "invalid_reviewer_output"` に遷移する。

Example — `reviewer_assessment` と `runner_decision` が乖離するケース:

```
reviewer_assessment = "needs_fix" かつ findings は minor のみ
→ runner_decision = ready_with_minor_findings

reviewer_assessment = "pass" かつ findings に major が含まれる
→ runner_decision = needs_revision
```

## Semantic rules

- `review-N.json` は normalization 成功時にのみ書く。normalization 失敗時は書かない
- `review_number` は対応するファイル名の `N` と一致しなければならない
- `findings[].file` および `findings[].line` は optional であり、該当する場合のみ持つ
- `related_finding_ids` は関連 finding がない場合は空配列とする

## Recovery policy

### Normalization 失敗（write 前）

```
reviewer output の parse / schema validation / normalization に失敗
→ review-N.json を書かない
→ review-N.raw.txt を保存（すでに存在する場合は上書きしない）
→ run_state = "failed" + last_error.code = "invalid_reviewer_output" に遷移
→ require re-review or human confirmation
```

### 書き込み済み review-N.json の corruption（write 後）

一度正常に書かれた `review-N.json` が後から破損した場合、自動的に `pass` 扱いしてはならない。

```
review-N.json の読み込みに失敗
→ stop
→ retire as review-N.json.corrupt.<timestamp>; preserve raw output
→ require re-review or human confirmation
```
