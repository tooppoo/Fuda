package tracker

import (
	"context"
	"time"
)

type Issue struct {
	Number    int
	Title     string
	Body      string
	URL       string
	UpdatedAt time.Time
}

type Tracker interface {
	GetIssue(ctx context.Context, number int) (Issue, error)
	PostComment(ctx context.Context, number int, body string) (int64, error)
}
