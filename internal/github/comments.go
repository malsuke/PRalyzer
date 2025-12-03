package github

import (
	"context"

	"github.com/google/go-github/v77/github"
)

/**
 * issues/<prNumber>/commentsエンドポイントを使ってコメントを取得する
 */
func (c *Client) GetComments(prNumber int) ([]*github.IssueComment, error) {
	comments, _, err := c.github.Issues.ListComments(context.Background(), c.Owner, c.Name, prNumber, &github.IssueListCommentsOptions{})
	if err != nil {
		return nil, err
	}
	return comments, nil
}

/**
 * pulls/<prNumber>/commentsエンドポイントを使ってレビューコメントを取得する
 */
func (c *Client) GetReviewComments(prNumber int) ([]*github.PullRequestComment, error) {
	comments, _, err := c.github.PullRequests.ListComments(context.Background(), c.Owner, c.Name, prNumber, &github.PullRequestListCommentsOptions{})
	if err != nil {
		return nil, err
	}
	return comments, nil
}
