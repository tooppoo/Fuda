package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/tooppoo/Kogoto/internal/tracker"
)

type Client struct {
	owner  string
	repo   string
	token  string
	host   string
	client *http.Client
}

func New(owner, repo, token, host string) *Client {
	return &Client{
		owner:  owner,
		repo:   repo,
		token:  token,
		host:   host,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

type apiIssue struct {
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	HTMLURL   string    `json:"html_url"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (c *Client) GetIssue(ctx context.Context, number int) (tracker.Issue, error) {
	url := fmt.Sprintf("https://api.%s/repos/%s/%s/issues/%d", c.host, c.owner, c.repo, number)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return tracker.Issue{}, err
	}
	req.Header.Set("Authorization", "token "+c.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := c.client.Do(req)
	if err != nil {
		return tracker.Issue{}, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusNotFound:
		return tracker.Issue{}, fmt.Errorf("issue #%d not found", number)
	case http.StatusUnauthorized:
		return tracker.Issue{}, fmt.Errorf("GitHub authentication failed")
	default:
		return tracker.Issue{}, fmt.Errorf("GitHub API error: %s", resp.Status)
	}

	var issue apiIssue
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return tracker.Issue{}, fmt.Errorf("decode issue response: %w", err)
	}

	return tracker.Issue{
		Number:    issue.Number,
		Title:     issue.Title,
		Body:      issue.Body,
		URL:       issue.HTMLURL,
		UpdatedAt: issue.UpdatedAt,
	}, nil
}

func (c *Client) PostComment(ctx context.Context, number int, body string) (int64, error) {
	url := fmt.Sprintf("https://api.%s/repos/%s/%s/issues/%d/comments", c.host, c.owner, c.repo, number)

	payload, err := json.Marshal(map[string]string{"body": body})
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "token "+c.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return 0, fmt.Errorf("post comment failed: %s", resp.Status)
	}

	var result struct {
		ID int64 `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("decode comment response: %w", err)
	}

	return result.ID, nil
}
