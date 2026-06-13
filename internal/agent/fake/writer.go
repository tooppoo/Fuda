package fake

import (
	"context"

	"github.com/tooppoo/Kogoto/internal/agent"
)

// Writer is a deterministic fake that always returns blocked_by_ambiguity.
// It is used in place of the real Claude adapter for the blocked flow.
type Writer struct{}

func (w *Writer) Plan(_ context.Context, _ agent.PlanInput) (agent.PlanResult, error) {
	return agent.PlanResult{
		Plan: agent.Plan{
			PlanningResult: agent.PlanningResultBlockedByAmbiguity,
			Summary:        "Cannot proceed: clarification required before implementation.",
			Questions: []agent.Question{
				{
					ID:       "q1",
					Question: "Please clarify the acceptance criteria and expected behavior for this issue.",
					Blocking: true,
				},
			},
		},
	}, nil
}
