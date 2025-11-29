package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	gh "github.com/google/go-github/v77/github"
	"github.com/malsuke/PRalyzer/internal/github"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: PRalyzer <repository-url> [github-pat]\nNote: GitHub PAT is optional but recommended to avoid rate limiting")
	}

	repoURL := os.Args[1]
	var githubPAT string
	if len(os.Args) >= 3 {
		githubPAT = os.Args[2]
	} else {
		fmt.Println("Warning: No GitHub PAT provided. Rate limiting may occur.")
	}

	client, err := github.NewClient(githubPAT, repoURL, nil)
	if err != nil {
		log.Fatalf("Failed to create GitHub client: %v", err)
	}

	words, err := loadWordList("word_list.json")
	if err != nil {
		log.Fatalf("Failed to load word list: %v", err)
	}

	prNumbersMap := make(map[int]bool)

	for _, word := range words {
		fmt.Printf("Searching PRs with keyword: %s\n", word)

		prs, err := client.SearchPullRequestsWithCommentKeyword(word)
		if err != nil {
			if isRateLimitError(err) {
				log.Printf("Rate limit exceeded while searching PRs with keyword '%s'. Stopping.", word)
				log.Printf("Tip: Provide a GitHub PAT to increase rate limits.")
				break
			}
			log.Printf("Failed to search PRs with keyword '%s': %v", word, err)
			continue
		}

		for _, pr := range prs {
			if pr.Number != nil {
				prNumbersMap[*pr.Number] = true
			}
		}
	}

	// 4. 配列の中身の重複を削除（マップを使っているので既に重複は削除済み）
	prNumbers := make([]int, 0, len(prNumbersMap))
	for prNumber := range prNumbersMap {
		prNumbers = append(prNumbers, prNumber)
	}

	fmt.Printf("Found %d unique PRs\n", len(prNumbers))

	// PRが見つからなかった場合はディレクトリを作成せずに終了
	if len(prNumbers) == 0 {
		fmt.Println("No PRs found. Exiting without creating directory.")
		return
	}

	// データディレクトリの作成
	dataDir := filepath.Join("data", client.Owner, client.Name)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	// 5. issues/<PRの番号>/commentsエンドポイントを使ってコメントを取得
	// 6. <PRの番号>.jsonでPR内のコメントをファイルに書き込む
	for _, prNumber := range prNumbers {
		fmt.Printf("Fetching comments for PR #%d\n", prNumber)

		comments, err := client.GetComments(prNumber)
		if err != nil {
			if isRateLimitError(err) {
				log.Printf("Rate limit exceeded while fetching comments for PR #%d. Stopping.", prNumber)
				log.Printf("Tip: Provide a GitHub PAT to increase rate limits.")
				log.Printf("Progress: Successfully processed %d PRs before rate limit.", prNumber-1)
				break
			}
			log.Printf("Failed to get comments for PR #%d: %v", prNumber, err)
			continue
		}

		// JSONファイルに書き込む
		outputPath := filepath.Join(dataDir, fmt.Sprintf("%d.json", prNumber))
		if err := writeCommentsToFile(comments, outputPath); err != nil {
			log.Printf("Failed to write comments to file for PR #%d: %v", prNumber, err)
			continue
		}

		fmt.Printf("Saved comments for PR #%d to %s\n", prNumber, outputPath)
	}

	fmt.Println("Done!")
}

// loadWordList はword_list.jsonファイルを読み込む
func loadWordList(filename string) ([]string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read word list file: %w", err)
	}

	var words []string
	if err := json.Unmarshal(data, &words); err != nil {
		return nil, fmt.Errorf("failed to parse word list JSON: %w", err)
	}

	return words, nil
}

// writeCommentsToFile はコメントをJSONファイルに書き込む
func writeCommentsToFile(comments []*gh.IssueComment, filepath string) error {
	data, err := json.MarshalIndent(comments, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal comments: %w", err)
	}

	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// isRateLimitError はエラーがレート制限エラーかどうかを判定する
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
	if rateLimitErr, ok := err.(*gh.RateLimitError); ok {
		return rateLimitErr != nil
	}

	// AbuseRateLimitError型をチェック
	if abuseRateLimitErr, ok := err.(*gh.AbuseRateLimitError); ok {
		return abuseRateLimitErr != nil
	}

	return false
}

// waitForRateLimitReset はレート制限がリセットされるまで待機する（オプション機能）
func waitForRateLimitReset(err error) {
	if rateLimitErr, ok := err.(*gh.RateLimitError); ok {
		resetTime := rateLimitErr.Rate.Reset.Time
		waitDuration := time.Until(resetTime)
		if waitDuration > 0 {
			log.Printf("Waiting %v for rate limit reset...", waitDuration)
			time.Sleep(waitDuration)
		}
	}
}
