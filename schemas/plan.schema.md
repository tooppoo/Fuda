# plan.json — Semantic Rules and Recovery Policy

Schema definition: [plan.schema.json](plan.schema.json)

## Purpose

`plan.json` は writer が Issue をどう解釈し、作業可能かどうかを返した正規化済み出力である。raw writer output ではない。

## Semantic rules

- `planning_result = "ready_to_write"` の場合、`questions` は空配列でよい
- `planning_result = "blocked_by_ambiguity"` の場合、`questions` は `minItems: 1`
- blocking question がある場合、runner は Issue に質問コメントを投稿し、`run_state = "blocked"` に遷移する

## Recovery policy

`plan.json` は再生成可能に見えるが、過去の判断根拠でもあるため勝手に上書きしない。

```
plan.json invalid
→ stop
→ retire corrupt file as plan.json.corrupt.<timestamp>
→ allow explicit re-plan only by human command
```
