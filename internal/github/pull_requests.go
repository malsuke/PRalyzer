package github

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-github/v77/github"
)

/**
 * /search/issueを使ってコメントにkeywordが含まれるPRを検索する
 * PR番号のスライスを返す（API呼び出しを削減するため、完全なPRオブジェクトは取得しない）
 */
func (c *Client) SearchPullRequestsWithCommentKeyword(keyword string) ([]int, error) {
	ctx := context.Background()

	query := fmt.Sprintf("repo:%s/%s in:comments type:pr is:merged %s", c.Owner, c.Name, keyword)

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

/**
 * リポジトリ内のすべてのPull Requestを取得する
 * レート制限エラーが発生した場合、1時間5分待機してからリトライする
 */
func (c *Client) ListAllPullRequests() ([]*github.PullRequest, error) {
	ctx := context.Background()

	var allPRs []*github.PullRequest
	page := 1
	perPage := 100

	fmt.Printf("Starting to fetch pull requests...\n")

	for {
		var prs []*github.PullRequest
		var resp *github.Response
		var err error

		// レート制限エラーが発生するまでリトライ
		for {
			opts := &github.PullRequestListOptions{
				State: "all", // open, closed, all
				ListOptions: github.ListOptions{
					Page:    page,
					PerPage: perPage,
				},
			}

			fmt.Printf("Fetching page %d (per page: %d)...\n", page, perPage)
			prs, resp, err = c.github.PullRequests.List(ctx, c.Owner, c.Name, opts)
			if err != nil {
				if isRateLimitError(err) {
					// レート制限エラーの場合、1時間5分待機してからリトライ
					waitDuration := 65 * time.Minute // 1時間5分
					fmt.Printf("\n⚠️  Rate limit exceeded. Waiting %v before retrying page %d...\n", waitDuration, page)
					waitForRateLimit(waitDuration)
					fmt.Printf("Retrying page %d...\n", page)
					continue // リトライ
				}
				return nil, fmt.Errorf("failed to list pull requests: %w", err)
			}
			break // 成功したらループを抜ける
		}

		allPRs = append(allPRs, prs...)
		fmt.Printf("  ✓ Fetched %d PRs from page %d (total: %d PRs)\n", len(prs), page, len(allPRs))

		if resp.NextPage == 0 {
			fmt.Printf("Reached last page. Total PRs fetched: %d\n", len(allPRs))
			break
		}
		page = resp.NextPage
	}

	return allPRs, nil
}

/**
 * エラーがレート制限エラーかどうかを判定する
 */
func isRateLimitError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// HTTPレスポンスエラーの場合
	if strings.Contains(errStr, "403") || strings.Contains(errStr, "429") {
		return true
	}

	// go-githubのレート制限エラーメッセージをチェック
	if strings.Contains(strings.ToLower(errStr), "rate limit") {
		return true
	}

	// RateLimitError型をチェック
	if rateLimitErr, ok := err.(*github.RateLimitError); ok {
		return rateLimitErr != nil
	}

	// AbuseRateLimitError型をチェック
	if abuseRateLimitErr, ok := err.(*github.AbuseRateLimitError); ok {
		return abuseRateLimitErr != nil
	}

	return false
}

/**
 * 指定時間待機する（レート制限リセット待ち）
 */
func waitForRateLimit(waitDuration time.Duration) {
	// 待機中は定期的に進捗を表示
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	done := make(chan bool)
	go func() {
		time.Sleep(waitDuration)
		done <- true
	}()

	elapsed := time.Duration(0)
	for {
		select {
		case <-done:
			fmt.Printf("Rate limit wait completed. Resuming processing...\n")
			return
		case <-ticker.C:
			elapsed += 10 * time.Minute
			remaining := waitDuration - elapsed
			if remaining > 0 {
				fmt.Printf("Still waiting... %v remaining\n", remaining)
			}
		}
	}
}

/**
 * 指定されたPR番号のPull Requestの詳細を取得する
 */
func (c *Client) GetPullRequest(prNumber int) (*github.PullRequest, error) {
	ctx := context.Background()

	pr, _, err := c.github.PullRequests.Get(ctx, c.Owner, c.Name, prNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get pull request #%d: %w", prNumber, err)
	}

	return pr, nil
}
