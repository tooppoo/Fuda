package schemas

import _ "embed"

//go:embed run.schema.json
var RunSchemaJSON []byte

//go:embed issue-state.schema.json
var IssueStateSchemaJSON []byte
