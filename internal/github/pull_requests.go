package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v77/github"
)

/**
 * /search/issueを使ってコメントにkeywordが含まれるPRを検索する
 */
func (c *Client) SearchPullRequestsWithCommentKeyword(keyword string) ([]*github.PullRequest, error) {
	ctx := context.Background()

	query := fmt.Sprintf("repo:%s/%s in:comments type:pr %s", c.Owner, c.Name, keyword)

	var allPRs []*github.PullRequest
	page := 1
	perPage := 100

	for {
		opts := &github.SearchOptions{
			ListOptions: github.ListOptions{
				Page:    page,
				PerPage: perPage,
			},
		}

		result, resp, err := c.github.Search.Issues(ctx, query, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to search issues: %w", err)
		}

		for _, issue := range result.Issues {
			if issue.PullRequestLinks == nil {
				continue
			}

			pr, _, err := c.github.PullRequests.Get(ctx, c.Owner, c.Name, *issue.Number)
			if err != nil {
				return nil, fmt.Errorf("failed to get pull request #%d: %w", *issue.Number, err)
			}
			allPRs = append(allPRs, pr)
		}

		if resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
	}

	return allPRs, nil
}
