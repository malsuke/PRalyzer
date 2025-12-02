package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v77/github"
)

/**
 * /search/issueを使ってコメントにkeywordが含まれるPRを検索する
 * PR番号のスライスを返す（API呼び出しを削減するため、完全なPRオブジェクトは取得しない）
 */
func (c *Client) SearchPullRequestsWithCommentKeyword(keyword string) ([]int, error) {
	ctx := context.Background()

	query := fmt.Sprintf("repo:%s/%s in:comments type:pr %s", c.Owner, c.Name, keyword)

	var allPRNumbers []int
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

			if issue.Number == nil {
				continue
			}

			allPRNumbers = append(allPRNumbers, *issue.Number)
		}

		if resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
	}

	return allPRNumbers, nil
}
