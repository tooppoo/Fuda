package github

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	gogithub "github.com/google/go-github/v73/github"

	"github.com/tooppoo/Kogoto/internal/tracker"
)

type Client struct {
	owner  string
	repo   string
	client *gogithub.Client
}

func New(owner, repo, token, host string) (*Client, error) {
	var ghClient *gogithub.Client
	if host == "github.com" {
		ghClient = gogithub.NewClient(nil).WithAuthToken(token)
	} else {
		var err error
		ghClient, err = gogithub.NewEnterpriseClient(
			fmt.Sprintf("https://%s/api/v3/", host),
			fmt.Sprintf("https://%s/api/uploads/", host),
			nil,
		)
		if err != nil {
			return nil, fmt.Errorf("create GitHub enterprise client for %s: %w", host, err)
		}
		ghClient = ghClient.WithAuthToken(token)
	}
	return &Client{owner: owner, repo: repo, client: ghClient}, nil
}

func (c *Client) GetIssue(ctx context.Context, number int) (_ tracker.Issue, err error) {
	issue, resp, err := c.client.Issues.Get(ctx, c.owner, c.repo, number)
	if err != nil {
		var gerr *gogithub.ErrorResponse
		if errors.As(err, &gerr) {
			switch gerr.Response.StatusCode {
			case http.StatusNotFound:
				return tracker.Issue{}, fmt.Errorf("issue #%d not found", number)
			case http.StatusUnauthorized:
				return tracker.Issue{}, fmt.Errorf("GitHub authentication failed")
			}
		}
		return tracker.Issue{}, err
	}
	defer func() {
		err = errors.Join(err, resp.Body.Close())
	}()

	if issue.IsPullRequest() {
		return tracker.Issue{}, fmt.Errorf("#%d is a pull request, not an issue", number)
	}

	return tracker.Issue{
		Number:    issue.GetNumber(),
		Title:     issue.GetTitle(),
		Body:      issue.GetBody(),
		URL:       issue.GetHTMLURL(),
		UpdatedAt: issue.GetUpdatedAt().Time,
	}, nil
}

func (c *Client) PostComment(ctx context.Context, number int, body string) (_ int64, err error) {
	comment, resp, err := c.client.Issues.CreateComment(ctx, c.owner, c.repo, number, &gogithub.IssueComment{
		Body: gogithub.Ptr(body),
	})
	if err != nil {
		return 0, err
	}
	defer func() {
		err = errors.Join(err, resp.Body.Close())
	}()

	return comment.GetID(), nil
}
