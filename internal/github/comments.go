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
